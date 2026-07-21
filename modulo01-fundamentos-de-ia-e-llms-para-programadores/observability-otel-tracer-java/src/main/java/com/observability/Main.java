package com.observability;

import io.github.cdimascio.dotenv.Dotenv;
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.metrics.Meter;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Scope;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;
import io.opentelemetry.exporter.otlp.metrics.OtlpGrpcMetricExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.metrics.SdkMeterProvider;
import io.opentelemetry.sdk.metrics.export.PeriodicMetricReader;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import io.opentelemetry.semconv.ResourceAttributes;

import java.time.Duration;
import java.util.Random;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

/**
 * Demonstrates OpenTelemetry instrumentation in Java.
 *
 * <p>Simulates a simple e-commerce checkout flow with:
 * <ul>
 *   <li>Nested spans: checkout → payment + inventory-check</li>
 *   <li>Span attributes (user.id, order.id, amount, currency, …)</li>
 *   <li>A counter metric: orders.placed</li>
 *   <li>All telemetry exported to an OTLP Collector via gRPC</li>
 * </ul>
 *
 * <p>Environment variables (loaded from .env):
 * <pre>
 *   OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
 *   SERVICE_NAME=java-ecommerce
 * </pre>
 */
public class Main {

    private static final Logger LOG = Logger.getLogger(Main.class.getName());
    private static final Random RANDOM = new Random();

    // ── Simulated product catalogue ──────────────────────────────────────────
    private static final String[] PRODUCTS = {
        "prod-001-sneakers",
        "prod-002-headphones",
        "prod-003-laptop-bag",
        "prod-004-mechanical-keyboard",
        "prod-005-usb-hub"
    };

    private static final String[] USERS = {
        "user-42", "user-17", "user-88", "user-3", "user-99"
    };

    public static void main(String[] args) throws InterruptedException {

        // ── 1. Load configuration from .env ──────────────────────────────────
        Dotenv dotenv = Dotenv.configure()
                .ignoreIfMissing()
                .load();

        String otlpEndpoint = dotenv.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317");
        String serviceName  = dotenv.get("SERVICE_NAME", "java-ecommerce");

        LOG.info("=============================================================");
        LOG.info(" OpenTelemetry Java E-Commerce Demo");
        LOG.info("=============================================================");
        LOG.info(" Service name : " + serviceName);
        LOG.info(" OTLP endpoint: " + otlpEndpoint);
        LOG.info("=============================================================");

        // ── 2. Build an OpenTelemetry SDK instance ────────────────────────────
        OpenTelemetrySdk sdk = buildSdk(otlpEndpoint, serviceName);

        // ── 3. Obtain a Tracer and a Meter ────────────────────────────────────
        Tracer tracer = sdk.getTracer("com.observability.ecommerce", "1.0.0");
        Meter  meter  = sdk.getMeterProvider()
                           .get("com.observability.ecommerce");

        // ── 4. Define a counter for placed orders ─────────────────────────────
        LongCounter ordersPlacedCounter = meter
                .counterBuilder("orders.placed")
                .setDescription("Total number of orders successfully placed")
                .setUnit("{order}")
                .build();

        // ── 5. Simulate 5 checkout flows ──────────────────────────────────────
        int totalOrders = 5;
        LOG.info("Starting simulation – " + totalOrders + " checkout flows");

        for (int i = 1; i <= totalOrders; i++) {
            String orderId = "order-" + UUID.randomUUID().toString().substring(0, 8);
            String userId  = USERS[RANDOM.nextInt(USERS.length)];
            String product = PRODUCTS[RANDOM.nextInt(PRODUCTS.length)];
            double amount  = 10.0 + RANDOM.nextInt(490);   // $10 – $500
            int    qty     = 1 + RANDOM.nextInt(5);

            LOG.info(String.format("[%d/%d] Checkout | order=%s user=%s product=%s amount=%.2f qty=%d",
                    i, totalOrders, orderId, userId, product, amount, qty));

            boolean success = simulateCheckout(tracer, ordersPlacedCounter,
                    orderId, userId, product, amount, qty);

            LOG.info(String.format("[%d/%d] Result: %s", i, totalOrders,
                    success ? "SUCCESS" : "FAILED"));

            // Small pause between orders so spans appear separated in Tempo
            TimeUnit.MILLISECONDS.sleep(300 + RANDOM.nextInt(700));
        }

        // ── 6. Flush and shut down ────────────────────────────────────────────
        LOG.info("Flushing telemetry and shutting down SDK…");
        sdk.getSdkTracerProvider().forceFlush().join(10, TimeUnit.SECONDS);
        sdk.getSdkMeterProvider().forceFlush().join(10, TimeUnit.SECONDS);

        TimeUnit.SECONDS.sleep(2); // give the collector a moment to receive

        sdk.getSdkTracerProvider().shutdown().join(10, TimeUnit.SECONDS);
        sdk.getSdkMeterProvider().shutdown().join(10, TimeUnit.SECONDS);

        LOG.info("Done. Check Grafana at http://localhost:3000 to see your data.");
    }

