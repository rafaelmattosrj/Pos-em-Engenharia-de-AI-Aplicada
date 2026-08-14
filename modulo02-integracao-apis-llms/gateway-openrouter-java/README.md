# Gateway OpenRouter (Java / Spring Boot)

API HTTP em Spring Boot + Spring AI que expõe `POST /chat` e roteia a pergunta por uma cadeia de modelos de fallback via OpenRouter, retornando o primeiro modelo que responder com sucesso. Porte do "Smart Model Router Gateway" original (`01-smart-model-router-gateway`), servindo também de referência para o porte em Go [`gateway-openrouter-go`](../gateway-openrouter-go).

## Fluxo

```
POST /chat → ResilientChatClient tenta cada modelo em ordem → primeira resposta bem-sucedida → { model, content }
```

`ResilientChatClient` monta um `ChatClient` do Spring AI por modelo (`ChatClient.builder(chatModel).defaultOptions(OpenAiChatOptions.builder().model(model)...)`) e itera a lista de fallback até obter sucesso; `OpenRouterService` extrai `content` e `model` do `ChatResponse` retornado.

## O que foi mantido 1:1

- Mesmo endpoint `POST /chat`, mesmo contrato de request (`{"question": "..."}`, mínimo 5 caracteres via `@Size(min = 5)`) e response (`{"model": "...", "content": "..."}`).
- Mesma lógica de fallback: tenta cada modelo da lista em ordem, retorna a primeira resposta bem-sucedida; se todos falharem, lança erro (`"All fallback models failed"`).
- Mesmos defaults de configuração: porta 3000, `temperature=0.2`, `max_tokens=100`, mesma lista de 5 modelos gratuitos, mesmo system prompt (`"You are a helpful assistant."`).
- Mesma base URL do OpenRouter (`https://openrouter.ai/api/v1`).

## O que foi adaptado

- **Spring AI** (`ChatModel`/`ChatClient`/`OpenAiChatOptions`) no lugar de um cliente HTTP manual para o endpoint de chat completions do OpenRouter.
- `LlmResponse` como `record` Java (`model`, `content`) — equivalente ao `Response{Model, Content}` do porte Go.
- Validação de `question` via Bean Validation (`@Valid` + `@Size`) em vez de checagem manual de tamanho no handler.

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY
export OPENROUTER_API_KEY=...   # ou exporte a variável de outra forma
JAVA_HOME="C:/Program Files/Amazon Corretto/jdk25.0.3_9" mvn spring-boot:run
```

```bash
curl -X POST http://localhost:3000/chat \
  -H "Content-Type: application/json" \
  -d '{"question":"O que e machine learning?"}'
```

## Como testar

```bash
JAVA_HOME="C:/Program Files/Amazon Corretto/jdk25.0.3_9" mvn compile
JAVA_HOME="C:/Program Files/Amazon Corretto/jdk25.0.3_9" mvn test
```

Os testes cobrem, sem depender da API real do OpenRouter (`ChatModel` sempre mockado via Mockito):

- `ResilientChatClientTest` — fallback para o próximo modelo quando o primeiro falha, e erro quando todos falham.
- `OpenRouterServiceTest` — extração de `model`/`content` da `ChatResponse`, e propagação do erro quando todos os modelos de fallback falham.
- `ChatControllerTest` (`@WebMvcTest` + `MockMvc`) — contrato HTTP do `/chat` (sucesso), validação de `question` com menos de 5 caracteres (400) e método não permitido em `GET /chat` (405).
