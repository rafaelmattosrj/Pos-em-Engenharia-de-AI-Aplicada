# CFP Platform API em Java

Porte em Java/Spring Boot do backend NestJS mínimo que sustenta o front-end Angular do **CFP Platform** (Call for Papers) — presente de forma idêntica em [`modulo-03/cfp-platform`](../modulo-03/cfp-platform) e [`modulo-04/cfp-plataform_v1`](../modulo-04/cfp-plataform_v1) (mesmo backend, apenas o front-end Angular evoluiu entre as duas versões do curso).

Este porte segue a diretriz da skill `portar-projetos-java-go`: **apenas a lógica de backend que sustenta o front-end é portada — a UI (Angular) permanece fora de escopo.**

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/api` | Health check — `{"message": "Hello API"}` |
| POST | `/api/events` | Cria um evento (`nome`, `endereco`, `capacidade`, `data`) |
| GET | `/api/events` | Lista todos os eventos |
| POST | `/api/speakers` | Cria um palestrante (`name`, `email`, `talkTitle`, `isGDE`) |
| GET | `/api/speakers` | Lista todos os palestrantes |

## O que foi mantido 1:1

- Mesmo prefixo global `/api` (equivalente a `app.setGlobalPrefix('api')` do `main.ts`), aqui via `server.servlet.context-path=/api`.
- Mesmos 2 recursos (`events`, `speakers`), mesmos campos e mesma validação (`class-validator` → Bean Validation: `nome`/`endereco`/`talkTitle`/`name` não vazios, `email` válido, `capacidade` numérica, `isGDE` booleano).
- Mesmo armazenamento em memória, sem persistência — mesmo estágio de "minimal backend" explicitado no `main.ts` original.
- Mesma porta padrão (3000, configurável via `PORT`).

## O que foi adaptado

| Aspecto | TypeScript/NestJS | Java/Spring Boot |
|---|---|---|
| Framework | NestJS (`@Controller`, `@Injectable`, `class-validator`) | Spring Boot (`@RestController`, `@Service`, Bean Validation) |
| Geração de ID | `Math.random().toString(36).substring(2, 9)` (7 chars base36) | `UUID.randomUUID()` truncado a 7 chars hex — mesmo propósito (id curto, não criptográfico), formato ligeiramente diferente |
| Tipos compartilhados | `@cfp-platform/shared-types` (lib Nx) | `model.Event`/`model.Speaker` (records) definidos localmente, já que não há workspace compartilhado entre Java e o monorepo Nx |

## Como executar

```bash
mvn spring-boot:run
```

```bash
curl http://localhost:3000/api
curl -X POST http://localhost:3000/api/events -H "Content-Type: application/json" \
  -d '{"nome":"DevFest","endereco":"Centro de Convenções","capacidade":500,"data":"2026-08-15"}'
```

## Testes

```bash
mvn test
```

Cobre o health check (equivalente a `api-e2e/src/api/api.spec.ts`) e os endpoints de `events`/`speakers` (criação com id gerado, validação retornando 400, listagem).
