# Agent Evals em Go

Porte em Go do projeto Spring Boot [`agent-evals-java`](../agent-evals-java) — framework de avaliação de agentes: benchmark comparativo entre arquiteturas cognitivas, avaliação de seleção de ferramentas e avaliação de impacto de memória.

Equivalente a `benchmark_runner.py` (aula 09), `tool_selection_eval.py` (aula 12) e `memory_eval.py` (aula 15) do curso Python.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| POST | `/evals/benchmark` | Benchmark comparativo simulado das 3 arquiteturas cognitivas (ReAct, Plan-Execute, Reflection) |
| POST | `/evals/tool-selection` | Avalia seleção de ferramentas via LLM sobre 5 cenários fixos; retorna **422** se não atingir o threshold mínimo |
| POST | `/evals/memory-impact` | Compara métricas simuladas de execução com/sem memória |
| GET | `/evals/reports` | Lista os relatórios de benchmark salvos em disco |

## O que foi mantido 1:1

- Mesmo dataset de 5 cenários (`eval-dataset.json`), embarcado no binário via `go:embed` (equivalente ao `ClassPathResource` que empacota o arquivo dentro do jar — sempre disponível, independente do diretório de trabalho).
- Mesmos 5 casos de teste de `ToolSelectionEvaluator` (mesmo contexto, ferramenta esperada, argumentos esperados e ferramentas proibidas por caso) e mesmo critério de aprovação (`accuracy >= 0.80` E `unnecessaryRate <= 0.10`, ambos configuráveis).
- Mesma simulação de `BenchmarkRunner` (variação de ±5-10% sobre valores-base por arquitetura) e mesmo veredicto comparativo (`melhor_conclusao`, `mais_eficiente`, `menor_custo`, `melhor_cobertura_ferramentas`, `recomendado_geral` com score ponderado 40/30/30).
- Mesma simulação de `MemoryEvaluator` (6 métricas com base + variância, seed fixo).
- Mesmo comportamento de persistência: cada benchmark salva um `benchmark_<timestamp>.json` em `./reports/`; falha ao salvar é logada mas não interrompe a resposta.
- Mesmos defaults: porta 8080, `REPORTS_PATH=./reports`, `TOOL_SELECTION_MIN_ACCURACY=0.80`, `TOOL_SELECTION_MAX_UNNECESSARY=0.10`.

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Framework web | Spring Boot / `@RestController` | `net/http` com `http.ServeMux` |
| LLM | Spring AI `ChatClient` | Cliente OpenAI HTTP próprio (`openai/client.go`) |
| Recurso embarcado | `ClassPathResource` (dentro do jar) | `go:embed` (dentro do binário) |
| Gerador de números aleatórios | `java.util.Random` (seed 42 para benchmark, 99 para memória) | `math/rand` com os mesmos seeds — preserva a **reprodutibilidade** (mesma execução sempre gera os mesmos números) e a mesma faixa/distribuição estatística das métricas simuladas, mas não replica a sequência bit-a-bit do LCG de `java.util.Random` (mesmo princípio já usado nos demais portes deste repositório que envolvem simulação, ex. `ecommerce-neural-recommender-go`) |
| Timestamp de criação do relatório (`GET /evals/reports`) | `BasicFileAttributes.creationTime()` | `os.FileInfo.ModTime()` — Go não expõe data de criação de arquivo de forma portável entre SOs; como os relatórios nunca são modificados após gerados, `ModTime` é equivalente na prática |

## Como executar

```bash
cp .env.example .env   # preencha OPENAI_API_KEY
go run .
```

```bash
curl -X POST http://localhost:8080/evals/benchmark
curl -X POST http://localhost:8080/evals/tool-selection
curl -X POST http://localhost:8080/evals/memory-impact
curl http://localhost:8080/evals/reports
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem os 4 cenários de `EvalsTest.java` (dataset com 5 cenários, benchmark com 3 arquiteturas e veredicto completo, `MemoryEvaluator` em faixas válidas) — o cenário de `ToolSelectionReport` foi adaptado para exercitar a lógica real de `ToolSelectionEvaluator.Evaluate` (em vez de apenas montar o struct diretamente como no teste Java) usando duplos de teste que sempre acertam ou sempre erram a ferramenta esperada, além da persistência e listagem de relatórios — tudo sem depender da API real da OpenAI.
