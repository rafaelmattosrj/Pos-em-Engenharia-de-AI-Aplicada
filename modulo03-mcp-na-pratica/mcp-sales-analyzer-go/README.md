# MCP Sales Analyzer em Go

Porte em Go do projeto Spring Boot [`mcp-sales-analyzer-java`](../mcp-sales-analyzer-java) — agente de análise de vendas com pipeline `IntentNode → ExecutorNode`, usando a OpenAI diretamente (não OpenRouter) com **function calling** para a tool local `csvToJson`.

## Fluxo

```
POST /analyze {question, data}
  → IntentNode: extrai intenção estruturada (dataType, parsedData, suggestedTools) via prompt JSON
  → ExecutorNode: chama o LLM com a tool csvToJson disponível (function calling),
                  executa o loop de tool-calling até obter o relatório final
  → AnalysisResult {report, toolsUsed, processingSteps}
```

## O que foi mantido 1:1

- Mesmo endpoint `POST /analyze`, mesmo contrato de request (`question`, `data`) e response (`report`, `toolsUsed`, `processingSteps`).
- Mesmo prompt de extração de intenção (JSON estruturado com `dataType`/`parsedData`/`question`/`suggestedTools`) e a mesma detecção heurística de tipo de dado (`[`/`{` → json, senão → csv).
- Mesmo fallback seguro do `IntentNode`: se o LLM falhar ou a resposta não for JSON válido, retorna os dados originais com o tipo detectado heuristicamente.
- Mesmo prompt de execução do `ExecutorNode`, e mesma tool local `csvToJson` (mesma descrição, mesmo schema), com parsing tolerante a cercas markdown (` ```json ` ) na resposta do `IntentNode`.
- Mesma lógica de `toolsUsed`: as tools sugeridas pelo `IntentNode`, mais `csvToJson` se o tipo de dado for `csv` e ainda não estiver na lista.
- `CsvToJsonTool`: mesmo comportamento para CSV vazio/em branco (`"[]"`), header + linhas, e campos entre aspas contendo vírgulas — portado para `csvtool/convert.go` usando `encoding/csv` da stdlib (equivalente ao Jackson CSV).

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Framework web | Spring Boot / `@RestController` | `net/http` com `http.ServeMux` |
| LLM + function calling | Spring AI `ChatClient` com `.defaultTools(csvToJsonTool)` (execução automática do loop de tool-calling) | Cliente OpenAI HTTP próprio (`openai/client.go`) com o loop de function calling implementado manualmente (`ChatWithTools`) — até 5 rounds, mesma semântica observável |
| MCP Client | `spring-ai-mcp-client-spring-boot-starter` com `List<ToolCallbackProvider>` (injetado automaticamente se houver servidor MCP configurado) | Não implementado — **igual à versão Java**, nenhum servidor MCP está configurado em `application.properties`/`.env`, então essa lista está sempre vazia na prática. Fica documentado aqui como ponto de extensão, não como funcionalidade ausente. |
| Provedor do LLM | OpenAI direto (`spring.ai.openai.*`) | OpenAI direto (`OPENAI_API_KEY`, `OPENAI_MODEL`, default `gpt-4o-mini`) — mesmo provedor, diferente dos outros portes deste repositório que usam OpenRouter |

## Como executar

```bash
cp .env.example .env   # preencha OPENAI_API_KEY
go run .
```

```bash
curl -X POST http://localhost:8080/analyze -d '{
  "question": "Qual produto teve maior receita?",
  "data": "produto,quantidade,preco\nNotebook,10,2500.00\nMouse,50,45.00"
}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem `csvtool.Convert` (header + linhas, CSV vazio/em branco, campos com vírgulas entre aspas — replicando os 3 cenários de `CsvToJsonToolTest.java`), o cliente OpenAI (chat simples, function calling sem tool call, function calling com execução de tool), `IntentNode` (JSON puro, cercado em markdown, fallback heurístico) e `ExecutorNode`/handler ponta-a-ponta — todos via `httptest`, sem depender da API real da OpenAI.
