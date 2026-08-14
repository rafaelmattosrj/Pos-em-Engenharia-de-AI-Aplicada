# OpenTelemetry E-Commerce Demo em Go

Porte em Go do projeto [`observability-otel-tracer-java`](../observability-otel-tracer-java), usando o SDK oficial [`go.opentelemetry.io/otel`](https://opentelemetry.io/docs/languages/go/).

## O que foi mantido 1:1

- Mesmo fluxo simulado: checkout com spans filhos `inventory-check` e `payment`.
- Mesmos atributos de span (`user.id`, `order.id`, `payment.amount` etc.), mesmas taxas de falha simuladas (~15% sem estoque, ~10% pagamento recusado).
- Mesma métrica de contador `orders.placed` com atributos `order.product`/`order.currency`.
- Mesma topologia de exportação: OTLP via gRPC, `BatchSpanProcessor` para traces e `PeriodicReader` (5s) para métricas.
- Mesmo carregamento de configuração via `.env` (`OTEL_EXPORTER_OTLP_ENDPOINT`, `SERVICE_NAME`), sem sobrescrever variáveis já exportadas no shell.

## O que foi adaptado

- Go não tem `try-with-resources`/`Scope` — o encerramento do span usa `defer span.End()`, idiomático em Go.
- `Attributes.builder()` do Java vira `attribute.String(...)`/`trace.WithAttributes(...)` do SDK Go.
- O SDK Go usa `otlptracegrpc`/`otlpmetricgrpc` como pacotes de exportador (equivalentes a `OtlpGrpcSpanExporter`/`OtlpGrpcMetricExporter`), com `WithInsecure()` explícito (comportamento padrão do exemplo, sem TLS, como o Collector local usado no laboratório).

## Pré-requisitos

- Go 1.22+
- Um OTel Collector rodando localmente (`localhost:4317`) — ex.: a stack Grafana/Tempo já usada no módulo.

## Como executar

```bash
cp .env.example .env   # opcional, os defaults já apontam para localhost:4317
go run .
```

## Estrutura

```
observability-otel-tracer-go/
├── main.go       # equivalente a Main.java
├── dotenv.go      # carregamento de .env + leitura com default (dotenv.get(key, default))
├── go.mod / go.sum
├── .env.example
└── README.md
```
