# Portfolio Intelligence Advisor — CarteirAI

> Projeto Java com IA — Módulo 01: Fundamentos de IA e LLMs para Programadores
> Inspirado em: exemplo-01-ecommerce-recomendations-z + embeddings-vector-search + pdf-rag-knowledge-base

---

## O Problema

Planilhas de controle de carteira são ótimas para registrar dados, mas **estáticas na tomada de decisão**:
- A calculadora de aportes usa regras fixas (% atual vs % ideal)
- Não considera múltiplos fatores combinados (P/VP + DY + desvio + setor + timing)
- Não aprende com decisões passadas
- Não explica o raciocínio em linguagem natural
- Não responde perguntas de follow-up

---

## A Solução: 3 Camadas de Inteligência

```
┌─────────────────────────────────────────────────────────────────┐
│              PORTFOLIO INTELLIGENCE ADVISOR                     │
├──────────────────┬─────────────────────┬────────────────────────┤
│  CAMADA 1        │  CAMADA 2           │  CAMADA 3              │
│  Neural Network  │  Embeddings + RAG   │  LLM + Chatbot         │
│  "Cérebro Quant" │  "Memória"          │  "Comunicação"         │
│                  │                     │                        │
│  Analisa cada    │  Busca cenários     │  Explica a             │
│  ativo e gera    │  similares no       │  recomendação e        │
│  BUY SCORE [0-1] │  histórico (Neo4j)  │  responde perguntas    │
└──────────────────┴─────────────────────┴────────────────────────┘
```

---

## Analogia com o Projeto Ecommerce (exemplo-01)

| E-commerce | Portfolio Intelligence |
|---|---|
| **Usuário** (idade, histórico de compras) | **Estado da carteira** (alocação atual, desvios, capital disponível) |
| **Produto** (preço, categoria, cor) | **Ativo** (FII/Ação/Tesouro, setor, P/VP, DY) |
| **Comprou / Não comprou** → treina NN | **Aportar / Aguardar** → treina NN |
| **Score de recomendação** [0-1] por produto | **BUY PRIORITY SCORE** [0-1] por ativo |
| Features: one-hot categoria/cor + preço norm. | Features: one-hot setor/tipo + P/VP, DY, desvio norm. |
| Ordena produtos por score | Ordena ativos por prioridade de aporte |

---

## Camada 1 — Neural Network (Java puro, como ecommerce-neural-recommender)

### Features de entrada por ativo

```
Desvio do peso ideal (normalizado)    → ex: -0.067 (7% abaixo do ideal)
P/VP (normalizado)                    → ex: 0.94 (bom para FII)
Dividend Yield % (normalizado)        → ex: 0.12 (12% a.a.)
Preço vs Preço Médio de Compra        → ex: 0.95 (5% abaixo do PM)
Dias desde último aporte (normalizado)→ ex: 0.3 (aportou há 30 dias)
Setor (one-hot encoded):
  [1,0,0,0,0,0] = Logística
  [0,1,0,0,0,0] = Shoppings
  [0,0,1,0,0,0] = Papel
  [0,0,0,1,0,0] = Renda Urbana
  [0,0,0,0,1,0] = Agro
  [0,0,0,0,0,1] = Híbridos
Tipo de ativo (one-hot encoded):
  [1,0,0,0] = FII
  [0,1,0,0] = Ação
  [0,0,1,0] = Tesouro
  [0,0,0,1] = Cripto
```

### Arquitetura da rede

```
Input:  [desvio, P/VP, DY, preço_vs_PM, dias, setor_onehot, tipo_onehot]
                              ↓
Layer 1: 128 neurônios, ReLU   ← aprende combinações de features
Layer 2:  64 neurônios, ReLU   ← comprime padrões relevantes
Layer 3:  32 neurônios, ReLU   ← distila decisão
Layer 4:   1 neurônio, Sigmoid ← BUY PRIORITY SCORE [0.0 - 1.0]
```

### Treinamento

