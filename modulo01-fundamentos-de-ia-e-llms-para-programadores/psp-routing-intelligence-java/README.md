# PSP Routing Intelligence — Java + Spring Boot + LangChain4j

Sistema de recomendação de Payment Service Provider (PSP) via **Embeddings + RAG + LLM**:
para cada transação recebida, busca transações históricas similares no Neo4j e usa um LLM
para recomendar o PSP com maior chance de aprovação, com justificativa em pt-BR.

---

## Origem deste projeto

Este projeto **não é o porte de um exemplo original em TypeScript/Python** (diferente dos
demais pares do módulo 01, como `pdf-rag-knowledge-base-java`/`-go`, portados de
`exemplo-13-embeddings-neo4j-rag`). Ele nasceu de [`IDEIA.md`](./IDEIA.md) — uma
especificação escrita por Rafael Mattos Moreira descrevendo o problema (roteamento estático
de PSPs no gateway de pagamentos), a solução proposta (embeddings + RAG + LLM) e o contrato
completo da API (schema de request/response, pipeline de 4 passos, prompt RAG, padrões de
dados de seed).

Antes desta implementação, a pasta continha **apenas** `IDEIA.md` e um `backend/pom.xml` com
as dependências planejadas — nenhum arquivo `.java` existia. Este projeto é, portanto, a
**primeira implementação real** da especificação, escrita seguindo à risca o contrato definido
em `IDEIA.md` (endpoints, JSON de entrada/saída, prompt, dados de seed) e reaproveitando os
padrões já validados nos outros projetos Java do módulo:

| Reaproveitado de | Padrão |
|---|---|
| `embeddings-vector-search-java` | `AllMiniLmL6V2EmbeddingModel` (ONNX local) + `Neo4jEmbeddingStore` (builder, label, índice) |
| `pdf-rag-knowledge-base-java` | `OpenAiChatModel` apontando para o OpenRouter, extração de lógica de negócio em classes puras/testáveis (`PromptBuilder` → `RoutingPromptBuilder`, `RagAnswerService` → `ReasoningService`) |
| `openrouter-multi-model-chat-java` | Carregamento de configuração via `dotenv-java` |

**Escopo backend apenas**: o `IDEIA.md` também descreve um frontend React/Vite. Por decisão
explícita ao criar esta versão, apenas o **backend/API** foi implementado — o frontend não
faz parte deste porte (a API expõe CORS liberado em `/api/**` para que um frontend real, se
construído depois, consiga consumi-la sem bloqueio do navegador).

## Adaptação de versão

`IDEIA.md` menciona Java 21, mas o ambiente disponível e os demais projetos Java do módulo
(`embeddings-vector-search-java`, `pdf-rag-knowledge-base-java`, `openrouter-multi-model-chat-java`)
usam **Java 17** — o `pom.xml` foi ajustado para 17 para manter consistência com o resto do
módulo e com a toolchain real.

---

## Arquitetura

```
backend/
├── pom.xml
├── docker-compose.yml                     ← Neo4j
├── .env.example
└── src/main/java/com/psprouting/
    ├── PspRoutingApplication.java         ← carrega .env, sobe o Spring Boot
    ├── domain/
    │   ├── Transaction.java               ← features da transação recebida
    │   ├── HistoricalTransaction.java     ← transação de seed (com psp + status)
    │   ├── PSP.java, PaymentMethod.java, TransactionStatus.java
    │   ├── SimilarCase.java               ← caso histórico recuperado por similaridade
    │   └── RoutingRecommendation.java     ← resposta final de /recommend
    ├── application/
    │   ├── TransactionTextSerializer.java ← (1) transação → texto natural
    │   ├── EmbeddingService.java          ← (1) texto → vetor (384d, ONNX local)
    │   ├── SimilaritySearchService.java   ← (2) busca top-5 casos similares no Neo4j
    │   ├── RoutingPromptBuilder.java      ← monta o prompt RAG a partir do template
    │   ├── RecommendationParser.java      ← decodifica o JSON de resposta do LLM
    │   ├── ReasoningService.java          ← (3) prompt + chamada ao LLM + parsing
    │   ├── RoutingService.java            ← (4) orquestra o pipeline completo
    │   └── SeedService.java               ← carrega o seed e popula o Neo4j
    ├── infrastructure/
    │   ├── Neo4jTransactionRepository.java
    │   └── config/
    │       ├── AppConfig.java             ← beans: EmbeddingModel, ChatModel, Neo4jEmbeddingStore
    │       └── CorsConfig.java
    ├── controller/
    │   ├── RoutingController.java         ← POST /api/routing/recommend, POST /api/routing/seed
    │   └── GlobalExceptionHandler.java    ← 400 (validação) / 502 (resposta inválida do LLM)
    └── dto/
        ├── RecommendRequest.java          ← Bean Validation do corpo de /recommend
        └── SeedResponse.java

└── src/main/resources/
    ├── application.properties
    ├── prompts/routing-prompt.txt         ← template do prompt RAG (extraído 1:1 de IDEIA.md)
    └── data/transactions-seed.json        ← 50 transações fictícias, ~7-8 por PSP
```

