# PSP Routing Intelligence — Go

Sistema de recomendação de Payment Service Provider (PSP) via **Embeddings + RAG + LLM**:
para cada transação recebida, busca transações históricas similares no Neo4j e usa um LLM
para recomendar o PSP com maior chance de aprovação, com justificativa em pt-BR.

---

## Origem deste projeto

Este é o **porte Go de [`../psp-routing-intelligence-java`](../psp-routing-intelligence-java)**
— e não diretamente da especificação em `IDEIA.md`. A versão Java, por sua vez, foi a primeira
implementação real daquela especificação (a pasta Java só tinha `IDEIA.md` + um `pom.xml` vazio
antes dela existir — ver o README da versão Java para o histórico completo).

O contrato HTTP, o pipeline de 4 passos, o prompt RAG e os 50 dados de seed foram mantidos
**1:1** com a versão Java. O que muda é a forma de acessar cada dependência externa, adaptada
para o que é idiomático em Go:

| Aspecto | Java (`psp-routing-intelligence-java`) | Go (este projeto) |
|---|---|---|
| Servidor HTTP | Spring Boot (`spring-boot-starter-web`) | `net/http` puro — `http.ServeMux` com padrões de método (`"POST /api/routing/recommend"`, Go 1.22+); só 2 rotas, não justifica `chi`/`gin` (ver Passo 2 da skill de portabilidade) |
| Modelo de embeddings | `AllMiniLmL6V2EmbeddingModel` (ONNX in-process, LangChain4j) | Ollama local servindo `all-minilm` via HTTP (`embeddings/ollama.go`) — Go não tem um binding ONNX Runtime conveniente sem cgo; mesmo padrão já usado em `pdf-rag-knowledge-base-go` e `embeddings-vector-search-go` |
| LLM via OpenRouter | `OpenAiChatModel` (LangChain4j) | Cliente HTTP próprio (`openrouter/client.go`), com os mesmos defaults (`temperature=0.3`, `maxRetries=2`) — mesmo pacote/padrão de `pdf-rag-knowledge-base-go` |
| Neo4j | `Neo4jEmbeddingStore` (LangChain4j) + `neo4j-java-driver` | `neo4j-go-driver/v5` com Cypher manual (`store/neo4j_store.go`) — nó `Transaction` com propriedades `text`, `embedding`, `psp`, `status`, `amount`, `method`; índice vetorial `psp_routing_index` criado explicitamente via `CREATE VECTOR INDEX` |
| Validação do corpo da requisição | Bean Validation (`jakarta.validation` no `RecommendRequest`) | Validação manual em `httpserver.recommendRequest.validate()`, retornando o mesmo shape de erro (`{"error":"VALIDATION_ERROR","fields":{...}}`) |
| Injeção de dependência | Beans Spring (`@Service`, `@Configuration`) | Campos de função em `routing.Service`/`routing.SeedService` (`Embed`, `FindSimilar`, `Chat`, `Store`, `Clear`) — mesmo espírito do `ChatFn` funcional usado em `RagAnswerService` na versão Java, e o jeito idiomático em Go de trocar as dependências por fakes nos testes sem mocking framework |
| Template do prompt / seed data | Carregados do classpath (`resources/prompts`, `resources/data`) | Embutidos no binário via `//go:embed` (`routing/routing-prompt.txt`, `routing/seed_transactions.json`) — sem depender de arquivos externos em tempo de execução |
| Carregamento de `.env` | `dotenv-java` | `loadDotEnv` próprio (`dotenv.go`), mesmo padrão de `pdf-rag-knowledge-base-go`/`embeddings-vector-search-go` |

**Escopo backend apenas** (herdado da versão Java): o `IDEIA.md` original também descreve um
frontend React/Vite, que não foi portado nesta versão — apenas o backend/API. O CORS continua
liberado em `/api/**` para um cliente local futuro.

---

## Arquitetura

