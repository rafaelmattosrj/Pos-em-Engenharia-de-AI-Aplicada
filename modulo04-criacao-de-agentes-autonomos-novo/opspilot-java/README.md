# OpsPilot — núcleo multi-agente (porte Java)

Porte do **núcleo arquitetural** de [`09-multi-agent-systems/`](../09-multi-agent-systems/),
o snapshot final (e cumulativo) do OpsPilot na Unidade 9 de
`modulo04-criacao-de-agentes-autonomos-novo`. Irmão de [`opspilot-go/`](../opspilot-go/).

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

## O que foi portado (origem → porte)

| Original (`09-multi-agent-systems/src/`) | Porte Java |
|---|---|
| `domain/types.ts` (Alert, Incident, Runbook, TraceEvent, ReasoningStrategy...) | `domain/*.java` |
| `store/in-memory-store.ts` (o próprio original usa em testes) | `store/InMemoryOpsStore.java` |
| `agents/tools.ts` (list_alerts, list_incidents, open_incident, resolve_incident, consultar_runbook, check_provider_status) | `tools/OpsTools.java` |
| `team/blackboard.ts` | `team/Blackboard.java`, `BlackboardEntry.java` |
| `team/supervisor.ts` + `supervisor-prompt.ts` | `team/Supervisor.java`, `SupervisorDecision.java`, `DecideNextFn.java` |
| `team/roles.ts` (analista/planejador/executor) | `team/RoleRunners.java`, `RoleRunner.java` |
| `team/team-graph.ts` (StateGraph do LangGraph) | `team/TeamGraph.java` (máquina de estado explícita, mesmo `MAX_HANDOFFS=8` e mesmas mensagens de fallback) |
| `team/team-strategy.ts` | `team/TeamStrategy.java` |
| `agents/model.ts` (`OpsResilientChatModel`, retry+fallback) | `llm/ChatModel.java`, `ResilientChatModel.java` |
| `strategies/react.ts` / `agents/react.ts` | `strategies/ReactStrategy.java` (loop ReAct simplificado, estratégia de referência ao lado do modo equipe) |
| `http/server.ts` (só a rota `POST /chat`) | `web/ChatHttpHandler.java`, `ServerMain.java` |

Toda a lógica do `TeamGraph` foi portada 1:1 comportamentalmente: teto de handoffs
(`MAX_HANDOFFS=8`), decisão inválida do supervisor termina graciosamente com o prefixo
`decisao invalida do supervisor`, resposta final usa o brief do supervisor OU cai para
um resumo do blackboard OU uma mensagem de "sem contribuições" — nessa ordem, igual ao
original.

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

## Adaptações

- **`ChatModel` é uma interface própria**, não uma tradução do `ChatOpenAI`/LangChain do
  original. Uma implementação real (não incluída) chamaria um provedor OpenAI-compatible
  via `java.net.http.HttpClient`. Por padrão, `ServerMain` roda com `FakeChatModel`
  (resposta fixa) — sem depender de `OPENROUTER_API_KEY`, ao contrário do original, que
  falha ao subir sem a chave. Isso é deliberado: mantém o núcleo demonstrável/testável
  sem rede.
- **Decisão estruturada do supervisor** (`withStructuredOutput` do LangChain) virou
  "peça um JSON no prompt e faça o parse manual" (`Supervisor.parseDecision`) — não há
  equivalente direto fora do ecossistema JS/Python para essa funcionalidade específica.
- **Sem framework web** — mesma decisão do porte da Notas API: `com.sun.net.httpserver`.

## Nota de correção

`Supervisor.parseDecision` estava com visibilidade padrão (package-private), o que
impedia `SupervisorTest` (pacote `com.opspilot`) de acessá-lo e quebrava `mvn test`.
Corrigido para `public static` -- sem mudança de comportamento, só visibilidade.

## Como rodar

```bash
mvn compile exec:java   # sobe em :3000, POST /chat {"message": "...", "strategy": "team"}
mvn test
```
