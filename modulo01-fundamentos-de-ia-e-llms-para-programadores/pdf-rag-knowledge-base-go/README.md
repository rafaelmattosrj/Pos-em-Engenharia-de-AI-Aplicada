# PDF RAG Knowledge Base em Go

Porte em Go do projeto [`pdf-rag-knowledge-base-java`](../pdf-rag-knowledge-base-java) — pipeline RAG (Retrieval-Augmented Generation) completo: PDF → chunks → embeddings → Neo4j → busca por similaridade → resposta gerada por LLM via OpenRouter.

## O que foi mantido 1:1

- Mesmo fluxo em 6 etapas (carregar PDF, dividir em chunks, carregar embeddings, limpar/popular Neo4j, configurar LLM, pipeline pergunta→busca→resposta).
- Mesmo chunking (1000 chars / 200 overlap) e mesmo schema no Neo4j (`Chunk`, índice `tensors_index`, cosseno).
- Mesmo prompt template (idêntico, incluindo as 8 instruções e o formato de contexto).
- Mesmas 5 perguntas de demonstração, mesma ordem.
- Mesmo filtro de relevância (`score > 0.5`) e mesma junção de contexto (`\n\n---\n\n`).
- Mesma configuração de LLM: `temperature=0.3`, `maxRetries=2`.
- Mesmo truncamento de mensagem de erro do LLM em 200 caracteres.

## O que foi adaptado

Reaproveita as mesmas soluções já documentadas em [`embeddings-vector-search-go`](../embeddings-vector-search-go) (que tem exatamente o mesmo pipeline de ingestão, sem a etapa de LLM):

| Aspecto | Java | Go |
|---|---|---|
| Modelo de embeddings | `AllMiniLmL6V2EmbeddingModel` (ONNX in-process) | Ollama local servindo `all-minilm` via HTTP |
| Parser de PDF | Apache PDFBox | [`github.com/ledongthuc/pdf`](https://github.com/ledongthuc/pdf) |
| Splitter recursivo | `DocumentSplitters.recursive` (LangChain4j) | Implementação própria (`splitter/`) |
| Neo4j driver | `neo4j-java-driver` + `Neo4jEmbeddingStore` | `neo4j-go-driver/v5` com Cypher manual (`store/`) |
| LLM (OpenRouter) | `OpenAiChatModel` (LangChain4j) | Cliente HTTP próprio (`openrouter/`), com retry manual equivalente a `.maxRetries(2)` |

## Pré-requisitos

- Go 1.22+
- Docker e Docker Compose
- Ollama local com o modelo de embeddings: `ollama serve` + `ollama pull all-minilm`
- Uma API key da [OpenRouter](https://openrouter.ai/keys)

## Configuração

```bash
cp ../pdf-rag-knowledge-base-java/tensores.pdf ./tensores.pdf
cp .env.example .env   # preencha OPENROUTER_API_KEY e NLP_MODEL
docker-compose up -d
```

## Como executar

```bash
go run .
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem o splitter, o cliente de embeddings e o cliente OpenRouter (sucesso, retry, API key vazia) via `httptest.Server` — sem depender de Ollama/Neo4j/OpenRouter reais.
