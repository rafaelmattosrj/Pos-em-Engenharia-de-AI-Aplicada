package com.observability;

import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.LongCounter;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Scope;

import java.util.Random;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

/**
 * Simulates a checkout flow (inventory check + payment) instrumented with
 * OpenTelemetry spans, events and a counter metric.
 *
 * <p>Extracted from {@link Main} so the checkout logic (nested spans,
 * attributes, error status paths and the {@code orders.placed} counter) can
 * be unit-tested with an in-memory span exporter/metric reader, without
 * needing a real OTLP Collector.
 */
public class CheckoutSimulator {

    private final Tracer tracer;
    private final LongCounter ordersPlacedCounter;
    private final Random random;

    public CheckoutSimulator(Tracer tracer, LongCounter ordersPlacedCounter, Random random) {
        this.tracer = tracer;
        this.ordersPlacedCounter = ordersPlacedCounter;
        this.random = random;
    }

    /**
     * Simulates a full checkout: validates inventory, processes payment,
     * and records the completed order.
     *
     * @return {@code true} if the checkout succeeded
     */
    public boolean simulateCheckout(
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
            boolean inStock = simulateInventoryCheck(product, qty);

            if (!inStock) {
                checkoutSpan.setStatus(StatusCode.ERROR, "Item out of stock");
                checkoutSpan.setAttribute("checkout.outcome", "out_of_stock");
                return false;
            }

            // CHILD span 2: payment
            boolean paymentOk = simulatePayment(orderId, userId, amount);

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
    boolean simulateInventoryCheck(String product, int qty) throws InterruptedException {

        Span span = tracer.spanBuilder("inventory-check")
                .setSpanKind(SpanKind.CLIENT)
                .setAttribute("inventory.product", product)
                .setAttribute("inventory.requested_qty", (long) qty)
                .startSpan();

        try (Scope ignored = span.makeCurrent()) {
            // Simulate network latency to the inventory service
            TimeUnit.MILLISECONDS.sleep(50 + random.nextInt(150));

            // 15 % chance the item is out of stock
            boolean inStock = random.nextDouble() > 0.15;
            int availableQty = inStock ? qty + random.nextInt(50) : 0;

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
    boolean simulatePayment(String orderId, String userId, double amount) throws InterruptedException {

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
            TimeUnit.MILLISECONDS.sleep(80 + random.nextInt(220));

            // 10 % chance of payment failure
            boolean approved = random.nextDouble() > 0.10;

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
}
