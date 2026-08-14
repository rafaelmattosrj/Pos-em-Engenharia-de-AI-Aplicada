# Brag Bot API em Java

Porte em Java/Spring Boot da rota de API do backend Express/Angular SSR de [`modulo-05/brag-bot`](../modulo-05/brag-bot) — transforma um rascunho informal de uma realização profissional em um **"Brag Document"** executivo estruturado, usando o Gemini.

Segue a diretriz da skill `portar-projetos-java-go`: **apenas a lógica de backend (o flow Genkit + a rota Express) é portada — a UI Angular permanece fora de escopo.**

## Endpoint

`POST /api/brag`

```json
// Request
{ "definition": "otimizei o pool de conexões da api de pagamentos e reduzi os timeouts" }

// Response 200
{
  "id": "a1b2c3d4-...",
  "title": "Reduziu timeouts na API de pagamentos via otimização do connection pool",
  "context": "...",
  "actionTaken": "...",
  "businessImpact": "...",
  "metrics": ["50% reduction", "10ms latency"],
  "technologiesUsed": ["Java", "Redis"]
}
```

- `400` se `definition` estiver ausente ou vazio.
- `500` se a geração falhar (erro do Gemini ou resposta que não é um JSON válido).

## O que foi mantido 1:1

- Mesmo prompt (persona, objetivo, regras 1-4) do `bragGeneratorFlow` em `flows.ts`.
- Mesmo schema de saída (`title`, `context`, `actionTaken`, `businessImpact`, `metrics[]`, `technologiesUsed[]`), com o `id` gerado pelo **servidor** após a resposta do modelo — igual a `{...output, id: uuidv4()}` no flow original.
- Mesma temperatura (`0.8`).
- Mesma validação e códigos de status da rota Express (`400` sem `definition`, `500` em caso de falha).
- Mesmo modelo (`gemini-2.5-flash`) e mesma porta padrão (4000).

## O que foi adaptado

| Aspecto | TypeScript/Genkit | Java |
|---|---|---|
| Framework de IA | Genkit (`@genkit-ai/google-genai`, `ai.defineFlow`, `ai.generate` com `output.schema`) | Chamada direta à API REST do Gemini (`gemini/GeminiClient.java`) via `RestClient`, solicitando `responseMimeType: application/json` e fazendo o parse do JSON manualmente — o Genkit valida a saída contra um schema Zod automaticamente; aqui a validação de shape é implícita ao ler os campos esperados do JSON retornado |
| Framework web | Express (rota registrada em `server.ts`, ao lado do serving da SPA Angular) | Spring Boot `@RestController` |
| Geração de UUID | `uuid` (npm) | `java.util.UUID` |

## Como executar

```bash
export GEMINI_API_KEY=your-key-here
mvn spring-boot:run
```

```bash
curl -X POST http://localhost:4000/api/brag \
  -H "Content-Type: application/json" \
  -d '{"definition":"migrei o sistema de filas de RabbitMQ para Kafka reduzindo perda de mensagens"}'
```

## Testes

```bash
mvn test
```

O `GeminiClient` é mockado (`@MockBean`) para evitar chamadas reais à API do Gemini. Cobre: `400` sem `definition`, `200` com `BragDocument` completo, `500` quando o cliente Gemini lança exceção, `500` quando a resposta não é um JSON válido.

**Nota de compatibilidade**: o projeto compila e roda em Java 17+, mas os testes exigem JDK ≤ 21 nesta versão do Spring Boot (3.2.5) — o Mockito/Byte Buddy embutido não reconhece ainda o bytecode gerado por JDK 25. Rode `mvn test` com um JDK 17-21 (ex.: `JAVA_HOME` apontando para uma instalação Java 17).
