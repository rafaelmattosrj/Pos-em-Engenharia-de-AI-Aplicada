# Brag Bot API em Go

Porte em Go da rota de API do backend Express/Angular SSR de [`modulo-05/brag-bot`](../modulo-05/brag-bot) — transforma um rascunho informal de uma realização profissional em um **"Brag Document"** executivo estruturado, usando o Gemini. Ver também [`brag-bot-java`](../brag-bot-java), o porte Java irmão deste projeto.

Segue a diretriz da skill `portar-projetos-java-go`: **apenas a lógica de backend (o flow Genkit + a rota Express) é portada — a UI Angular permanece fora de escopo.**

## Endpoint

`POST /api/brag`

```json
// Request
{ "definition": "otimizei o pool de conexoes da api de pagamentos e reduzi os timeouts" }

// Response 200
{
  "id": "a1b2c3d4-...",
  "title": "Reduziu timeouts na API de pagamentos via otimização do connection pool",
  "context": "...",
  "actionTaken": "...",
  "businessImpact": "...",
  "metrics": ["50% reduction", "10ms latency"],
  "technologiesUsed": ["Go", "Redis"]
}
```

- `400` se `definition` estiver ausente ou vazio.
- `500` se a geração falhar (erro do Gemini ou resposta que não é um JSON válido).

## O que foi mantido 1:1

- Mesmo prompt (persona, objetivo, regras 1-4) do `bragGeneratorFlow` em `flows.ts`.
- Mesmo schema de saída (`title`, `context`, `actionTaken`, `businessImpact`, `metrics[]`, `technologiesUsed[]`), com o `id` gerado pelo **servidor** após a resposta do modelo.
- Mesma temperatura (`0.8`), mesmo modelo (`gemini-2.5-flash`) e mesma porta padrão (4000).
- Mesma validação e códigos de status da rota Express (`400` sem `definition`, `500` em caso de falha).

## O que foi adaptado

| Aspecto | TypeScript/Genkit | Go |
|---|---|---|
| Framework de IA | Genkit (`@genkit-ai/google-genai`, `ai.defineFlow`, `ai.generate` com `output.schema`) | Chamada direta à API REST do Gemini (`gemini/client.go`), solicitando `responseMimeType: application/json` e fazendo o parse manual — sem validação de schema automática, apenas leitura tolerante dos campos esperados |
| Framework web | Express (rota registrada em `server.ts`, ao lado do serving da SPA Angular) | `net/http` com `http.ServeMux` |
| Geração de UUID | `uuid` (npm) | `github.com/google/uuid` |

## Como executar

```bash
cp .env.example .env   # preencha GEMINI_API_KEY
go run .
```

```bash
curl -X POST http://localhost:4000/api/brag \
  -d '{"definition":"migrei o sistema de filas de RabbitMQ para Kafka reduzindo perda de mensagens"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem o cliente Gemini (sucesso, erro de API) e, via duplos de teste no lugar de chamadas reais, o `BragService` e o handler HTTP: `400` sem `definition`, `200` com `BragDocument` completo, `500` quando o Gemini falha, `500` quando a resposta não é um JSON válido — mesmos cenários de `BragBotApiTest.java`.
