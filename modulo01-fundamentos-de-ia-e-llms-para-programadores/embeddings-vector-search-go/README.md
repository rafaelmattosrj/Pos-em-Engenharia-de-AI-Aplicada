# Embeddings + Vector Search com Neo4j em Go

Porte em Go do projeto [`embeddings-vector-search-java`](../embeddings-vector-search-java) — geração de embeddings localmente e busca por similaridade em Neo4j, **sem nenhuma chamada a LLM**.

## O que foi mantido 1:1

- Mesmo fluxo em 5 etapas: carregar PDF → dividir em chunks → gerar embeddings → armazenar no Neo4j → buscar por similaridade.
- Mesmo chunking: 1000 caracteres por chunk, 200 de overlap.
- Mesmas 5 perguntas de demonstração, mesma ordem.
- Mesmo schema no Neo4j: nós `Chunk` com propriedades `text`/`embedding`, índice vetorial `tensors_index`, similaridade de cosseno, top-3 resultados.
- Mesmo preview de texto (300 caracteres) e formatação de saída.
- Neo4j via Docker Compose (`docker-compose.yml`, mesmas credenciais `neo4j/password`).

## O que foi adaptado

| Aspecto | Java | Go | Motivo |
|---|---|---|---|
| Modelo de embeddings | `AllMiniLmL6V2EmbeddingModel` — ONNX rodando in-process | Ollama local servindo `all-minilm` via HTTP (`embeddings/ollama.go`) | Go não tem um binding ONNX Runtime conveniente sem cgo. Ollama mantém o modelo 100% local (sem API paga, sem internet), só muda para um processo separado em vez de in-process — mesma dimensão (384) e mesma família de modelo. |
| Parser de PDF | Apache PDFBox | [`github.com/ledongthuc/pdf`](https://github.com/ledongthuc/pdf) (`pdfx/extract.go`) | Biblioteca pura-Go para extração de texto de PDF, sem dependências nativas. |
| Splitter recursivo | `DocumentSplitters.recursive` (LangChain4j) | Implementação própria (`splitter/splitter.go`) | Sem equivalente pronto em Go; reimplementado com a mesma estratégia (parágrafo → linha → frase → palavra, empacotados até `chunkSize` com overlap). |
| Neo4j driver | `neo4j-java-driver` + `Neo4jEmbeddingStore` | `neo4j-go-driver/v5` com Cypher manual (`store/neo4j_store.go`) | O driver oficial Go não tem um wrapper de embedding store como o LangChain4j; o Cypher (criação de índice vetorial, inserção, `db.index.vector.queryNodes`) foi escrito diretamente. |

## Pré-requisitos

- Go 1.22+
- Docker e Docker Compose
- [Ollama](https://ollama.com) rodando localmente com o modelo de embeddings baixado:
  ```bash
  ollama serve
  ollama pull all-minilm
  ```

## Configuração

### 1. Copiar o PDF

```bash
cp ../pdf-rag-knowledge-base-java/tensores.pdf ./tensores.pdf
```

### 2. Variáveis de ambiente

```bash
cp .env.example .env
```

### 3. Subir o Neo4j

```bash
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

Cobrem o splitter recursivo (chunking, overlap, hard-split sem separadores) e o cliente de embeddings (sucesso, erro do servidor, embedding vazio) via `httptest.Server` — sem depender de Ollama/Neo4j reais.