## Pipeline (`POST /api/routing/recommend`)

```
① TransactionTextSerializer + EmbeddingService
    → "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA,
       regiao SP, hora 20h, categoria STREAMING."
    → vetor float[384] via all-MiniLM-L6-v2 (ONNX local, sem API)

② SimilaritySearchService → Neo4jTransactionRepository
    → busca as 5 transações históricas mais similares (cosine similarity)

③ ReasoningService
    → monta o prompt RAG (transação atual + casos similares)
    → chama o LLM via OpenRouter (RoutingPromptBuilder + ChatModel)
    → RecommendationParser decodifica { primary, confidence, reasoning, fallback[] }

④ RoutingService combina a recomendação do LLM com os similarCases encontrados
```

## Dados de seed

`data/transactions-seed.json` tem 50 transações fictícias, distribuídas pelos 7 PSPs
seguindo os padrões descritos em `IDEIA.md` (ex.: ADYEN com cartões internacionais de valor
alto em SP/RJ à noite; BRASPAG melhor em PIX/ELO em horário comercial; BRADESCO piorando após
22h; PICPAY/MERCADOPAGO exclusivos das respectivas carteiras; NUPAY em cartões Nubank de valor
médio; SANTANDER em clientes Santander em horário comercial).

---

## Pré-requisitos

- Java 17+
- Maven 3.8+
- Docker e Docker Compose (para o Neo4j)
- Uma API key da [OpenRouter](https://openrouter.ai/keys)

## Como rodar

```bash
cd backend

# 1. Suba o Neo4j
docker-compose up -d

# 2. Configure as variáveis de ambiente
cp .env.example .env
# edite .env com sua OPENROUTER_API_KEY

# 3. Rode o backend
mvn spring-boot:run

# 4. Carregue os dados históricos
curl -X POST http://localhost:8080/api/routing/seed

# 5. Peça uma recomendação
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
cd backend
mvn test
```

19 testes JUnit 5 + AssertJ + Mockito, sem depender de Neo4j, do modelo de embeddings ou de
uma chamada HTTP real ao OpenRouter — a lógica de cada etapa foi extraída para classes puras
e injetáveis, seguindo o mesmo padrão de `pdf-rag-knowledge-base-java`:

| Classe de teste | Cobre |
|---|---|
| `TransactionTextSerializerTest` | Serialização com/sem bandeira, arredondamento de valor |
| `RoutingPromptBuilderTest` | Placeholders substituídos, instruções fixas do template preservadas |
| `RecommendationParserTest` | JSON válido, JSON com cerca markdown, campos ausentes, PSP desconhecido |
| `ReasoningServiceTest` | Integração prompt → `ChatModel` fake → parsing |
| `RoutingServiceTest` | Orquestração dos 4 passos, combinação de `similarCases`, caso sem resultados |
| `RoutingControllerTest` | Contrato HTTP dos dois endpoints, validação de corpo (400), sucesso (200) |

## Build

```bash
cd backend
mvn compile   # BUILD SUCCESS
mvn test      # Tests run: 19, Failures: 0, Errors: 0
```

---

## Relação com o porte Go

Este projeto também tem uma versão em Go, [`../psp-routing-intelligence-go`](../psp-routing-intelligence-go),
portada **a partir desta implementação Java** (e não diretamente de `IDEIA.md`) — ver o
README do porte Go para o detalhamento de paridade e adaptações.

## Dependências principais

| Dependência | Versão | Uso |
|---|---|---|
| `spring-boot-starter-web` | 3.4.1 | API REST |
| `spring-boot-starter-validation` | 3.4.1 | Bean Validation do `RecommendRequest` |
| `langchain4j` | 1.12.2 | Framework principal |
| `langchain4j-open-ai` | 1.12.2 | Cliente HTTP OpenAI-compatible (OpenRouter) |
| `langchain4j-community-neo4j` | 1.12.2-beta22 | `Neo4jEmbeddingStore` |
| `langchain4j-embeddings-all-minilm-l6-v2` | 1.12.2-beta22 | Modelo de embeddings ONNX in-process |
| `neo4j-java-driver` | 5.28.0 | Limpeza de dados antes do (re)seed |
| `dotenv-java` | 3.0.0 | Leitura do arquivo `.env` |
