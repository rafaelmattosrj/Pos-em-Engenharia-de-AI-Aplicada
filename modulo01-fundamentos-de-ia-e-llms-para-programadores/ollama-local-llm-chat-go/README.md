# Ollama Local LLM Chat em Go

Porte em Go do projeto [`ollama-local-llm-chat-java`](../ollama-local-llm-chat-java) — chat com um modelo rodando localmente via [Ollama](https://ollama.com), usando sua API OpenAI-compatible (`http://localhost:11434/v1`).

## O que foi mantido 1:1

- Mesmas 5 perguntas de demonstração, mesma ordem.
- Mesmos defaults (`temperature=0.7`, 1 retry, endpoint/modelo configuráveis via `.env`).
- Mesma medição de tempo de resposta por pergunta.
- Mesma mensagem de orientação (`ollama serve` / `ollama pull <modelo>`) quando a chamada falha.

## O que foi adaptado

Como no porte `openrouter-multi-model-chat-go`, não há LangChain4j em Go — o cliente (`ollama/client.go`) fala HTTP diretamente com o endpoint `/chat/completions`, incluindo o retry manual que o LangChain4j fazia via `.maxRetries(1)`.

## Como executar

Pré-requisito: Ollama rodando localmente (`ollama serve`) com o modelo baixado (`ollama pull llama3.2`).

```bash
cp .env.example .env   # opcional, os defaults já funcionam
go run .
```

## Testes

```bash
go test ./...
```

Cobrem resposta de sucesso, retry após falha transitória e esgotamento das tentativas — usando `httptest.Server`, sem depender de um Ollama real rodando.
