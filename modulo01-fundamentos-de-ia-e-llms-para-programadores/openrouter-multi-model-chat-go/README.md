# OpenRouter Multi-Model Chat em Go

Porte em Go do projeto [`openrouter-multi-model-chat-java`](../openrouter-multi-model-chat-java), que por sua vez replica o exemplo original em shell/curl (Exemplo 11) usando [OpenRouter](https://openrouter.ai) como gateway unificado de modelos de linguagem.

---

## O que foi mantido 1:1 em relação à versão Java

- Mesma pergunta enviada aos três mesmos modelos gratuitos.
- Mesmo formato de saída no terminal (separadores, cabeçalhos).
- Mesma estratégia de configuração via `.env` (chave `OPENROUTER_API_KEY`), sem sobrescrever variáveis já exportadas no shell.
- Mesmo timeout de 60s por requisição.
- Suporte aos headers opcionais `HTTP-Referer` e `X-Title` recomendados pelo OpenRouter — que na versão Java **não** eram enviados porque o `OpenAiChatModel` do LangChain4j não expõe headers customizados na API de alto nível. Na versão Go, como o cliente HTTP é escrito à mão, esses headers já estão implementados (campos `Referer`/`Title` do `Client`).

## O que foi adaptado por não ter equivalente direto no ecossistema Go

O Go não tem uma biblioteca de orquestração de LLMs tão madura e amplamente adotada quanto o **LangChain4j** do Java. Em vez de forçar uma abstração pesada, o porte usa um cliente HTTP mínimo (pacote `openrouter/`, apenas `net/http` + `encoding/json` da biblioteca padrão) que fala diretamente com o endpoint `/chat/completions` do OpenRouter — no mesmo espírito do exemplo em `curl` que inspirou o projeto originalmente. Essa escolha é deliberada: um cliente HTTP explícito e testável é mais idiomático em Go do que replicar a API fluente de builder do LangChain4j.

---

## Modelos gratuitos demonstrados

| Modelo | Provedor | Parâmetros |
|---|---|---|
| `google/gemma-3-27b-it:free` | Google | 27B |
| `meta-llama/llama-3.2-3b-instruct:free` | Meta | 3B |
| `mistralai/mistral-7b-instruct:free` | Mistral | 7B |

---

## Pré-requisitos

- Go 1.22+
- Uma conta gratuita no OpenRouter e uma chave de API (veja instruções no README do projeto Java irmão)

## Configuração

```bash
cp .env.example .env
# edite .env e cole sua chave real
```

## Como executar

```bash
go run .
```

## Como rodar os testes

```bash
go test ./...
```

Os testes usam `httptest.Server` para simular a API do OpenRouter — não fazem chamada de rede real, cobrindo resposta de sucesso, chave de API vazia e erro retornado pela API (ex.: rate limit).

---

## Estrutura do projeto

```
openrouter-multi-model-chat-go/
├── main.go              # Entry point — equivalente a Main.java
├── dotenv.go             # Carregamento simples de .env (equivalente ao dotenv-java)
├── openrouter/
│   ├── client.go          # Cliente HTTP do OpenRouter — equivalente ao uso do OpenAiChatModel
│   └── client_test.go     # Testes com servidor HTTP simulado
├── go.mod
├── .env.example
└── README.md
```

## Como funciona

```go
client := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))
resposta, err := client.Complete(ctx, "google/gemma-3-27b-it:free", "Me conte uma curiosidade sobre LLMs.")
```

O `Client.Complete` monta o payload JSON no mesmo formato aceito pela API compatível com OpenAI do OpenRouter, envia via `http.Client` com timeout de 60s, e decodifica a primeira `choice` da resposta — comportamento equivalente a `chatModel.generate(mensagem)` na versão Java.
