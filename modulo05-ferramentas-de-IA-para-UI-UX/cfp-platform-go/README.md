# CFP Platform API em Go

Porte em Go do backend NestJS mínimo que sustenta o front-end Angular do **CFP Platform** (Call for Papers) — presente de forma idêntica em [`modulo-03/cfp-platform`](../modulo-03/cfp-platform) e [`modulo-04/cfp-plataform_v1`](../modulo-04/cfp-plataform_v1). Ver também [`cfp-platform-java`](../cfp-platform-java), o porte Java irmão deste projeto.

Segue a diretriz da skill `portar-projetos-java-go`: **apenas a lógica de backend que sustenta o front-end é portada — a UI (Angular) permanece fora de escopo.**

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/api` | Health check — `{"message": "Hello API"}` |
| POST | `/api/events` | Cria um evento (`nome`, `endereco`, `capacidade`, `data`) |
| GET | `/api/events` | Lista todos os eventos |
| POST | `/api/speakers` | Cria um palestrante (`name`, `email`, `talkTitle`, `isGDE`) |
| GET | `/api/speakers` | Lista todos os palestrantes |

## O que foi mantido 1:1

- Mesmo prefixo `/api` em todas as rotas (equivalente a `app.setGlobalPrefix('api')`).
- Mesmos 2 recursos (`events`, `speakers`), mesmos campos e mesma validação (`class-validator` → validação manual em `dto.Validate()`: campos de texto não vazios, e-mail com formato válido, `capacidade` numérica positiva, `isGDE` booleano obrigatório).
- Mesmo armazenamento em memória, sem persistência.
- Mesma porta padrão (3000, configurável via `PORT`).

## O que foi adaptado

| Aspecto | TypeScript/NestJS | Go |
|---|---|---|
| Framework web | NestJS (`@Controller`, `class-validator`, `ValidationPipe`) | `net/http` com `http.ServeMux` (Go 1.22+) e validação manual (`dto.Validate()`) |
| Geração de ID | `Math.random().toString(36).substring(2, 9)` | `uuid.New()` truncado a 7 chars hex — mesmo propósito (id curto, não criptográfico) |
| Concorrência | Node.js é single-threaded por padrão | `EventService`/`SpeakerService` protegidos por `sync.Mutex`, já que o servidor Go atende requisições concorrentemente |

## Como executar

```bash
cp .env.example .env
go run .
```

```bash
curl http://localhost:3000/api
curl -X POST http://localhost:3000/api/events -d '{"nome":"DevFest","endereco":"Centro de Convenções","capacidade":500,"data":"2026-08-15"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobre o health check, a validação de `dto` (equivalente aos decorators do class-validator), os services (criação + listagem) e os handlers HTTP ponta-a-ponta (sucesso e validação retornando 400) — mesmos cenários de `CfpPlatformApiTest.java`.
