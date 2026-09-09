# Notas API (porte Java)

Porte 1:1 de [`notas-api/`](../notas-api/) (TypeScript/Node), a API de gerenciamento de
tarefas construída na Unidade 1 do módulo `modulo04-criacao-de-agentes-autonomos-novo`
(GitHub Copilot com instructions, spec-driven development e guardrails).

## Origem e paridade

- `domain/Task.java`, `TaskStatus.java`, `TaskListFilter.java` — porte de `src/domain/task.ts`
  (schemas Zod viram validação explícita no construtor do record e nos enums).
- `service/TaskServiceImpl.java` — porte de `src/service/task-service.ts` (mesma validação:
  título/id não podem ser vazios após trim).
- `store/InMemoryTaskStore.java` e `store/JsonFileTaskStore.java` — portes de
  `store/in-memory-task-store.ts` e `store/json-file-task-store.ts`, incluindo a mesma
  estratégia de persistência atômica (escreve em arquivo temporário e faz rename).
- `web/TaskHttpHandler.java` + `web/ServerMain.java` — porte de `src/http/task-routes.ts` +
  `src/index.ts` (mesmas 4 rotas: `POST /tasks`, `GET /tasks?status=`, `PATCH /tasks/:id/complete`,
  `DELETE /tasks/:id`).
- `cli/TaskCli.java` + `cli/CliMain.java` — porte de `src/cli/commands.ts` + `src/cli.ts`
  (mesmos comandos: `create`, `list`, `complete`, `remove`, mesma mensagem de uso).

## Adaptações

- **Sem framework web.** As 4 rotas são simples o bastante para `com.sun.net.httpserver`
  (builtin do JDK) em vez de Spring Boot — decisão de "flexibilidade de framework" da skill
  `portar-projetos-java-go`: Spring Boot só adicionaria uma dependência pesada sem ganho de
  legibilidade para este tamanho de API.
- **Zod → validação explícita.** `createTaskInputSchema`/`taskIdSchema` (trim + min-length)
  viram checagem manual em `TaskServiceImpl` e no construtor de `Task`, lançando
  `TaskValidationException` (equivalente a `TaskValidationError`).
- **JSON manual com Jackson.** O store de arquivo usa `ObjectMapper`/`ObjectNode` em vez de
  um schema declarativo (não há Zod idiomático em Java para isso); a validação de estrutura
  do arquivo (`{tasks: [...]}`) é feita a mão, lançando `TaskStorePersistenceException` nos
  mesmos casos (JSON inválido, campo `tasks` ausente, item malformado).

## Como rodar

```bash
mvn compile exec:java                      # sobe o servidor HTTP em :3000
mvn compile exec:java -Dexec.mainClass=com.notasapi.cli.CliMain -Dexec.args="list"
mvn test
```
