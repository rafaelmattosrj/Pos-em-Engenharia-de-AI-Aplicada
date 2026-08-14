# observability-otel-tracer

Java 17 + Maven demonstration of **OpenTelemetry** instrumentation.

The application simulates a simple e-commerce checkout flow and sends
**traces**, **metrics**, and structural **logs** to an OpenTelemetry Collector,
which fans the data out to Grafana Tempo (traces), Prometheus (metrics) and
Loki (logs).

---

## What is OpenTelemetry?

[OpenTelemetry (OTel)](https://opentelemetry.io/) is a vendor-neutral, CNCF-hosted
observability framework that provides a single, standardised API and SDK to
instrument applications for:

| Signal  | What it captures                                 | Backend used here |
|---------|--------------------------------------------------|-------------------|
| Traces  | Request flows across services (spans + context propagation) | Grafana Tempo     |
| Metrics | Numeric measurements over time (counters, histograms, …) | Prometheus        |
| Logs    | Structured log records                           | Grafana Loki      |

The key idea is **write once, export anywhere**: you instrument your code
against the OTel API once, and swap exporters (Jaeger, Zipkin, Datadog, …)
without touching application code.

---

## Architecture

```
┌─────────────────────────────┐
│   Java App (this project)   │
│                             │
│  checkout span              │
│  ├── inventory-check span   │
│  └── payment span           │
│                             │
│  orders.placed counter      │
└────────────┬────────────────┘
             │ OTLP / gRPC :4317
             ▼
┌────────────────────────────────┐
│   OpenTelemetry Collector      │
│   (otel/opentelemetry-collector│
│    running in Docker)          │
└──────┬────────────┬────────────┘
       │            │            │
       ▼            ▼            ▼
  Grafana Tempo  Prometheus    Grafana Loki
  (traces)       (metrics)     (logs)
       │            │            │
       └────────────┴────────────┘
                    │
                    ▼
              Grafana :3000
              (dashboards)
```

The infrastructure (`docker-compose-infra.yaml`) is located in the sibling
project `../exemplo-09-grafana-mcp/alumnus/infra/` and is **shared** between
the Node.js reference app and this Java application.

---

## How the Grafana Stack integrates with OTel

### Tempo (Traces)

Grafana Tempo receives spans via **OTLP gRPC** from the collector.
In Grafana, open **Explore → Tempo** and search by `service.name = java-ecommerce`
to see the checkout → payment / inventory-check trace trees.

### Prometheus (Metrics)

The collector scrapes metrics from itself on port `8889` (Prometheus exporter)
and Prometheus scrapes that endpoint.  The `orders_placed_total` counter
created in this app will appear in Grafana's **Explore → Prometheus** after the
first export interval (~5 s).

### Loki (Logs)

The collector forwards log records to Loki via OTLP HTTP.
Application logs emitted through `java.util.logging` / SLF4J are captured as
OTel log records and can be searched in **Explore → Loki** using the label
`{service_name="java-ecommerce"}`.

---

## Project structure

```
observability-otel-tracer/
├── pom.xml                                 Maven build descriptor
├── .env.example                            Copy to .env and adjust
└── src/
    ├── main/java/com/observability/
    │   ├── Main.java                       Application entry point (config, SDK wiring, run loop)
    │   └── CheckoutSimulator.java          Checkout flow (spans, attributes, counter) — testable in isolation
    └── test/java/com/observability/
        └── CheckoutSimulatorTest.java      Unit tests for the checkout flow
```

### Responsibilities

| Class / Method                  | Responsibility                                                   |
|----------------------------------|------------------------------------------------------------------|
| `Main.main`                      | Loads .env, builds SDK, runs 5 simulated checkouts, shuts down  |
| `Main.buildSdk`                  | Wires `SdkTracerProvider` + `SdkMeterProvider` with OTLP gRPC   |
| `CheckoutSimulator.simulateCheckout`     | Root `checkout` span (SpanKind.SERVER) with order attributes     |
| `CheckoutSimulator.simulateInventoryCheck` | Child `inventory-check` span (SpanKind.CLIENT), 15 % failure   |
| `CheckoutSimulator.simulatePayment`      | Child `payment` span (SpanKind.CLIENT), 10 % decline            |

`CheckoutSimulator` was extracted from `Main` (same logic, unchanged behavior) specifically so the
checkout flow — nested spans, attributes, error status paths, and the `orders.placed` counter —
could be unit-tested with an in-memory exporter instead of requiring a live OTel Collector.

---

## Getting started

### Prerequisites

- Java 17+
- Maven 3.8+
- Docker & Docker Compose

### Step 1 – Start the observability infrastructure

The Grafana / Collector / Tempo / Prometheus / Loki stack lives in the sibling
project directory:

```bash
cd ../exemplo-09-grafana-mcp/alumnus

# Start all containers (collector, grafana, tempo, prometheus, loki)
docker compose -f infra/docker-compose-infra.yaml up -d --wait
```

Wait until all containers report healthy (usually ~30 s).
Grafana will be available at <http://localhost:3000> (no login required).

### Step 2 – Configure this project

```bash
cd observability-otel-tracer

# Copy the example env file
cp .env.example .env

# Edit if necessary (defaults work out of the box with the local stack)
# OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
# SERVICE_NAME=java-ecommerce
```

### Step 3 – Build and run

```bash
# Option A – compile and run directly with Maven
mvn compile exec:java

# Option B – build a fat JAR, then run it
mvn package
java -jar target/observability-otel-tracer-1.0.0.jar
```

### Step 4 – Explore the data in Grafana

Open <http://localhost:3000> and navigate to:

- **Explore → Tempo** — search for `service.name = java-ecommerce` to see
  the trace trees (checkout → payment + inventory-check).
- **Explore → Prometheus** — query `orders_placed_total` to see the counter.
- **Explore → Loki** — use label filter `{service_name="java-ecommerce"}` to
  read structured log lines.

### Step 5 – Tear down

```bash
cd ../exemplo-09-grafana-mcp/alumnus
docker compose -f infra/docker-compose-infra.yaml down
```

---

## Testing

```bash
mvn test
```

Tests run fully offline — no OTel Collector, Docker, or network required. They use
`io.opentelemetry:opentelemetry-sdk-testing` (`InMemorySpanExporter` / `InMemoryMetricReader`) to
capture spans/metrics in-process, and a Mockito-mocked `Random` to deterministically force each
branch instead of relying on the ~15 %/~10 % production probabilities.

`CheckoutSimulatorTest` covers the three possible outcomes of a checkout:

- **Success** — `checkout`, `inventory-check` and `payment` spans all end with `StatusCode.OK`,
  `checkout.outcome=success`, and the `orders.placed` counter is incremented by 1.
- **Out of stock** — `inventory-check` fails, `checkout` ends with `StatusCode.ERROR` /
  `"Item out of stock"` / `checkout.outcome=out_of_stock`, no `payment` span is created, and the
  counter is not incremented.
- **Payment declined** — `inventory-check` passes but `payment` fails; `checkout` ends with
  `StatusCode.ERROR` / `"Payment declined"` / `checkout.outcome=payment_declined`, and the counter
  is not incremented.

---

## Dependencies

| Artifact | Version | Purpose |
|---|---|---|
| `io.opentelemetry:opentelemetry-api` | 1.44.1 | OTel API (Tracer, Meter, …) |
| `io.opentelemetry:opentelemetry-sdk` | 1.44.1 | SDK implementation |
| `io.opentelemetry:opentelemetry-exporter-otlp` | 1.44.1 | OTLP gRPC exporter |
| `io.opentelemetry:opentelemetry-sdk-extension-autoconfigure` | 1.44.1 | Auto-configure support |
| `io.github.cdimascio:dotenv-java` | 3.0.0 | `.env` file loading |
| `org.slf4j:slf4j-simple` | 2.0.16 | Console log implementation |

---

## Relation to the Node.js reference project

The Node.js application in `../exemplo-09-grafana-mcp/alumnus/` uses the
identical OTLP gRPC pipeline (`@opentelemetry/exporter-trace-otlp-grpc`,
`@opentelemetry/exporter-metrics-otlp-grpc`, `@opentelemetry/exporter-logs-otlp-grpc`)
to instrument a Fastify HTTP server.

This Java project replicates the same **concept** — send traces/metrics/logs to
the shared collector — but from a standalone Java process instead of a running
HTTP server, making it easy to understand the OTel SDK without HTTP framework
complexity.
