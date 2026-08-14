# PDF RAG Knowledge Base — Java + LangChain4j

Pipeline RAG (Retrieval-Augmented Generation) completo em Java: **PDF → chunks → embeddings →
Neo4j → busca por similaridade → resposta gerada por LLM via OpenRouter**.

Este projeto é o porte Java do exemplo original em TypeScript
[`exemplo-13-embeddings-neo4j-rag`](../exemplo-13-embeddings-neo4j-rag), usando
[LangChain4j](https://docs.langchain4j.dev) para orquestrar todas as etapas.

---

## Arquitetura

```
tensores.pdf → ApachePdfBoxDocumentParser → DocumentSplitters.recursive(1000, 200)
             → AllMiniLmL6V2EmbeddingModel (ONNX, in-process)
             → Neo4jEmbeddingStore (label "Chunk", índice "tensors_index", cosseno)
             → busca por similaridade (top 3, score > 0.5)
             → PromptBuilder (template fixo em pt-BR)
             → OpenAiChatModel apontando para https://openrouter.ai/api/v1
```

O pipeline roda em 6 etapas, todas em `Main.java`:

1. Carrega e divide o PDF em chunks (1000 caracteres, overlap de 200).
2. Carrega o modelo de embeddings local `all-MiniLM-L6-v2` (ONNX, roda no próprio processo,
   sem depender de um servidor externo).
3. Conecta ao Neo4j e remove dados de execuções anteriores (`Chunk`, índice `tensors_index`).
4. Gera o embedding de cada chunk e o armazena no Neo4j.
5. Configura o LLM via OpenRouter (`OpenAiChatModel`, `temperature=0.3`, `maxRetries=2`).
6. Para cada pergunta de demonstração: busca os 3 trechos mais similares, filtra por
   score mínimo (`> 0.5`), monta o prompt e gera a resposta com o LLM.

A lógica de negócio de cada etapa 6 (filtro de relevância, montagem do prompt, tratamento de
erro do LLM) foi extraída para classes isoladas e testáveis:

| Classe | Responsabilidade |
|---|---|
| `RagContextBuilder` | Filtra trechos por score mínimo (`> 0.5`) e junta o contexto (`\n\n---\n\n`) |
| `PromptBuilder` | Monta o prompt final a partir do template fixo, pergunta e contexto |
| `ErrorMessages` | Trunca mensagens de erro do LLM em 200 caracteres (com segurança contra mensagem nula) |
| `RagAnswerService` | Orquestra busca → contexto → prompt → chamada ao LLM para uma pergunta |

---

## Pré-requisitos

- Java 17+
- Maven 3.8+
- Docker e Docker Compose (para o Neo4j)
- Uma API key da [OpenRouter](https://openrouter.ai/keys)

---

## Configuração

1. Copie o arquivo de exemplo de variáveis de ambiente:

   ```bash
   cp .env.example .env
   ```

2. Edite o `.env` e preencha sua chave da OpenRouter:

   ```env
   NEO4J_URI=bolt://localhost:7687
   NEO4J_USER=neo4j
   NEO4J_PASSWORD=password

   OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   NLP_MODEL=google/gemma-3-27b-it:free
   ```

3. Suba o Neo4j:

   ```bash
   docker-compose up -d
   ```

   O Neo4j Browser fica disponível em `http://localhost:7474` (usuário/senha padrão:
   `neo4j`/`password`).

> **Nunca commite o arquivo `.env`** — ele já está no `.gitignore` do repositório. Se uma
> chave de API real chegou a ser exposta em algum momento, revogue-a e gere uma nova em
> https://openrouter.ai/keys.

---

## Como executar

```bash
mvn compile exec:java
```

O programa processa `tensores.pdf`, popula o Neo4j e, em seguida, executa 5 perguntas de
demonstração sobre TensorFlow.js/machine learning, imprimindo a resposta gerada pelo LLM
para cada uma.

---

## Testes

```bash
mvn test
```

Como o pipeline depende de Neo4j, do parser de PDF e de uma chamada HTTP real ao OpenRouter,
a lógica de negócio de cada etapa foi extraída para classes puras/injetáveis e testada com
JUnit 5 + AssertJ, sem exigir nenhuma dessas dependências externas:

- `RagContextBuilderTest` — filtro por score mínimo (`> 0.5`), junção de múltiplos trechos,
  lista vazia, score no limite exato.
- `PromptBuilderTest` — o prompt final contém a pergunta, o contexto e as instruções fixas
  do template.
- `ErrorMessagesTest` — truncamento em 200 caracteres, mensagem dentro do limite, mensagem
  nula (evita `NullPointerException`).
- `RagAnswerServiceTest` — sem resultados de busca, sem contexto relevante, sucesso na
  chamada ao LLM, erro da API propagado e truncado — cenários equivalentes aos testados no
  cliente OpenRouter do porte Go irmão
  (`pdf-rag-knowledge-base-go/openrouter/client_test.go`).

---

## Relação com o porte Go

Este projeto também tem uma versão em Go: [`pdf-rag-knowledge-base-go`](../pdf-rag-knowledge-base-go).
O mesmo fluxo de 6 etapas, chunking, schema do Neo4j, prompt template e perguntas de
demonstração foram mantidos 1:1. Como o ecossistema Go não tem uma abstração tão madura
quanto o LangChain4j, algumas peças foram adaptadas para clientes HTTP próprios:

| Aspecto | Java (este projeto) | Go |
|---|---|---|
| Modelo de embeddings | `AllMiniLmL6V2EmbeddingModel` (ONNX in-process) | Ollama local servindo `all-minilm` via HTTP |
| Parser de PDF | Apache PDFBox | `github.com/ledongthuc/pdf` |
| Splitter recursivo | `DocumentSplitters.recursive` (LangChain4j) | Implementação própria |
| Neo4j driver | `neo4j-java-driver` + `Neo4jEmbeddingStore` | `neo4j-go-driver/v5` com Cypher manual |
| LLM (OpenRouter) | `OpenAiChatModel` (LangChain4j) | Cliente HTTP próprio, com retry manual |

---

## Dependências principais

| Dependência | Versão | Uso |
|---|---|---|
| `langchain4j` | 1.12.2 | Framework principal |
| `langchain4j-open-ai` | 1.12.2 | Cliente HTTP OpenAI-compatible (OpenRouter) |
| `langchain4j-community-neo4j` | 1.12.2-beta22 | `Neo4jEmbeddingStore` |
| `langchain4j-embeddings-all-minilm-l6-v2` | 1.12.2-beta22 | Modelo de embeddings ONNX in-process |
| `langchain4j-document-parser-apache-pdfbox` | 1.12.2-beta22 | Parser de PDF |
| `dotenv-java` | 3.0.0 | Leitura do arquivo `.env` |
| `slf4j-simple` | 2.0.16 | Logger minimalista |
| `junit-jupiter` | 5.10.2 | Testes |
| `assertj-core` | 3.25.3 | Asserções fluentes nos testes |

---

## Estrutura do projeto

```
pdf-rag-knowledge-base-java/
├── src/
│   ├── main/java/com/example/
│   │   ├── Main.java               # Pipeline RAG completo (6 etapas)
│   │   ├── RagContextBuilder.java  # Filtro de relevância + montagem do contexto
│   │   ├── PromptBuilder.java      # Template do prompt
│   │   ├── ErrorMessages.java      # Truncamento de mensagens de erro
│   │   └── RagAnswerService.java   # Orquestração busca → contexto → prompt → LLM
│   └── test/java/com/example/      # Testes JUnit 5 + AssertJ
├── tensores.pdf                    # Documento de origem usado no pipeline
├── docker-compose.yml              # Neo4j
├── .env.example
└── pom.xml
```