Dados sintéticos gerados a partir das regras de value investing:
```java
// Exemplos de treinamento:
// Ativo com desvio -10%, P/VP=0.88, DY=12%, setor=Logística → score 0.95 (COMPRAR!)
// Ativo com desvio +8%, P/VP=1.10, DY=6%,  setor=Papel     → score 0.12 (AGUARDAR)
// Ativo com desvio -3%, P/VP=0.95, DY=9%,  setor=Shoppings → score 0.67 (MONITORAR)
```

**Saída da NN para cada ativo:**
```json
{ "ticker": "HGLG11", "buyScore": 0.89, "action": "COMPRAR" }
{ "ticker": "XPML11", "buyScore": 0.76, "action": "COMPRAR" }
{ "ticker": "BTLG11", "buyScore": 0.31, "action": "MONITORAR" }
{ "ticker": "HGRU11", "buyScore": 0.08, "action": "AGUARDAR" }
```

---

## Camada 2 — Embeddings + Neo4j (LangChain4j)

### O que é armazenado no Neo4j

**Nó 1: Portfolio Snapshots (estados históricos)**
Cada "momento da carteira" é serializado como texto e vetorizado:
```
"Carteira com Logística 36%, Shoppings 16%, Papel 16%. Capital disponível R$5.000.
 HGLG11 abaixo do ideal (-7%), P/VP=0.94. XPML11 abaixo do ideal (-3%), P/VP=0.99.
 Resultado do aporte: +R$3.000 HGLG11, +R$2.000 XPML11. Sucesso: DY médio subiu 0.3%"
```
→ Embedding 384d → armazenado com metadata: data, aportes realizados, resultado

**Nó 2: Regras de Investimento (corpus de conhecimento)**
Princípios de Bazin, Graham, análise de FIIs:
```
"Critério Bazin: ativo com Dividend Yield acima de 6% ao ano e preço
 abaixo do valor patrimonial é candidato à compra em renda variável."

"Para FIIs de papel (CRI/CRA), vacância não se aplica. Avaliar qualidade
 dos devedores, duration da carteira e spread médio de crédito."

"FIIs de logística premium tendem a manter contratos atípicos de longo prazo.
 Vacância abaixo de 5% com ABL acima de 200.000m² indica portfólio resiliente."
```

### O que é buscado no momento da análise

Quando o usuário pede recomendação:
1. Serializa estado atual da carteira como texto
2. Gera embedding → busca Neo4j
3. Retorna: **top-3 estados similares** + que ação foi tomada e qual foi o resultado
4. Retorna: **regras relevantes** para os ativos com maior buy score

---

## Camada 3 — LLM + Chatbot (LangChain4j + OpenRouter)

### Prompt RAG para recomendação

```
Você é um consultor de investimentos especializado em value investing e FIIs.

ESTADO ATUAL DA CARTEIRA:
{portfolio_state}

CAPITAL DISPONÍVEL: R$ {available_capital}

ANÁLISE DA REDE NEURAL (scores de prioridade por ativo):
{neural_network_scores}

CENÁRIOS HISTÓRICOS SIMILARES ENCONTRADOS:
{similar_cases_from_neo4j}

REGRAS DE INVESTIMENTO RELEVANTES:
{investment_rules_from_neo4j}

Com base em TODOS os dados acima, gere:
1. Plano de aporte detalhado com valores por ativo
2. Justificativa para cada recomendação (máximo 2 frases por ativo)
3. Alerta de riscos se houver
4. Percentual de exposição após o aporte

Responda em JSON estruturado + um parágrafo de análise geral em PT-BR.
```

### Chatbot de follow-up

Após a recomendação, o usuário pode perguntar:
- *"Por que HGLG11 e não BTLG11? Os dois são logística."*
- *"Se eu tiver só R$2.000, como ficaria?"*
- *"Qual o risco de aportar tudo em FIIs de papel agora?"*
- *"Quando devo revisar essa recomendação?"*

O chatbot mantém o contexto da sessão (recomendação + estado da carteira) e responde com base no RAG + LLM.

---

## Fluxo Completo do Sistema

