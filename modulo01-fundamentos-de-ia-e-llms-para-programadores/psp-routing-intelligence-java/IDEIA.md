# PSP Routing Intelligence

> Projeto Java com IA — Módulo 01: Fundamentos de IA e LLMs para Programadores
> Autor: Rafael Mattos Moreira | Software Engineer @ Globo | Payment Services

---

## O Problema

O gateway de pagamentos da Globo orquestra múltiplos PSPs (Adyen, Braspag, Bradesco, Santander, NuPay, PicPay, MercadoPago). Hoje, a decisão de qual PSP processar uma transação é baseada em **regras estáticas** (Strategy Pattern com critérios fixos).

Cada transação tem um conjunto rico de características — valor, método (PIX, cartão, wallet), bandeira, região do usuário, hora do dia, categoria do produto — e cada PSP tem padrões de aprovação distintos para esses perfis. Uma regra estática não consegue capturar toda essa complexidade.

**A IA pode aprender com o histórico e recomendar o PSP com maior chance de aprovação, explicando o raciocínio.**

---

## A Solução

Sistema que usa **Embeddings + RAG + LLM** para recomendar o PSP ideal para cada transação:

1. Transações históricas são representadas como vetores (embeddings) e armazenadas no Neo4j
2. Quando chega uma nova transação, busca-se as mais similares no histórico
3. Um LLM analisa os casos similares e recomenda o PSP com justificativa em PT-BR

### Exemplo de resposta:

```json
POST /api/routing/recommend
{
  "amount": 1500.00,
  "method": "CREDIT_CARD",
  "brand": "VISA",
  "userRegion": "SP",
  "hour": 20,
  "merchantCategory": "STREAMING"
}

Response:
{
  "primary": "ADYEN",
  "confidence": 0.87,
  "reasoning": "Transações VISA em SP entre 18h-22h têm 87% de aprovação no Adyen. Bradesco apresentou queda de 12% nesse perfil nos últimos 30 dias.",
  "fallback": ["BRASPAG", "BRADESCO"],
  "similarCases": [
    { "psp": "ADYEN", "status": "SUCCESS", "amount": 1320.00, "method": "CREDIT_CARD", "similarity": 0.94 },
    { "psp": "ADYEN", "status": "SUCCESS", "amount": 1780.00, "method": "CREDIT_CARD", "similarity": 0.91 },
    { "psp": "BRADESCO", "status": "FAILED", "amount": 1450.00, "method": "CREDIT_CARD", "similarity": 0.89 }
  ]
}
```

---

## Conceitos do Curso Aplicados

| Conceito | Onde é usado |
|---|---|
| **Embeddings locais (ONNX)** | Representar cada transação como vetor de 384 dimensões |
| **Neo4j Vector Search** | Armazenar e buscar transações similares por cosine similarity |
| **RAG** | Recuperar casos históricos similares para alimentar o LLM |
| **LangChain4j + OpenRouter** | LLM analisa os casos e gera recomendação em JSON |
| **Spring Boot + Clean Architecture** | Mesmo padrão usado na Globo |

---

## Stack Técnico

### Backend (Java 21 + Spring Boot 3.4)
- **LangChain4j 1.12.2** — mesma versão dos projetos do curso
- **all-MiniLM-L6-v2** (ONNX, local) — embeddings sem API key
- **Neo4j** — vector store para transações históricas
- **OpenRouter** — Gemma 3 / Llama 3 (free tier) para o raciocínio
- **Clean Architecture + DDD** — mesmo padrão do gateway da Globo
- **Docker Compose** — Neo4j containerizado

### Frontend (React + Vite)
- Formulário para inserir uma transação
- Card com PSP recomendado + barra de confiança
- Painel com raciocínio do LLM
- Tabela com os casos históricos similares encontrados

---

## Arquitetura

```
psp-routing-intelligence/
├── backend/                               ← Spring Boot 3.4 (Java 21)
│   ├── pom.xml
│   ├── docker-compose.yml                 ← Neo4j
│   ├── .env.example
│   └── src/main/java/com/psprouting/
│       ├── PspRoutingApplication.java
│       ├── domain/
│       │   ├── Transaction.java           ← record com features da transação
│       │   ├── PSP.java                   ← enum: ADYEN, BRASPAG, BRADESCO...
│       │   └── RoutingRecommendation.java ← resultado com primary + reasoning
│       ├── application/
│       │   ├── EmbeddingService.java      ← serializa + gera vetor da transação
│       │   ├── SimilaritySearchService.java ← busca no Neo4j
│       │   ├── ReasoningService.java      ← RAG + LLM via LangChain4j
│       │   └── RoutingService.java        ← orquestra o pipeline completo
│       ├── infrastructure/
│       │   ├── Neo4jTransactionRepository.java ← store e search no Neo4j
│       │   └── config/
│       │       ├── AppConfig.java         ← beans: EmbeddingModel, ChatModel, Neo4jStore
│       │       └── CorsConfig.java        ← permite chamadas do frontend
│       ├── controller/
│       │   └── RoutingController.java     ← POST /recommend, POST /seed
│       └── resources/
│           ├── application.properties
│           ├── prompts/routing-prompt.txt ← template do prompt RAG
│           └── data/transactions-seed.json ← 50 transações fictícias por PSP
│
└── frontend/                              ← React + Vite
    ├── package.json
    └── src/
        ├── App.jsx
        ├── components/
        │   ├── TransactionForm.jsx         ← inputs da transação
        │   ├── RecommendationCard.jsx      ← PSP + confidence bar
        │   ├── ReasoningPanel.jsx          ← explicação do LLM
        │   └── SimilarCasesTable.jsx       ← histórico similar
        └── services/api.js
```

