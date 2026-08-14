package com.observability;

import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.metrics.Meter;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.sdk.metrics.SdkMeterProvider;
import io.opentelemetry.sdk.metrics.data.MetricData;
import io.opentelemetry.sdk.testing.exporter.InMemoryMetricReader;
import io.opentelemetry.sdk.testing.exporter.InMemorySpanExporter;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.data.SpanData;
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Random;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.Mockito.lenient;
import static org.mockito.Mockito.when;

/**
 * Cobre os tres desfechos possiveis do fluxo de checkout demonstrado em
 * Main/README (checkout -> inventory-check -> payment):
 *  - sucesso: ambos os spans filhos OK, checkout OK, contador de pedidos incrementado
 *  - falta de estoque: checkout ERROR "out of stock", span de pagamento nao criado
 *  - pagamento recusado: checkout ERROR "payment declined", contador nao incrementado
 *
 * Usa InMemorySpanExporter/InMemoryMetricReader do OpenTelemetry SDK Testing
 * para inspecionar spans e metricas sem precisar de um Collector real, e um
 * Random mockado para forcar deterministicamente cada ramo (em vez de
 * depender das probabilidades de ~15%/~10% usadas em produção).
 */
@ExtendWith(MockitoExtension.class)
class CheckoutSimulatorTest {

    private InMemorySpanExporter spanExporter;
    private InMemoryMetricReader metricReader;
    private SdkTracerProvider tracerProvider;
    private SdkMeterProvider meterProvider;
    private LongCounter ordersPlacedCounter;

    @Mock
    private Random random;

    @BeforeEach
    void setUp() {
        spanExporter = InMemorySpanExporter.create();
        tracerProvider = SdkTracerProvider.builder()
                .addSpanProcessor(SimpleSpanProcessor.create(spanExporter))
                .build();

        metricReader = InMemoryMetricReader.create();
        meterProvider = SdkMeterProvider.builder()
                .registerMetricReader(metricReader)
                .build();

        Tracer tracer = tracerProvider.get("test-tracer");
        Meter meter = meterProvider.get("test-meter");
        ordersPlacedCounter = meter.counterBuilder("orders.placed").build();

        // nextInt() controla apenas a duracao do sleep simulado / quantidade
        // disponivel: mantido em 0 para tornar os testes rapidos e previsiveis.
        lenient().when(random.nextInt(anyInt())).thenReturn(0);
    }

    @AfterEach
    void tearDown() {
        tracerProvider.close();
        meterProvider.close();
    }

    private CheckoutSimulator newSimulator() {
        return new CheckoutSimulator(tracerProvider.get("test-tracer"), ordersPlacedCounter, random);
    }

    @Test
    void successfulCheckoutRecordsOkSpansAndIncrementsCounter() throws InterruptedException {
        // nextDouble() > 0.15 (estoque) e > 0.10 (pagamento): 0.5 aprova ambos
        when(random.nextDouble()).thenReturn(0.5);

        CheckoutSimulator simulator = newSimulator();
        boolean success = simulator.simulateCheckout("order-1", "user-1", "prod-001-sneakers", 99.90, 2);

        assertThat(success).isTrue();

        List<SpanData> spans = spanExporter.getFinishedSpanItems();
        assertThat(spans).extracting(SpanData::getName)
                .containsExactlyInAnyOrder("checkout", "inventory-check", "payment");

        SpanData checkoutSpan = findSpan(spans, "checkout");
        assertThat(checkoutSpan.getStatus().getStatusCode()).isEqualTo(StatusCode.OK);
        assertThat(checkoutSpan.getAttributes().get(AttributeKey.stringKey("checkout.outcome"))).isEqualTo("success");
        assertThat(checkoutSpan.getAttributes().get(AttributeKey.stringKey("order.id"))).isEqualTo("order-1");

        SpanData inventorySpan = findSpan(spans, "inventory-check");
        assertThat(inventorySpan.getStatus().getStatusCode()).isEqualTo(StatusCode.OK);
        assertThat(inventorySpan.getParentSpanId()).isEqualTo(checkoutSpan.getSpanId());

        SpanData paymentSpan = findSpan(spans, "payment");
        assertThat(paymentSpan.getStatus().getStatusCode()).isEqualTo(StatusCode.OK);
        assertThat(paymentSpan.getParentSpanId()).isEqualTo(checkoutSpan.getSpanId());

        List<MetricData> metrics = metricReader.collectAllMetrics().stream()
                .filter(m -> m.getName().equals("orders.placed"))
                .toList();
        assertThat(metrics).hasSize(1);
        long total = metrics.get(0).getLongSumData().getPoints().stream()
                .mapToLong(point -> point.getValue()).sum();
        assertThat(total).isEqualTo(1);
    }