```
POST /api/portfolio/analyze
  Body: {
    available_capital: 5000,
    assets: [
      { ticker: "HGLG11", sector: "LOGISTICA", type: "FII",
        current_weight: 0.1278, ideal_weight: 0.15,
        position: 255, avg_price: 157.17, current_price: 155.60,
        dy: 0.089, pvp: 0.94 },
      ...
    ]
  }

  ① Neural Network (Java)
      → Normaliza features de cada ativo
      → Prediz BUY SCORE [0-1] para cada um
      → Ordena por prioridade

  ② EmbeddingService (LangChain4j + ONNX)
      → Serializa estado da carteira em texto
      → Gera embedding all-MiniLM-L6-v2

  ③ Neo4j Search
      → Busca top-3 portfolio snapshots similares
      → Busca regras relevantes para os top assets

  ④ LLM (OpenRouter via LangChain4j)
      → Monta prompt RAG com tudo acima
      → Gera recomendação estruturada em JSON + análise

  ⑤ Armazena sessão (recomendação + contexto)
      → Libera chatbot para follow-up

  Response: {
    allocations: [
      { ticker: "HGLG11", buyScore: 0.89, action: "COMPRAR",
        suggestedAmount: 3120, units: 2, reasoning: "..." },
      { ticker: "XPML11", buyScore: 0.76, action: "COMPRAR",
        suggestedAmount: 1880, units: 1, reasoning: "..." },
      { ticker: "BTLG11", buyScore: 0.31, action: "MONITORAR",
        suggestedAmount: 0, reasoning: "..." }
    ],
    portfolioAfterAport: { logistica: "38.1%", shoppings: "18.2%", ... },
    generalAnalysis: "...",
    sessionId: "abc-123"
  }

POST /api/chat/{sessionId}
  Body: { message: "Por que não BTLG11?" }
  → LLM responde com contexto da recomendação
```

---

## Stack Técnico

### Backend (Java 21 + Spring Boot 3.4)
```
├── Neural Network
│   └── Java puro (como ecommerce-neural-recommender)
│       Sem framework ML externo — implementação do zero
│       Backpropagation + Adam optimizer em Java
│
├── Embeddings + Vector DB
│   ├── LangChain4j + all-MiniLM-L6-v2 (ONNX local)
│   └── Neo4j — portfolio snapshots + regras de investimento
│
├── LLM + Chatbot
│   ├── LangChain4j + OpenRouter (Gemma 3 / Llama 3)
│   └── Session management para contexto do chatbot
│
└── API REST (Spring Boot 3.4)
    ├── POST /api/portfolio/analyze
    ├── POST /api/portfolio/seed         ← carrega regras de investimento no Neo4j
    ├── POST /api/chat/{sessionId}
    └── GET  /api/chat/{sessionId}/history
```

### Frontend (React + Vite)
```
├── PortfolioInputForm        ← input dos ativos + capital disponível
├── RecommendationDashboard   ← lista de ativos com buyScore + ação
│   ├── AssetCard             ← ticker, score visual, valor sugerido
│   ├── AllocationChart       ← gráfico antes vs depois do aporte
│   └── AnalysisPanel         ← análise geral do LLM
└── ChatInterface             ← chatbot de follow-up com a IA
```

---

## Estrutura de Arquivos

