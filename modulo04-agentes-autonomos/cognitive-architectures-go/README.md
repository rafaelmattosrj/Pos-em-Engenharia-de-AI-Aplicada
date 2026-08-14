# Cognitive Architectures em Go

Porte em Go do projeto Spring Boot [`cognitive-architectures-java`](../cognitive-architectures-java) — 3 arquiteturas cognitivas (**ReAct**, **Plan-and-Execute**, **Reflection**) expostas via `POST /agents/run`, roteadas pelo campo `architecture` do request.

Equivalente a `react_agent.py` (aula 07), `plan_execute_agent.py`/`reflection_agent.py`/`critique_evaluator.py` (aula 08) do curso Python.

## Arquiteturas

- **ReAct** (`react`): ciclo Thought → Action → Observation a cada step; para ao emitir `Action: FINAL_ANSWER(...)` ou atingir `maxSteps`. Proteção anti-loop: encerra se a mesma Action repetir 3x.
- **Plan-and-Execute** (`plan-execute`): uma única chamada ao LLM gera o plano completo (JSON estruturado); os steps são então executados deterministicamente (simulados), sem novas chamadas ao LLM.
- **Reflection** (`reflection`): gera um output, avalia com o `CritiqueEvaluator` (LLM como juiz em 3 dimensões), regenera incorporando o feedback — repete até aprovar ou atingir `maxCycles`.

## O que foi mantido 1:1

- Mesmo endpoint `POST /agents/run` (`{"architecture": "...", "input": "..."}` → `{"result": "...", "metrics": {...}}`) e mesmo roteamento por `architecture` (case-insensitive, trim).
- Mesma validação: 400 se `architecture`/`input` ausentes; **200 com mensagem de erro no body** (não 400) para arquitetura desconhecida — comportamento intencional replicado da versão Java.
- Mesmos prompts (ReAct, planejamento, geração/regeneração da Reflection, avaliação crítica) e mesma lógica de parsing (extração de campos `Thought`/`Action`, JSON do plano tolerante a texto extra, scores `CORRECTNESS`/`COMPLETENESS`/`QUALITY`/`FEEDBACK` via regex).
- Mesmas 6 ferramentas simuladas do Plan-and-Execute (`search`, `calculate`, `summarize`, `validate`, `format`, `store`) com as mesmas mensagens de resultado.
- Mesma métrica de tokens estimados (`(prompt+resposta) / 4` caracteres) e mesmos defaults: porta 8080, `REACT_MAX_STEPS=10`, `REFLECTION_MAX_CYCLES=3`, `REFLECTION_THRESHOLD=0.7`.

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Framework web | Spring Boot / `@RestController` | `net/http` com `http.ServeMux` |
| LLM | Spring AI `ChatClient` | Cliente OpenAI HTTP próprio (`openai/client.go`) |
| Tratamento de erro do LLM | Exceção não capturada propaga até o Spring (500 default) | `ChatClient.Chat` retorna `error`; agentes propagam via `(model.AgentResponse, error)`, handler traduz para 500 — mesmo resultado observável |

## Como executar

```bash
cp .env.example .env   # preencha OPENAI_API_KEY
go run .
```

```bash
curl -X POST http://localhost:8080/agents/run -d '{"architecture":"react","input":"Qual a capital do Brasil?"}'
curl -X POST http://localhost:8080/agents/run -d '{"architecture":"plan-execute","input":"Pesquisar e resumir noticias de tecnologia"}'
curl -X POST http://localhost:8080/agents/run -d '{"architecture":"reflection","input":"Explique machine learning para iniciantes"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem os 3 cenários de `CognitiveArchitecturesTest.java` (`CritiqueEvaluator` reprovando abaixo do threshold, `ReflectionAgent` encerrando após `maxCycles` sem aprovar, roteamento do handler para o agente correto e erro gracioso para arquitetura desconhecida), além de cenários adicionais para `ReactAgent` (resposta final, loop de mesma action, `maxSteps`) e `PlanExecuteAgent` (parse de plano válido, fallback para JSON inválido) — todos com duplos de teste simples, sem depender da API real da OpenAI.
