# Ollama Local LLM Chat — Java + LangChain4j

Demonstracao de como usar um LLM local via **Ollama** com **Java 17** e **LangChain4j**.

O Ollama expoe uma API compativel com OpenAI em `http://localhost:11434/v1`, permitindo
usar o mesmo cliente `langchain4j-open-ai` que seria utilizado com a API da OpenAI — apenas
apontando o `baseUrl` para o servidor local.

Este projeto e o equivalente Java do script `exemplo-10-ollama/request.sh`.

---

## Arquitetura

```
.env  →  Main.java  →  OpenAiChatModel (LangChain4j)  →  Ollama API  →  LLM local
```

- **dotenv-java** le as variaveis `OLLAMA_BASE_URL` e `OLLAMA_MODEL` do arquivo `.env`
- **LangChain4j** (`langchain4j-open-ai`) faz as requisicoes HTTP para o Ollama
- O Ollama processa a inferencia localmente usando o modelo escolhido

---

## Pre-requisitos

### 1. Instalar o Ollama

**Linux / macOS:**
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

**Windows:**
Baixe o instalador em: https://ollama.com/download

### 2. Iniciar o servidor Ollama

```bash
ollama serve
```

O servidor ficara disponivel em `http://localhost:11434`.

### 3. Baixar um modelo

```bash
# Modelo recomendado (pequeno, rapido, ~2GB):
ollama pull llama3.2

# Alternativas:
ollama pull llama2-uncensored:7b
ollama pull mistral
ollama pull phi3

# Verificar modelos instalados:
ollama list
```

### 4. Instalar Java 17+ e Maven

Verifique as versoes instaladas:
```bash
java -version
mvn -version
```

---

## Configuracao

Copie o arquivo de exemplo e ajuste conforme necessario:

```bash
cp .env.example .env
```

Conteudo do `.env`:
```
OLLAMA_MODEL=llama3.2
OLLAMA_BASE_URL=http://localhost:11434/v1
```

Se o `.env` nao existir, os valores padrao serao usados automaticamente.

---

## Executar

```bash
# Compilar e executar com Maven:
mvn compile exec:java

# Ou: compilar, empacotar e executar o jar:
mvn package
java -jar target/ollama-local-llm-chat-1.0.0.jar
```

### Saida esperada

```
======================================================================
  Ollama Local LLM Chat - Java + LangChain4j
======================================================================

Configuracao:
  Base URL : http://localhost:11434/v1
  Modelo   : llama3.2

----------------------------------------------------------------------
[1/5] PERGUNTA:
Explique o que e inteligencia artificial em 3 frases simples.

RESPOSTA:
Inteligencia artificial (IA) e a capacidade de maquinas realizarem
tarefas que normalmente requerem inteligencia humana...

(tempo: 3421 ms)

...
```

---

## Testes

A logica de orquestracao (`ChatSession`) e de configuracao (`ChatConfig`) foi extraida de
`Main.java` para classes isoladas e testadas com JUnit 5 + Mockito + AssertJ — sem depender
de uma instancia real do Ollama, mesma abordagem usada no porte Go
(`ollama-local-llm-chat-go/ollama/client_test.go`, testado com `httptest`).

Cenarios cobertos (equivalentes ao pacote Go `ollama`):

- resposta com sucesso;
- falha em uma pergunta sem interromper as demais (recuperacao);
- falha persistente registrada como erro (sem derrubar a aplicacao);
- lista vazia de perguntas;
- uma chamada ao `ChatModel` por pergunta.

A retentativa de rede em si (`maxRetries`) e responsabilidade do `OpenAiChatModel` do
LangChain4j, configurado em `Main.java`.

```bash
mvn test
```

---

## Dependencias

| Dependencia | Versao | Uso |
|---|---|---|
| `langchain4j` | 1.12.2 | Framework principal |
| `langchain4j-open-ai` | 1.12.2 | Cliente HTTP OpenAI-compatible |
| `dotenv-java` | 3.0.0 | Leitura do arquivo `.env` |
| `slf4j-simple` | 2.0.16 | Logger minimalista |

---

## Relacao com o exemplo shell

O script `../exemplo-10-ollama/request.sh` faz:

```bash
curl -X POST http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "llama2-uncensored:7b", "messages": [{"role":"user","content":"..."}]}'
```

Este projeto Java faz exatamente o mesmo, porem usando LangChain4j como abstração
de alto nivel — sem necessidade de montar o JSON manualmente ou usar curl.

---

## Modelos Testados

| Modelo | Tamanho | Caracteristicas |
|---|---|---|
| `llama3.2` | ~2GB | Rapido, bom para conversas gerais |
| `llama2-uncensored:7b` | ~3.8GB | Sem filtros de conteudo |
| `mistral` | ~4GB | Bom equilibrio velocidade/qualidade |
| `phi3` | ~2GB | Microsoft, eficiente em hardware limitado |
| `gpt-oss:20b` | ~12GB | Maior, melhor qualidade, mais lento |