```
portfolio-intelligence-advisor/
├── backend/
│   ├── pom.xml
│   ├── docker-compose.yml                    ← Neo4j
│   ├── .env.example
│   └── src/main/java/com/portfolioai/
│       ├── PortfolioAdvisorApplication.java
│       ├── domain/
│       │   ├── Asset.java                    ← features de cada ativo
│       │   ├── Portfolio.java                ← estado da carteira
│       │   ├── AssetRecommendation.java      ← score + ação + valor sugerido
│       │   └── PortfolioRecommendation.java  ← recomendação completa
│       ├── neuralnetwork/
│       │   ├── NeuralNetwork.java            ← rede neural Java puro
│       │   ├── AssetFeatureEncoder.java      ← normalização + one-hot
│       │   └── RebalancingModel.java         ← treina + prediz
│       ├── application/
│       │   ├── PortfolioAnalysisService.java ← orquestra as 3 camadas
│       │   ├── EmbeddingService.java         ← LangChain4j ONNX
│       │   ├── PortfolioSnapshotRepository.java ← Neo4j search
│       │   ├── ReasoningService.java         ← RAG + LLM
│       │   └── ChatService.java              ← chatbot com sessão
│       ├── infrastructure/
│       │   ├── Neo4jRepository.java
│       │   └── config/
│       │       ├── AppConfig.java            ← beans NN, LLM, Neo4j
│       │       └── CorsConfig.java
│       ├── controller/
│       │   ├── PortfolioController.java
│       │   └── ChatController.java
│       └── resources/
│           ├── application.properties
│           ├── prompts/
│           │   ├── rebalancing-prompt.txt
│           │   └── chatbot-system-prompt.txt
│           └── data/
│               ├── training-scenarios.json   ← 200 cenários para treinar NN
│               └── investment-rules.json     ← regras Bazin/Graham/FIIs
│
└── frontend/
    └── src/
        ├── App.jsx
        ├── components/
        │   ├── PortfolioInputForm.jsx
        │   ├── RecommendationDashboard.jsx
        │   ├── AssetCard.jsx
        │   ├── AllocationChart.jsx
        │   └── ChatInterface.jsx
        └── services/api.js
```

---

## Dados de Treinamento da Neural Network (training-scenarios.json)

200 cenários sintéticos baseados em regras de value investing:

```json
[
  {
    "features": {
      "desvio_ideal": -0.10,
      "pvp": 0.88,
      "dy": 0.12,
      "preco_vs_pm": 0.97,
      "dias_ultimo_aporte": 45,
      "setor": "LOGISTICA",
      "tipo": "FII"
    },
    "label": 0.95
  },
  {
    "features": {
      "desvio_ideal": 0.08,
      "pvp": 1.15,
      "dy": 0.055,
      "preco_vs_pm": 1.05,
      "dias_ultimo_aporte": 5,
      "setor": "PAPEL",
      "tipo": "FII"
    },
    "label": 0.05
  },
  ...
]
```

**Regras codificadas nos dados:**
- Desvio negativo alto + P/VP < 1.0 + DY > 8% → score alto
- Desvio positivo + aporte recente → score baixo
- P/VP > 1.10 + DY < 6% (abaixo Bazin) → score muito baixo
- Setor abaixo do ideal → peso extra no score

---

## Corpus de Conhecimento — Neo4j (investment-rules.json)

```json
[
  "Critério Bazin: ativo com Dividend Yield superior a 6% ao ano é candidato à compra.",
  "Critério Graham: P/VP abaixo de 1.0 indica ativo negociado com desconto ao valor patrimonial.",
  "FIIs de logística com contratos atípicos de longo prazo oferecem maior previsibilidade de receita.",
  "Ao rebalancear, priorize ativos com maior desvio negativo do peso ideal combinado com bom DY.",
  "FIIs de papel (CRI/CRA) são mais sensíveis a juros. Verificar duration e qualidade dos devedores.",
  "Não concentrar mais de 40% da carteira de FIIs em um único setor.",
  "FIIs de shoppings performam melhor em períodos de crescimento do consumo e baixa vacância.",
  ...
]
```

---

## Por Que Este Projeto é Especial

| Aspecto | Detalhe |
|---|---|
| **NN aplicada a finanças** | Vai além de regras fixas — aprende padrões complexos de rebalanceamento |
| **RAG com memória histórica** | O sistema "lembra" de decisões passadas similares e seus resultados |
| **Chatbot contextual** | A IA explica cada decisão e responde dúvidas com contexto completo |
| **3 camadas complementares** | NN (quant) + Embeddings (memória) + LLM (comunicação) — cada uma resolve um problema diferente |
| **Extensível** | Pode evoluir para aprender com feedback real: "aporte foi bom?" → atualiza Neo4j |
| **Portfólio diferenciado** | Combina todos os conceitos do módulo em um único sistema coeso |
