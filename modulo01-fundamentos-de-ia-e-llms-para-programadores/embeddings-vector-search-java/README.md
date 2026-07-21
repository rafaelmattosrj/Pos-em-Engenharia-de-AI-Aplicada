# Embeddings + Vector Search com Neo4j (Java)

Demonstracao de como gerar embeddings localmente e executar busca por similaridade em um banco de dados vetorial Neo4j — **sem nenhuma chamada a LLM**.

Este projeto e o equivalente Java do exemplo TypeScript `exemplo-12-embeddings-neo4j`.

---

## Conceitos fundamentais

### O que sao Embeddings?

Embeddings sao representacoes numericas de texto na forma de vetores de numeros reais. Textos com significado semantico semelhante ficam proximos no espaco vetorial, independentemente de compartilharem as mesmas palavras.

Exemplo:
- "cachorro" e "cao" ficam proximos no espaco de embeddings
- "cachorro" e "banco de dados" ficam distantes

### Modelo all-MiniLM-L6-v2

O modelo `all-MiniLM-L6-v2` e um sentence transformer compacto e eficiente:

- Produz vetores de **384 dimensoes**
- Treinado para capturar similaridade semantica entre sentencas
- Roda **totalmente local** via ONNX Runtime — sem API, sem internet, sem custo
- Distribuido pelo artefato Maven `langchain4j-embeddings-all-minilm-l6-v2`

### Neo4j como Vector Store

O Neo4j armazena cada chunk de texto como um no (`Chunk`) com:
- O texto original como propriedade
- O vetor de 384 dimensoes como propriedade de embedding
- Um indice vetorial (`tensors_index`) que permite busca ANN (Approximate Nearest Neighbor)

A busca por similaridade usa **similaridade de cosseno** para encontrar os chunks cujos vetores sao mais proximos ao vetor da pergunta.

### Por que sem LLM?

Este exemplo foca exclusivamente na etapa de **recuperacao**: dado uma pergunta, quais trechos do documento sao mais relevantes? Isso e util para:
- Entender como o retrieval funciona isoladamente
- Avaliar a qualidade dos embeddings
- Depurar antes de adicionar geracao de respostas (RAG completo)

---

## Estrutura do projeto

```
embeddings-vector-search/
├── src/main/java/com/embeddings/
│   └── Main.java          # Logica principal
├── pom.xml                # Dependencias Maven
├── docker-compose.yml     # Neo4j via Docker
├── .env.example           # Template de variaveis de ambiente
├── tensores.pdf           # PDF de entrada (copiar manualmente — ver abaixo)
└── README.md
```

---

## Pre-requisitos

- Java 17+
- Maven 3.8+
- Docker e Docker Compose

---

## Configuracao

### 1. Copiar o PDF

O arquivo `tensores.pdf` precisa estar na raiz do projeto. Copie-o a partir do diretorio vizinho:

```bash
cp ../pdf-rag-knowledge-base/tensores.pdf ./tensores.pdf
```

### 2. Configurar variaveis de ambiente

```bash
cp .env.example .env
```

Edite `.env` se necessario (as credenciais do Neo4j devem bater com o `docker-compose.yml`):

```env
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
```

### 3. Subir o Neo4j

```bash
docker-compose up -d
```

Aguarde alguns segundos para o Neo4j inicializar completamente. Voce pode verificar em http://localhost:7474.

---

## Como executar

```bash
mvn compile exec:java
```

O Maven vai:
1. Compilar o projeto
2. Baixar o modelo ONNX (all-MiniLM-L6-v2) automaticamente na primeira execucao
3. Carregar e dividir `tensores.pdf` em chunks
4. Gerar embeddings localmente para cada chunk
5. Armazenar os vetores no Neo4j
6. Executar 5 buscas por similaridade e exibir os top-3 resultados com scores

---

## O que o programa faz (passo a passo)

| Etapa | Descricao |
|-------|-----------|
| 1 | Carrega `tensores.pdf` com Apache PDFBox e divide em chunks de 1000 caracteres com overlap de 200 |
| 2 | Inicializa `AllMiniLmL6V2EmbeddingModel` — modelo ONNX local, sem API |
| 3 | Conecta ao Neo4j e remove chunks da execucao anterior |
| 4 | Para cada chunk: gera embedding (vetor float[384]) e armazena no Neo4j |
| 5 | Para cada pergunta: gera embedding, busca os 3 chunks mais proximos por similaridade de cosseno e exibe texto + score |

Nenhuma chamada a LLM e feita — o output sao apenas os trechos brutos recuperados.

---

## Saida esperada

```
================================================================================
  Embeddings + Vector Search com Neo4j  (Java / LangChain4j)
  Modelo local: all-MiniLM-L6-v2  |  Sem chamada a LLM
================================================================================

ETAPA 1: Carregando PDF (tensores.pdf)...
  PDF carregado com sucesso.
  Dividido em 42 chunks (tamanho=1000, overlap=200)

ETAPA 2: Carregando modelo de embeddings local (all-MiniLM-L6-v2 ONNX)...
  Modelo carregado. Dimensao do vetor: 384

...

ETAPA 5: Executando buscas por similaridade...

================================================================================
PERGUNTA: O que sao tensores e como sao representados em JavaScript?
================================================================================
  Encontrados 3 resultados:

  Resultado #1  |  Score: 0.8921
  ------------------------------------------------------------
  Um tensor e uma estrutura de dados generalizada que representa...

  Resultado #2  |  Score: 0.8514
  ...
```

---

## Dependencias principais

| Artefato | Versao | Funcao |
|----------|--------|--------|
| `langchain4j` | 1.12.2 | Core: splitter, tipos de dados |
| `langchain4j-community-neo4j` | 1.12.2-beta22 | Neo4j vector store |
| `langchain4j-embeddings-all-minilm-l6-v2` | 1.12.2-beta22 | Modelo ONNX local |
| `langchain4j-document-parser-apache-pdfbox` | 1.12.2-beta22 | Parser de PDF |
| `neo4j-java-driver` | 5.28.0 | Driver Bolt para limpeza manual de dados |
| `dotenv-java` | 3.0.0 | Carregamento do `.env` |

---

## Relacao com o exemplo TypeScript (exemplo-12)

| Aspecto | TypeScript (exemplo-12) | Java (este projeto) |
|---------|------------------------|---------------------|
| Modelo de embedding | HuggingFace Transformers JS | AllMiniLmL6V2EmbeddingModel (ONNX) |
| Vector store | Neo4jVectorStore (@langchain/community) | Neo4jEmbeddingStore (LangChain4j) |
| Parser de PDF | PDFLoader (langchain) | ApachePdfBoxDocumentParser |
| LLM | Nenhum | Nenhum |
| Execucao do modelo | In-process (WASM/WebGPU) | In-process (ONNX Runtime JVM) |
