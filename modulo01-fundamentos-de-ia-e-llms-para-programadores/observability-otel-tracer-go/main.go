// Demonstração de instrumentação OpenTelemetry em Go — porte do projeto Java
// equivalente (observability-otel-tracer-java). Simula um fluxo de checkout
// de e-commerce com spans aninhados (checkout → payment + inventory-check),
// atributos de span e uma métrica de contador (orders.placed), exportando
// tudo para um OTLP Collector via gRPC — mesmo comportamento observável da
// versão Java, usando o SDK oficial go.opentelemetry.io/otel.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// catálogo simulado de produtos.
var products = []string{
	"prod-001-sneakers",
	"prod-002-headphones",
	"prod-003-laptop-bag",
	"prod-004-mechanical-keyboard",
	"prod-005-usb-hub",
}

var users = []string{"user-42", "user-17", "user-88", "user-3", "user-99"}

func main() {
	loadDotEnv(".env")

	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	serviceName := getEnv("SERVICE_NAME", "go-ecommerce")

	log.Println("=============================================================")
	log.Println(" OpenTelemetry Go E-Commerce Demo")
	log.Println("=============================================================")
	log.Println(" Service name :", serviceName)
	log.Println(" OTLP endpoint:", otlpEndpoint)
	log.Println("=============================================================")

	ctx := context.Background()

	tracerProvider, meterProvider, shutdown, err := buildSDK(ctx, otlpEndpoint, serviceName)
	if err != nil {
		log.Fatalf("falha ao inicializar OpenTelemetry SDK: %v", err)
	}
	defer shutdown(ctx)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)

	tracer := tracerProvider.Tracer("com.observability.ecommerce", trace.WithInstrumentationVersion("1.0.0"))
	meter := meterProvider.Meter("com.observability.ecommerce")

	ordersPlacedCounter, err := meter.Int64Counter(
		"orders.placed",
		metric.WithDescription("Total number of orders successfully placed"),
		metric.WithUnit("{order}"),
	)
	if err != nil {
		log.Fatalf("falha ao criar contador orders.placed: %v", err)
	}

	const totalOrders = 5
	log.Printf("Starting simulation – %d checkout flows", totalOrders)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 1; i <= totalOrders; i++ {
		orderID := fmt.Sprintf("order-%s", randomID(rng, 8))
		userID := users[rng.Intn(len(users))]
		product := products[rng.Intn(len(products))]
		amount := 10.0 + float64(rng.Intn(490))
		qty := 1 + rng.Intn(5)

		log.Printf("[%d/%d] Checkout | order=%s user=%s product=%s amount=%.2f qty=%d",
			i, totalOrders, orderID, userID, product, amount, qty)

		success := simulateCheckout(ctx, tracer, ordersPlacedCounter, rng, orderID, userID, product, amount, qty)

		outcome := "FAILED"
		if success {
			outcome = "SUCCESS"
		}
		log.Printf("[%d/%d] Result: %s", i, totalOrders, outcome)

		time.Sleep(time.Duration(300+rng.Intn(700)) * time.Millisecond)
	}

	log.Println("Flushing telemetry and shutting down SDK…")
	// shutdown (via defer) força o flush final dos exportadores de trace e métrica.

	log.Println("Done. Check Grafana at http://localhost:3000 to see your data.")
}

func randomID(rng *rand.Rand, n int) string {
	const charset = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}

// simulateCheckout cria o span raiz "checkout" e orquestra as etapas de
// inventory-check e payment, registrando a métrica orders.placed em caso de
// sucesso — mesmo fluxo da versão Java.
func simulateCheckout(
	ctx context.Context,
	tracer trace.Tracer,
	ordersPlacedCounter metric.Int64Counter,
	rng *rand.Rand,
	orderID, userID, product string,
	amount float64,
	qty int,
) bool {
	ctx, span := tracer.Start(ctx, "checkout",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("user.id", userID),
			attribute.String("order.id", orderID),
			attribute.String("order.product", product),
			attribute.Int("order.quantity", qty),
			attribute.Float64("order.amount", amount),
			attribute.String("order.currency", "USD"),
		),
	)
	defer span.End()

	inStock := simulateInventoryCheck(ctx, tracer, rng, product, qty)
	if !inStock {
		span.SetStatus(codes.Error, "Item out of stock")
		span.SetAttributes(attribute.String("checkout.outcome", "out_of_stock"))
		return false
	}

	paymentOK := simulatePayment(ctx, tracer, rng, orderID, userID, amount)
	if !paymentOK {
		span.SetStatus(codes.Error, "Payment declined")
		span.SetAttributes(attribute.String("checkout.outcome", "payment_declined"))
		return false
	}

	span.SetAttributes(attribute.String("checkout.outcome", "success"))
	span.SetStatus(codes.Ok, "")

	ordersPlacedCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("order.product", product),
		attribute.String("order.currency", "USD"),
	))

	return true
}