    // ── Checkout root span ────────────────────────────────────────────────────

    /**
     * Simulates a full checkout: validates inventory, processes payment,
     * and records the completed order.
     *
     * @return {@code true} if the checkout succeeded
     */
    private static boolean simulateCheckout(
            Tracer tracer,
            LongCounter ordersPlacedCounter,
            String orderId,
            String userId,
            String product,
            double amount,
            int qty) throws InterruptedException {

        // ROOT span: checkout
        Span checkoutSpan = tracer.spanBuilder("checkout")
                .setSpanKind(SpanKind.SERVER)
                .setAttribute("user.id",    userId)
                .setAttribute("order.id",   orderId)
                .setAttribute("order.product", product)
                .setAttribute("order.quantity", (long) qty)
                .setAttribute("order.amount",   amount)
                .setAttribute("order.currency", "USD")
                .startSpan();

        try (Scope checkoutScope = checkoutSpan.makeCurrent()) {

            // CHILD span 1: inventory-check
            boolean inStock = simulateInventoryCheck(tracer, product, qty);

            if (!inStock) {
                checkoutSpan.setStatus(StatusCode.ERROR, "Item out of stock");
                checkoutSpan.setAttribute("checkout.outcome", "out_of_stock");
                return false;
            }

            // CHILD span 2: payment
            boolean paymentOk = simulatePayment(tracer, orderId, userId, amount);

            if (!paymentOk) {
                checkoutSpan.setStatus(StatusCode.ERROR, "Payment declined");
                checkoutSpan.setAttribute("checkout.outcome", "payment_declined");
                return false;
            }

            // All good – record the metric
            checkoutSpan.setAttribute("checkout.outcome", "success");
            checkoutSpan.setStatus(StatusCode.OK);

            ordersPlacedCounter.add(1,
                    Attributes.builder()
                            .put("order.product", product)
                            .put("order.currency", "USD")
                            .build());

            return true;

        } catch (Exception ex) {
            checkoutSpan.recordException(ex);
            checkoutSpan.setStatus(StatusCode.ERROR, ex.getMessage());
            return false;
        } finally {
            checkoutSpan.end();
        }
    }

    // ── Inventory-check child span ─────────────────────────────────────────

    /**
     * Simulates a call to the inventory service.
     * Randomly reports out-of-stock ~15 % of the time to exercise error paths.
     */
    private static boolean simulateInventoryCheck(
            Tracer tracer,
            String product,
            int qty) throws InterruptedException {

        Span span = tracer.spanBuilder("inventory-check")
                .setSpanKind(SpanKind.CLIENT)
                .setAttribute("inventory.product", product)
                .setAttribute("inventory.requested_qty", (long) qty)
                .startSpan();

        try (Scope ignored = span.makeCurrent()) {
            // Simulate network latency to the inventory service
            TimeUnit.MILLISECONDS.sleep(50 + RANDOM.nextInt(150));

            // 15 % chance the item is out of stock
            boolean inStock = RANDOM.nextDouble() > 0.15;
            int availableQty = inStock ? qty + RANDOM.nextInt(50) : 0;

            span.setAttribute("inventory.available_qty", (long) availableQty);
            span.setAttribute("inventory.in_stock", inStock);

            if (inStock) {
                span.setStatus(StatusCode.OK);
            } else {
                span.setStatus(StatusCode.ERROR, "Product out of stock");
                span.addEvent("stock.exhausted",
                        Attributes.builder()
                                .put("product", product)
                                .build());
            }

            return inStock;

        } finally {
            span.end();
        }
    }

    // ── Payment child span ────────────────────────────────────────────────────

