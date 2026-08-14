# Gateway OpenRouter em Go

Porte em Go do projeto Spring Boot [`gateway-openrouter-java`](../gateway-openrouter-java) — API HTTP que expõe `POST /chat` e roteia a pergunta por uma cadeia de modelos de fallback via OpenRouter.

## O que foi mantido 1:1

- Mesmo endpoint `POST /chat`, mesmo contrato de request (`{"question": "..."}`, mínimo 5 caracteres) e response (`{"model": "...", "content": "..."}`).
- Mesma lógica de fallback: tenta cada modelo da lista em ordem, retorna a primeira resposta bem-sucedida; se todos falharem, erro `"all fallback models failed"`.
- Mesmos defaults de configuração: porta 3000, `temperature=0.2`, `max_tokens=100`, mesma lista de 5 modelos gratuitos, mesmo system prompt (`"You are a helpful assistant."`).
- Mesma porta e mesma base URL do OpenRouter (`https://openrouter.ai/api/v1`).

## O que foi adaptado

- **Sem Spring Boot / Spring AI**: o servidor HTTP usa `net/http` com `http.ServeMux` no lugar do `@RestController`; a validação de tamanho mínimo do campo `question` (que Java faz via `@Size(min=5)` + Bean Validation) é feita manualmente no handler.
- **Sem `ChatClient`/`OpenAiChatOptions`** do Spring AI: o cliente OpenRouter (`openrouter/client.go`) fala HTTP diretamente com o endpoint de chat completions, no mesmo espírito dos outros portes deste repositório.
- Configuração via variáveis de ambiente com prefixo `APP_` em vez de `application.properties` (`APP_SYSTEM_PROMPT`, `APP_FALLBACK_MODELS`, `APP_TEMPERATURE`, `APP_MAX_TOKENS`).

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY
go run .
```

```bash
curl -X POST http://localhost:3000/chat -d '{"question":"O que e machine learning?"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem o cliente OpenRouter (sucesso, API key vazia, erro de API), a cadeia de fallback (sucesso no primeiro modelo, fallback após falha, todos falhando) e o handler HTTP (sucesso, validação, método não permitido) — todos via `httptest`, sem depender da API real do OpenRouter.
