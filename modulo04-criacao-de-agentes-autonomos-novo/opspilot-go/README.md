# OpsPilot — núcleo multi-agente (porte Go)

Porte do **núcleo arquitetural** de [`09-multi-agent-systems/`](../09-multi-agent-systems/),
o snapshot final (e cumulativo) do OpsPilot na Unidade 9 de
`modulo04-criacao-de-agentes-autonomos-novo`. Irmão de [`opspilot-java/`](../opspilot-java/) --
mesmo escopo, mesmo comportamento observável, adaptado ao idioma de cada ecossistema.

## Por que só o núcleo, e por que a Unidade 9

As Unidades 2–9 são **snapshots cumulativos do mesmo produto** (cada pasta reimplementa
tudo que já existia nas unidades anteriores + a novidade daquela unidade). Portar as 8
pastas separadamente geraria ~8 cópias quase-duplicadas do mesmo código em Java/Go — não
é o espírito da skill `portar-projetos-java-go` ("replicar comportamento, não duplicar
arquivo por arquivo"). Este porte cobre só a **Unidade 9** (`09-multi-agent-systems/`),
o estado final e superset funcional de todas as anteriores — o padrão **multi-agente**
(supervisor + blackboard + papéis) que dá nome a esta unidade e a este módulo.

O código-fonte original da Unidade 9 tem **87 arquivos TypeScript** (SQLite, servidor MCP,
embeddings/memória semântica, sumarização de conversa, roteamento de estratégia, frontend
web em Vite/React, auditoria de requests). Portar tudo isso file-a-file não teria retorno
proporcional ao esforço nem seria idiomático. Este porte cobre o que **é o ponto da unidade**
— o padrão arquitetural multi-agente — e deixa o resto documentado abaixo como fora de escopo.
O mesmo recorte de escopo já usado em `opspilot-java/` foi aplicado aqui, para manter os
dois portes comparáveis lado a lado.

## O que foi portado (origem → porte)

| Original (`09-multi-agent-systems/src/`) | Porte Go |
|---|---|
| `domain/types.ts`, `domain/severity.ts`, `domain/errors.ts` | `domain/types.go`, `domain/errors.go` |
| `store/in-memory-store.ts` (o próprio original usa em testes) | `store/inmemory_ops_store.go` (`OpsStore` em `store/opsstore.go`) |
| `agents/tools.ts` (list_alerts, list_incidents, open_incident, resolve_incident, consultar_runbook, check_provider_status) | `tools/ops_tools.go` (`Tool` em `tools/tool.go`) |
| `team/blackboard.ts` | `team/blackboard.go` |
| `team/supervisor.ts` + `supervisor-prompt.ts` | `team/supervisor.go` |
| `team/roles.ts` (analista/planejador/executor) | `team/roles.go` |
| `team/team-graph.ts` (StateGraph do LangGraph) | `team/team_graph.go` (máquina de estado explícita: `TeamGraph.Run` guarda o estado da rodada em variáveis locais e usa uma função de transição em vez de nós/arestas declarativos; mesmo `MaxHandoffs=8` e mesmas mensagens de fallback) |
| `team/team-strategy.ts` | `team/team_strategy.go` |
| `agents/model.ts` (`OpsResilientChatModel`, retry+fallback) | `llm/chat_model.go`, `llm/resilient_chat_model.go` |
| `strategies/react.ts` / `agents/react.ts` | `strategies/react_strategy.go` (loop ReAct simplificado, estratégia de referência ao lado do modo equipe) |
| `http/server.ts` (só a rota `POST /chat`) | `httpapi/chat_handler.go`, `cmd/server/main.go` |

Toda a lógica do `TeamGraph` foi portada 1:1 comportamentalmente: teto de handoffs
(`MaxHandoffs=8`), decisão inválida do supervisor termina graciosamente com o prefixo
`decisao invalida do supervisor`, resposta final usa o brief do supervisor OU cai para
um resumo do blackboard OU uma mensagem de "sem contribuições" — nessa ordem, igual ao
original e ao porte Java. Os pacotes `analista`/`executor` seguem restritos às mesmas
ferramentas (leitura vs. mutação de incidentes) descritas em `team-strategy.ts` (FR-007).

## Organização dos pacotes

Seguindo a convenção Go "structs de estado + função de transição explícita" (a
mesma decisão de arquitetura do `WorkflowOrchestrator`/state pattern do porte Java, sem
tentar forçar um LangGraph que Go não tem maduro):

```
domain/      tipos centrais + ReasoningStrategy (contrato comum a react e team)
llm/         ChatModel (interface), FakeChatModel, ResilientChatModel
store/       OpsStore (interface) + InMemoryOpsStore
tools/       Tool (interface) + ferramentas concretas (list_alerts, open_incident, ...)
strategies/  ReactStrategy (baseline single-agent)
team/        Blackboard, Roles (RoleRunner: analista/planejador/executor),
             Supervisor (DecideNextFn), TeamGraph (coordenador), TeamStrategy
httpapi/     ChatHandler -- POST /chat
cmd/server/  main() -- bootstrap equivalente a src/index.ts
```

Cada papel do time (`analista`, `planejador`, `executor`) é uma struct que implementa a
interface `RoleRunner` com um método `Run(RoleRunInput) (RoleRunResult, error)`; `TeamGraph.Run`
é a função coordenadora que invoca o supervisor e delega para o `RoleRunner` certo a cada
rodada, replicando comportamentalmente o grafo do LangGraph sem depender de um framework
de grafos.

## O que ficou fora de escopo (e por quê)

- **MCP server** (`src/mcp/`) — protocolo específico do ecossistema de tooling do
  Claude/Copilot, sem contrapartida de comportamento a portar aqui.
- **SQLite** (`SqliteOpsStore`, `SqliteConversationStore`, etc.) — trocado por
  `InMemoryOpsStore`, que é exatamente o que o *próprio original* usa nos testes; a escolha
  de motor de persistência é implementação, não comportamento do domínio.
- **Embeddings / memória semântica** (`memory/`) — dependeria de um modelo de embeddings
  real; fora do escopo de um porte arquitetural sem chamada de rede.
- **Sumarização de histórico, roteamento de estratégia (`graph/`), gerenciamento de
  contexto/tokens (`context/`)** — camadas de otimização de prompt em cima do padrão
  multi-agente, não o padrão em si.
- **Frontend web (`web/`, Vite/React)** — fora do escopo de um porte de backend.
- **Auditoria de requests / `/stats`** — observabilidade adicional, não o padrão de agente.

## Adaptações (e onde este porte diverge deliberadamente do Java)

- **`ChatModel` é uma interface própria**, não uma tradução do `ChatOpenAI`/LangChain do
  original. Uma implementação real (não incluída) chamaria um provedor OpenAI-compatible
  via `net/http`. Por padrão, `cmd/server/main.go` roda com `FakeChatModel` (resposta fixa)
  — sem depender de `OPENROUTER_API_KEY`, ao contrário do original, que falha ao subir sem
  a chave. Isso é deliberado: mantém o núcleo demonstrável/testável sem rede.
- **Decisão estruturada do supervisor** (`withStructuredOutput` do LangChain) virou
  "peça um JSON no prompt e faça o parse manual" (`team.ParseDecision`) — não há
  equivalente direto fora do ecossistema JS/Python para essa funcionalidade específica.
- **Erros são valores, não exceções.** `ChatModel.Invoke`, `Tool.Execute`,
  `OpsStore.ResolveIncident/GetRunbook` e `ReasoningStrategy.Run` retornam `error`
  explícito. Isso é uma divergência deliberada do porte Java, que deixa exceções não
  verificadas (`IllegalArgumentException`, `ModelUnavailableException`) estourarem sem
  tratamento em alguns caminhos (ex.: `OpsTools` com severidade inválida, ou o handler HTTP
  que não captura falha do modelo). Em Go isso não é idiomático — o handler HTTP
  (`httpapi.ChatHandler`) converte qualquer erro de estratégia em `500` com corpo JSON, e as
  ferramentas convertem erro de argumento inválido numa observação de trace
  (`"Erro: ..."`) em vez de interromper o processo do agente.
- **`RoleRunner` não configurado** no `TeamGraph` produz uma entrada de erro no blackboard
  em vez de um nil-pointer/NPE (o porte Java, como o original TS, assume implicitamente que
  todos os papéis estão sempre presentes no mapa).
- **Sem framework web** — `net/http` puro (`http.ServeMux` + `http.Handler`), mesma decisão
  de escopo do porte Java (`com.sun.net.httpserver`) e da Notas API deste módulo.
- **Sem dependências externas** — geração de ID do incidente usa `crypto/rand` (em vez de
  `github.com/google/uuid`, usado em outros portes Go do repositório) para manter o
  `go.mod` sem dependências, já que o comportamento pedido (id legível e de baixa colisão)
  não exige um UUID RFC 4122 completo.

## Como rodar

```bash
go build ./...
go vet ./...
go test ./...

go run ./cmd/server   # sobe em :3000 (ou $PORT), POST /chat {"message": "...", "strategy": "team"}
```

Exemplo de chamada (com o servidor no ar):

```bash
curl -X POST http://localhost:3000/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"o que esta pegando no checkout-api?","strategy":"team"}'
```