---

## Fluxo do Pipeline

```
POST /api/routing/recommend
  Body: { amount, method, brand, userRegion, hour, merchantCategory }

  ① EmbeddingService
      → serializa a transação em texto natural:
        "Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA,
         região SP, hora 20h, categoria STREAMING."
      → gera float[384] via all-MiniLM-L6-v2 (ONNX local, sem API)

  ② Neo4jTransactionRepository
      → busca as 5 transações históricas mais similares (cosine similarity)
      → retorna: [{ psp, status, amount, method, similarity }]

  ③ ReasoningService
      → monta prompt RAG com a transação atual + casos similares
      → chama OpenRouter (Gemma 3 ou Llama 3 — free tier)
      → LLM responde: { primary, confidence, reasoning, fallback[] }

  ④ Response completa ao cliente
```

---

## Dados de Seed

50 transações fictícias representando padrões realistas por PSP:

```json
[
  { "amount": 299.90, "method": "PIX",         "brand": null,         "region": "SP", "hour": 14, "category": "STREAMING", "psp": "BRASPAG",     "status": "SUCCESS" },
  { "amount": 2800.00,"method": "CREDIT_CARD",  "brand": "MASTERCARD", "region": "RJ", "hour": 23, "category": "STREAMING", "psp": "ADYEN",       "status": "SUCCESS" },
  { "amount": 19.90,  "method": "WALLET",       "brand": "PICPAY",     "region": "MG", "hour": 8,  "category": "GAMING",    "psp": "PICPAY",      "status": "SUCCESS" },
  { "amount": 149.90, "method": "CREDIT_CARD",  "brand": "VISA",       "region": "SP", "hour": 3,  "category": "STREAMING", "psp": "BRADESCO",    "status": "FAILED"  },
  { "amount": 59.90,  "method": "PIX",          "brand": null,         "region": "BA", "hour": 10, "category": "STREAMING", "psp": "BRASPAG",     "status": "SUCCESS" },
  { "amount": 1200.00,"method": "CREDIT_CARD",  "brand": "AMEX",       "region": "SP", "hour": 19, "category": "STREAMING", "psp": "ADYEN",       "status": "SUCCESS" },
  { "amount": 9.90,   "method": "WALLET",       "brand": "MERCADOPAGO","region": "RJ", "hour": 15, "category": "NEWS",      "psp": "MERCADOPAGO", "status": "SUCCESS" },
  ...
]
```

**Padrões codificados nos dados de seed:**
- ADYEN: alto desempenho em cartões internacionais (VISA/MASTERCARD/AMEX) de valor alto, SP/RJ, noturno
- BRASPAG: melhor performance em PIX, ELO, qualquer região, horário comercial
- BRADESCO: fallback razoável para cartões nacionais, piora após 22h
- PICPAY: exclusivo para pagamentos via carteira PicPay
- MERCADOPAGO: exclusivo para carteira MercadoPago, bom em valores baixos
- NUPAY: competitivo em cartões Nubank para valores médios
- SANTANDER: especializado em clientes Santander, horário comercial

---

## Prompt RAG (routing-prompt.txt)

```
Você é um especialista em roteamento de pagamentos para plataformas de streaming e entretenimento digital.

Analise a transação a seguir e os casos históricos similares para recomendar o melhor PSP.

TRANSAÇÃO ATUAL:
{transaction}

CASOS HISTÓRICOS SIMILARES (ordenados por similaridade de perfil):
{similarCases}

PSPs disponíveis: ADYEN, BRASPAG, BRADESCO, SANTANDER, NUPAY, PICPAY, MERCADOPAGO

Responda EXCLUSIVAMENTE em JSON válido, sem markdown, sem texto adicional:
{
  "primary": "NOME_DO_PSP",
  "confidence": 0.0,
  "reasoning": "Explicação objetiva em português, máximo 2 frases.",
  "fallback": ["PSP2", "PSP3"]
}
```

---

## Reutilização dos Projetos do Curso

| Projeto do curso | O que reutilizar |
|---|---|
| `embeddings-vector-search/` | Padrão de EmbeddingModel + Neo4jEmbeddingStore (~80% reaproveitável) |
| `pdf-rag-knowledge-base/` | Padrão de RAG + OpenAiChatModel via OpenRouter |
| `openrouter-multi-model-chat/` | Configuração LangChain4j + OpenRouter |

---

## Como Rodar

```bash
# 1. Subir Neo4j
cd backend
docker-compose up -d

# 2. Configurar variáveis de ambiente
cp .env.example .env
# Editar .env com sua OPENROUTER_API_KEY

# 3. Rodar o backend
mvn spring-boot:run

# 4. Carregar dados históricos
curl -X POST http://localhost:8080/api/routing/seed

# 5. Testar uma recomendação
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

# 6. Rodar o frontend (outra aba do terminal)
cd ../frontend
npm install && npm run dev
# Abrir http://localhost:5173
```

---

## Variáveis de Ambiente (.env.example)

```env
OPENROUTER_API_KEY=sk-or-v1-...
NLP_MODEL=google/gemma-3-27b-it:free

NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
```

---

## Por Que Este Projeto é Relevante

1. **Aplicação real**: Resolve um problema que existe no gateway da Globo hoje
2. **Explainable AI**: O LLM justifica a decisão em linguagem natural — transparência para stakeholders
3. **Portfólio diferenciado**: Combina embeddings + RAG + LLM aplicados a pagamentos — raro no mercado
4. **Extensível**: Pode evoluir para aprender em tempo real com novas transações (feedback loop)
5. **Mesmo padrão de código**: Clean Architecture + DDD + Spring Boot — zero atrito para apresentar ao time