    @Test
    void outOfStockCheckoutFailsBeforePaymentAndDoesNotIncrementCounter() throws InterruptedException {
        // nextDouble() <= 0.15: fora de estoque
        when(random.nextDouble()).thenReturn(0.1);

        CheckoutSimulator simulator = newSimulator();
        boolean success = simulator.simulateCheckout("order-2", "user-2", "prod-002-headphones", 50.0, 1);

        assertThat(success).isFalse();

        List<SpanData> spans = spanExporter.getFinishedSpanItems();
        assertThat(spans).extracting(SpanData::getName)
                .containsExactlyInAnyOrder("checkout", "inventory-check");

        SpanData checkoutSpan = findSpan(spans, "checkout");
        assertThat(checkoutSpan.getStatus().getStatusCode()).isEqualTo(StatusCode.ERROR);
        assertThat(checkoutSpan.getStatus().getDescription()).isEqualTo("Item out of stock");
        assertThat(checkoutSpan.getAttributes().get(AttributeKey.stringKey("checkout.outcome"))).isEqualTo("out_of_stock");

        SpanData inventorySpan = findSpan(spans, "inventory-check");
        assertThat(inventorySpan.getStatus().getStatusCode()).isEqualTo(StatusCode.ERROR);

        List<MetricData> metrics = metricReader.collectAllMetrics().stream()
                .filter(m -> m.getName().equals("orders.placed"))
                .toList();
        assertThat(metrics).isEmpty();
    }

    @Test
    void paymentDeclinedFailsCheckoutAfterInventoryPassesAndDoesNotIncrementCounter() throws InterruptedException {
        // 1a chamada (estoque): 0.5 -> em estoque | 2a chamada (pagamento): 0.05 -> recusado
        when(random.nextDouble()).thenReturn(0.5, 0.05);

        CheckoutSimulator simulator = newSimulator();
        boolean success = simulator.simulateCheckout("order-3", "user-3", "prod-003-laptop-bag", 250.0, 1);

        assertThat(success).isFalse();

        List<SpanData> spans = spanExporter.getFinishedSpanItems();
        assertThat(spans).extracting(SpanData::getName)
                .containsExactlyInAnyOrder("checkout", "inventory-check", "payment");

        SpanData checkoutSpan = findSpan(spans, "checkout");
        assertThat(checkoutSpan.getStatus().getStatusCode()).isEqualTo(StatusCode.ERROR);
        assertThat(checkoutSpan.getStatus().getDescription()).isEqualTo("Payment declined");
        assertThat(checkoutSpan.getAttributes().get(AttributeKey.stringKey("checkout.outcome"))).isEqualTo("payment_declined");

        SpanData inventorySpan = findSpan(spans, "inventory-check");
        assertThat(inventorySpan.getStatus().getStatusCode()).isEqualTo(StatusCode.OK);

        SpanData paymentSpan = findSpan(spans, "payment");
        assertThat(paymentSpan.getStatus().getStatusCode()).isEqualTo(StatusCode.ERROR);
        assertThat(paymentSpan.getStatus().getDescription()).isEqualTo("Payment declined by gateway");

        List<MetricData> metrics = metricReader.collectAllMetrics().stream()
                .filter(m -> m.getName().equals("orders.placed"))
                .toList();
        assertThat(metrics).isEmpty();
    }

    private SpanData findSpan(List<SpanData> spans, String name) {
        return spans.stream()
                .filter(s -> s.getName().equals(name))
                .findFirst()
                .orElseThrow(() -> new AssertionError("Span not found: " + name));
    }
}