    /**
     * Simulates a call to the payment gateway.
     * Randomly declines ~10 % of payments to exercise error paths.
     */
    private static boolean simulatePayment(
            Tracer tracer,
            String orderId,
            String userId,
            double amount) throws InterruptedException {

        String transactionId = "txn-" + UUID.randomUUID().toString().substring(0, 12);

        Span span = tracer.spanBuilder("payment")
                .setSpanKind(SpanKind.CLIENT)
                .setAttribute("payment.order_id",      orderId)
                .setAttribute("payment.user_id",       userId)
                .setAttribute("payment.amount",        amount)
                .setAttribute("payment.currency",      "USD")
                .setAttribute("payment.transaction_id", transactionId)
                .setAttribute("payment.gateway",       "stripe-mock")
                .startSpan();

        try (Scope ignored = span.makeCurrent()) {
            // Simulate latency to payment gateway
            TimeUnit.MILLISECONDS.sleep(80 + RANDOM.nextInt(220));

            // 10 % chance of payment failure
            boolean approved = RANDOM.nextDouble() > 0.10;

            span.setAttribute("payment.approved", approved);
            span.setAttribute("payment.response_code", approved ? "00" : "51");

            if (approved) {
                span.setStatus(StatusCode.OK);
                span.addEvent("payment.authorized",
                        Attributes.builder()
                                .put("transaction.id", transactionId)
                                .put("amount", amount)
                                .build());
            } else {
                span.setStatus(StatusCode.ERROR, "Payment declined by gateway");
                span.addEvent("payment.declined",
                        Attributes.builder()
                                .put("reason", "insufficient_funds")
                                .put("transaction.id", transactionId)
                                .build());
            }

            return approved;

        } finally {
            span.end();
        }
    }

    // ── SDK factory ──────────────────────────────────────────────────────────

    /**
     * Builds and wires the OpenTelemetry SDK:
     * <ul>
     *   <li>Resource: service name + version</li>
     *   <li>Tracer provider with OTLP gRPC exporter (BatchSpanProcessor)</li>
     *   <li>Meter provider with OTLP gRPC exporter (PeriodicMetricReader, 5 s)</li>
     * </ul>
     */
    private static OpenTelemetrySdk buildSdk(String otlpEndpoint, String serviceName) {

        // Resource describes this service to the collector
        Resource resource = Resource.getDefault()
                .merge(Resource.create(
                        Attributes.builder()
                                .put(ResourceAttributes.SERVICE_NAME,    serviceName)
                                .put(ResourceAttributes.SERVICE_VERSION, "1.0.0")
                                .put(ResourceAttributes.DEPLOYMENT_ENVIRONMENT, "development")
                                .build()
                ));

        // ── Trace exporter & provider ──────────────────────────────────────
        OtlpGrpcSpanExporter spanExporter = OtlpGrpcSpanExporter.builder()
                .setEndpoint(otlpEndpoint)
                .setTimeout(Duration.ofSeconds(10))
                .build();

        SdkTracerProvider tracerProvider = SdkTracerProvider.builder()
                .setResource(resource)
                .addSpanProcessor(
                        BatchSpanProcessor.builder(spanExporter)
                                .setMaxQueueSize(2048)
                                .setMaxExportBatchSize(512)
                                .setScheduleDelay(Duration.ofSeconds(5))
                                .setExporterTimeout(Duration.ofSeconds(30))
                                .build()
                )
                .build();

        // ── Metric exporter & provider ─────────────────────────────────────
        OtlpGrpcMetricExporter metricExporter = OtlpGrpcMetricExporter.builder()
                .setEndpoint(otlpEndpoint)
                .setTimeout(Duration.ofSeconds(10))
                .build();

        SdkMeterProvider meterProvider = SdkMeterProvider.builder()
                .setResource(resource)
                .registerMetricReader(
                        PeriodicMetricReader.builder(metricExporter)
                                .setInterval(Duration.ofSeconds(5))
                                .build()
                )
                .build();

        // ── Assemble the SDK and register it globally ──────────────────────
        OpenTelemetrySdk sdk = OpenTelemetrySdk.builder()
                .setTracerProvider(tracerProvider)
                .setMeterProvider(meterProvider)
                .buildAndRegisterGlobal();

        LOG.info("OpenTelemetry SDK initialised (endpoint: " + otlpEndpoint + ")");
        return sdk;
    }
}