```
psp-routing-intelligence-go/
├── go.mod
├── main.go                    ← carrega .env, conecta Neo4j, sobe o servidor HTTP
├── dotenv.go                  ← loader de .env
├── docker-compose.yml         ← Neo4j
├── .env.example
├── data/transactions-seed.json  ← cópia de leitura (a fonte real embutida está em routing/)
├── domain/
│   └── domain.go              ← PSP, PaymentMethod, TransactionStatus, Transaction,
│                                 HistoricalTransaction, SimilarCase, RoutingRecommendation
├── routing/                    ← equivalente a com.psprouting.application
│   ├── text.go                ← (1) transação → texto natural (TransactionTextSerializer)
│   ├── prompt.go               ← monta o prompt RAG (go:embed routing-prompt.txt)
│   ├── routing-prompt.txt      ← template extraído 1:1 de IDEIA.md
│   ├── parser.go                ← decodifica o JSON de resposta do LLM
│   ├── service.go               ← (4) orquestra o pipeline completo
│   ├── seed.go                  ← carrega o seed (go:embed seed_transactions.json) e popula o Neo4j
│   └── seed_transactions.json   ← as mesmas 50 transações da versão Java
├── embeddings/
│   └── ollama.go                ← (1) texto → vetor 384d via Ollama
├── openrouter/
│   └── client.go                 ← (3) chamada ao LLM via OpenRouter
├── store/
│   └── neo4j_store.go            ← (2) persistência e busca por similaridade no Neo4j
└── httpserver/
    └── server.go                  ← POST /api/routing/recommend, POST /api/routing/seed
```

## Pipeline (`POST /api/routing/recommend`)

Idêntico ao da versão Java:

```
① routing.SerializeTransaction + embeddings.Client.Embed
    → "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA,
       regiao SP, hora 20h, categoria STREAMING."
    → vetor []float64 de 384 dimensões via Ollama (all-minilm)

② store.Store.FindSimilar
    → busca as 5 transações históricas mais similares (cosine similarity) no Neo4j

③ routing.BuildPrompt + openrouter.Client.Chat + routing.ParseRecommendation
    → monta o prompt RAG, chama o LLM via OpenRouter, decodifica
      { primary, confidence, reasoning, fallback[] }

④ routing.Service.Recommend combina a recomendação do LLM com os similarCases encontrados
```

---

## Pré-requisitos

- Go 1.22+ (usa `http.ServeMux` com padrões de método)
- Docker e Docker Compose (para o Neo4j)
- [Ollama](https://ollama.com) rodando localmente com `ollama pull all-minilm`
- Uma API key da [OpenRouter](https://openrouter.ai/keys)

## Como rodar

```bash
# 1. Suba o Neo4j
docker-compose up -d

# 2. Configure as variáveis de ambiente
cp .env.example .env
# edite .env com sua OPENROUTER_API_KEY

# 3. Garanta que o Ollama está servindo o modelo de embeddings
ollama pull all-minilm

# 4. Rode o servidor
go run .

# 5. Carregue os dados históricos
curl -X POST http://localhost:8080/api/routing/seed

# 6. Peça uma recomendação
curl -X POST http://localhost:8080/api/routing/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500.00,
    "method": "CREDIT_CARD",
    "brand": "VISA",
    "userRegion": "SP",
    "hour": 20,
    "merchantCategory": "STREAMING"
  }'
```

## Testes

```bash
go test ./...
```

24 testes com o pacote `testing` padrão, sem dependência de Neo4j, Ollama ou de uma chamada
HTTP real ao OpenRouter — a lógica de cada etapa foi extraída em funções e clientes puros,
testados isoladamente (mesmo espírito da versão Java, adaptado para o padrão Go de fakes via
campos de função e `httptest`):

| Pacote de teste | Cobre |
|---|---|
| `routing` (`text_test.go`) | Serialização com/sem bandeira |
| `routing` (`prompt_test.go`) | Placeholders substituídos, instruções fixas do template preservadas |
| `routing` (`parser_test.go`) | JSON válido, cerca markdown, campos ausentes, PSP desconhecido (no `primary` e no `fallback`) |
| `routing` (`service_test.go`) | Orquestração dos 4 passos, combinação de `SimilarCases`, propagação de erro de embedding/parsing |
| `routing` (`seed_test.go`) | Carrega as 50 transações embutidas, propaga erro de `Clear`/`Embed` |
| `embeddings` (`ollama_test.go`) | Sucesso, erro HTTP 500, embedding vazio (via `httptest`) |
| `openrouter` (`client_test.go`) | Sucesso, API key vazia, retry em falha, erro da API propagado |
| `httpserver` (`server_test.go`) | Contrato HTTP dos dois endpoints: 200, 400 (validação), 502 (resposta inválida do LLM), 405 (método errado) |

`store/neo4j_store.go` não tem teste unitário — depende de um Neo4j real (mesma decisão da
versão Java, que também deixa `Neo4jTransactionRepository` fora dos testes unitários, cobrindo
esse caminho apenas via teste manual/`curl`).

## Build

```bash
go build ./...   # sem erro
go vet ./...     # sem erro
go test ./...    # ok em todos os pacotes com testes
```
