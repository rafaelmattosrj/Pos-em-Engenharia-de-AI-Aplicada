# RAG com Neo4j Grafos em Go

Porte em Go do projeto Spring Boot [`rag-neo4j-grafos-java`](../rag-neo4j-grafos-java) — RAG sobre um knowledge graph Neo4j: converte a pergunta em Cypher via LLM, executa contra o banco com autocorreção em caso de erro de sintaxe, e gera a resposta final em português a partir dos resultados.

Ambos foram portados do projeto original em TypeScript/LangGraph [`06-rag-neo4j-students-z`](../06-rag-neo4j-students-z) (mesmo módulo).

## Fluxo

```
POST /chat → GetSchema → GenerateCypher (LLM) → Query (Neo4j)
                              ↑                      │ erro de sintaxe?
                              └── CorrectCypher (LLM) ┘ até MaxCorrectionAttempts
                                                       │
                                              GenerateResponse (LLM) → resposta
```

## O que foi mantido 1:1

- Mesmos dois endpoints: `POST /chat` (`{"question": "..."}` → `{"answer": "..."}`, com a mesma validação de tamanho mínimo de 5 caracteres) e `POST /seed` (repovoa o grafo com os mesmos 5 cursos, 3 alunos e 7 matrículas).
- Mesmo pipeline de self-correction: se a query Cypher gerada falhar na execução, pede uma correção ao LLM e tenta de novo, até `MaxCorrectionAttempts` (default 1); esgotadas as tentativas, retorna lista vazia em vez de propagar o erro.
- Mesma obtenção de schema via `CALL apoc.meta.schema()`, com fallback para `CALL db.labels()` + `CALL db.relationshipTypes()` quando APOC não está disponível.
- Mesmos prompts (geração de Cypher, correção de Cypher, resposta analítica) e mesma limpeza de cercas markdown (` ```cypher `/` ``` `) na resposta do LLM.
- Mesmo cliente resiliente com lista de modelos de fallback do OpenRouter, tentados em ordem até um responder.
- Mesmos defaults: porta 4000, `temperature=0.2`, `max_tokens=500`, `max-correction-attempts=1`, mesma lista de 5 modelos de fallback.

## O que foi adaptado

- Sem Spring Boot: `net/http`/`http.ServeMux` no lugar de `@RestController`; sem Spring Data Neo4j — usa o driver oficial `neo4j-go-driver/v5` diretamente (mesmo padrão já usado em `embeddings-vector-search-go`).
- Sem Spring AI: parse de resposta do LLM e chamada HTTP ao OpenRouter feitos à mão (`openrouter.Client` + `llm.ResilientClient`, mesmos pacotes já usados em `agendamento-medico-go`).
- `RagOrchestrator`, `CypherGeneratorService`, `AnalyticalResponseService` e `Neo4jService` viraram interfaces (`graph.Neo4jService`, `graph.CypherGenerator`, `graph.AnalyticalResponder`) para permitir testar o self-correction com fakes, sem precisar de um Neo4j real rodando — o Java não tinha testes automatizados para esse fluxo.
- `Neo4jService.getSchema`/`describeSchema` retornam a representação Go (`%v` dos records) em vez do `.toString()` da lista Java — texto equivalente, formatação nativa da linguagem.

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY e as credenciais do Neo4j
docker compose -f ../06-rag-neo4j-students-z/docker-compose.yaml up -d   # sobe o Neo4j do projeto original
go run .
```

```bash
curl -X POST http://localhost:4000/seed
curl -X POST http://localhost:4000/chat -d '{"question":"quais alunos estao matriculados no curso de React?"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem o `Orchestrator` (sucesso de primeira, autocorreção após falha de sintaxe, esgotamento de tentativas retornando lista vazia, e propagação de erro quando a própria geração de Cypher falha) via fakes de `Neo4jService`/`CypherGenerator`/`AnalyticalResponder` — não dependem de um Neo4j nem de uma chave de API real.
