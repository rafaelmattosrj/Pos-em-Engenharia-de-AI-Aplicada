# Agent Loop Framework em Go

Porte em Go do projeto Spring Boot [`agent-loop-framework-java`](../agent-loop-framework-java) — framework de agent loop com o ciclo **Percepção→Planejamento→Ação→Avaliação**, critérios de parada explícitos e telemetria, exposto via `POST /agent/run`.

Equivalente ao `agent_loop.py`/`planner.py`/`executor.py` das aulas 03-06 do curso Python.

## Critérios de parada

1. `done=true` na `PlanDecision` (tarefa concluída pelo agente)
2. `maxSteps` atingido (default 10)
3. `maxTimeSeconds` atingido (default 120s)
4. `noProgress` detectado (mesma tool repetida N vezes consecutivas, default 3)
5. `CircuitBreaker` aberto (3 respostas inválidas consecutivas do LLM)

## O que foi mantido 1:1

- Mesmo endpoint `POST /agent/run` (`{"input": "..."}` → `{"result": "...", "trace": {...}}`) e `GET /agent/health`.
- Mesmo `AgentContract` (persona "DevOps Agent", mesmo system prompt) e mesmos defaults configuráveis via env vars.
- Mesmo prompt de planejamento (percepção, situação, tools disponíveis, formato JSON esperado) e mesma tolerância a cercas markdown na resposta do LLM.
- Mesmas 4 tools simuladas: `getMetrics`, `getLogs`, `getDeployHistory`, `saveIncident` — mesmos dados simulados e mesmo formato de saída JSON.
- Mesma lógica de `CircuitBreaker` (abre após 3 respostas inválidas consecutivas, reseta em resposta válida) e de detecção de "sem progresso" (mesma tool repetida).
- Mesma telemetria (`AgentTrace`/`StepTrace`) e o mesmo `TraceExporter` (JSON pretty-print, salvo em arquivo).

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Framework web | Spring Boot / `@RestController` | `net/http` com `http.ServeMux` |
| LLM | Spring AI `ChatClient` (OpenAI) | Cliente OpenAI HTTP próprio (`openai/client.go`) |
| DI / singletons | Beans Spring (`@Component`) injetados por construtor | Structs montadas explicitamente em `main.go` |
| Organização de pacotes | Tudo em `core/` (incluindo `AgentLoop`) | `AgentLoop` fica em pacote próprio (`agentloop/`) porque depende tanto de `core` quanto de `observability`, e `observability` já depende de `core` — evita import cycle mantendo a mesma separação de responsabilidades |
| `CircuitBreaker.invalidCount` | `AtomicInteger` | `atomic.Int32` (equivalente direto) |

## Como executar

```bash
cp .env.example .env   # preencha OPENAI_API_KEY
go run .
```

```bash
curl -X POST http://localhost:8080/agent/run -d '{"input":"Diagnosticar degradacao no servico api-gateway"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem os 3 cenários de `AgentLoopTest.java` (respeito ao `maxSteps`, abertura do `CircuitBreaker` após 3 respostas inválidas, erro do `Executor` para tool desconhecida), além das 4 tools, o cliente OpenAI, a telemetria e o handler HTTP ponta-a-ponta — todos via `httptest`/duplos de teste simples, sem depender da API real da OpenAI.