// simulateInventoryCheck cria o span filho "inventory-check", com ~15% de
// chance de reportar item fora de estoque, para exercitar o caminho de erro.
func simulateInventoryCheck(ctx context.Context, tracer trace.Tracer, rng *rand.Rand, product string, qty int) bool {
	_, span := tracer.Start(ctx, "inventory-check",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("inventory.product", product),
			attribute.Int("inventory.requested_qty", qty),
		),
	)
	defer span.End()

	time.Sleep(time.Duration(50+rng.Intn(150)) * time.Millisecond)

	inStock := rng.Float64() > 0.15
	availableQty := 0
	if inStock {
		availableQty = qty + rng.Intn(50)
	}

	span.SetAttributes(
		attribute.Int("inventory.available_qty", availableQty),
		attribute.Bool("inventory.in_stock", inStock),
	)

	if inStock {
		span.SetStatus(codes.Ok, "")
	} else {
		span.SetStatus(codes.Error, "Product out of stock")
		span.AddEvent("stock.exhausted", trace.WithAttributes(attribute.String("product", product)))
	}

	return inStock
}

// simulatePayment cria o span filho "payment", com ~10% de chance de recusa,
// para exercitar o caminho de erro.
func simulatePayment(ctx context.Context, tracer trace.Tracer, rng *rand.Rand, orderID, userID string, amount float64) bool {
	transactionID := fmt.Sprintf("txn-%s", randomID(rng, 12))

	_, span := tracer.Start(ctx, "payment",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("payment.order_id", orderID),
			attribute.String("payment.user_id", userID),
			attribute.Float64("payment.amount", amount),
			attribute.String("payment.currency", "USD"),
			attribute.String("payment.transaction_id", transactionID),
			attribute.String("payment.gateway", "stripe-mock"),
		),
	)
	defer span.End()

	time.Sleep(time.Duration(80+rng.Intn(220)) * time.Millisecond)

	approved := rng.Float64() > 0.10
	responseCode := "51"
	if approved {
		responseCode = "00"
	}

	span.SetAttributes(
		attribute.Bool("payment.approved", approved),
		attribute.String("payment.response_code", responseCode),
	)

	if approved {
		span.SetStatus(codes.Ok, "")
		span.AddEvent("payment.authorized", trace.WithAttributes(
			attribute.String("transaction.id", transactionID),
			attribute.Float64("amount", amount),
		))
	} else {
		span.SetStatus(codes.Error, "Payment declined by gateway")
		span.AddEvent("payment.declined", trace.WithAttributes(
			attribute.String("reason", "insufficient_funds"),
			attribute.String("transaction.id", transactionID),
		))
	}

	return approved
}

// buildSDK monta o SDK do OpenTelemetry: Resource (nome/versão do serviço),
// TracerProvider com exportador OTLP gRPC (batch) e MeterProvider com leitor
// periódico (5s) — mesma topologia da versão Java.
func buildSDK(ctx context.Context, otlpEndpoint, serviceName string) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, func(context.Context) error, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment("development"),
		),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("falha ao montar resource: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("falha ao criar trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithMaxQueueSize(2048),
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithExportTimeout(30*time.Second),
		),
	)

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(otlpEndpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("falha ao criar metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(5*time.Second))),
	)

	log.Println("OpenTelemetry SDK initialised (endpoint:", otlpEndpoint, ")")

	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var errs []error
		if err := tracerProvider.ForceFlush(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := meterProvider.ForceFlush(ctx); err != nil {
			errs = append(errs, err)
		}
		time.Sleep(2 * time.Second) // dá tempo do collector receber, como na versão Java
		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("erros no shutdown: %v", errs)
		}
		return nil
	}

	return tracerProvider, meterProvider, shutdown, nil
}
