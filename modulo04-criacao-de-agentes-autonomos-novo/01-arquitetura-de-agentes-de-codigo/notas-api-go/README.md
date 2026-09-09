# Notas API (porte Go)

Porte 1:1 de [`notas-api/`](../notas-api/) (TypeScript/Node), a API de gerenciamento de
tarefas construída na Unidade 1 do módulo `modulo04-criacao-de-agentes-autonomos-novo`.
Irmão de [`notas-api-java/`](../notas-api-java/).

## Origem e paridade

- `domain/task.go` — porte de `src/domain/task.ts` (schemas Zod viram `Parse*`/`RequireNonBlank`).
- `service/service.go` — porte de `src/service/task-service.ts`.
- `store/in_memory_store.go` e `store/json_file_store.go` — portes de `store/in-memory-task-store.ts`
  e `store/json-file-task-store.ts`, incluindo a mesma escrita atômica (arquivo temporário + rename).
- `httpapi/handler.go` — porte de `src/http/task-routes.ts` (mesmas 4 rotas).
- `cli/cli.go` — porte de `src/cli/commands.ts` (mesmos comandos e mensagens de uso).
- `cmd/server/main.go` e `cmd/cli/main.go` — portes de `src/index.ts` e `src/cli.ts`.

## Adaptações

- **Sem framework web.** `net/http` puro (`http.ServeMux`), sem chi/gin — 4 rotas não
  justificam a dependência extra, mesma decisão tomada no porte Java.
- **IDs:** `github.com/google/uuid`, já usado em outros portes Go do repositório
  (substitui `randomUUID()` do Node).
- **Validação de schema de arquivo mais simples que o Zod:** o Go `encoding/json` não
  distingue "campo `tasks` ausente" de "campo `tasks` presente e vazio" sem um tipo ponteiro
  extra — um arquivo `{}` é aceito como armazenamento vazio aqui, enquanto o original rejeitaria
  por o Zod schema exigir a chave `tasks`. Documentado nos testes (`json_file_store_test.go`).

## Como rodar

```bash
go run ./cmd/server              # sobe o servidor HTTP em :3000
go run ./cmd/cli -- list
go build ./...
go vet ./...
go test ./...
```
