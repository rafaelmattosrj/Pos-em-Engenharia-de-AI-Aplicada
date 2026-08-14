# Legacy Customer API em Go

Porte em Go do projeto Spring Boot/Spring Security [`legacy-customer-api-java`](../legacy-customer-api-java) — API REST de clientes com autenticação JWT, service tokens, RBAC (role-based access control) e rate limiting por token.

## Endpoints

| Método | Rota | Acesso |
|---|---|---|
| POST | `/auth/login` | público |
| POST | `/auth/service-token` | público (requer header `X-Super-Secret`) |
| GET | `/health` | público |
| GET | `/customers`, `/customers/{id}` | `MEMBER` ou `ADMIN` |
| POST/PUT/DELETE | `/customers`, `/customers/{id}` | somente `ADMIN` |

## O que foi mantido 1:1

- Mesmos 2 usuários fixos (`admin`/`password123`→ADMIN, `member`/`pass456`→MEMBER) e mesmos 5 clientes pré-cadastrados.
- Mesma lógica de JWT (HS256, claim `role`, expiração configurável) e de service tokens (UUID, sempre role ADMIN, gerado via header `X-Super-Secret`).
- Mesmas regras de RBAC por rota e mesmos códigos de status: 401 sem token, 403 com role insuficiente, 404/400 nas operações de CRUD.
- **Rate limiting**: 90 requisições/minuto por token, com refill contínuo ("greedy", tokens voltando gradualmente conforme o tempo passa) — mesma semântica do `Bandwidth.refillGreedy` do Bucket4j usado na versão Java. Aplicado apenas a requisições autenticadas (com header `Authorization: Bearer`), um bucket por token.
- Mesma ordem de filtros: rate limit → autenticação (popula a role) → autorização por rota.
- Mesmos 4 cenários do teste de integração Java (`CustomerApiTest.java`) portados para `integration_test.go`: login válido/inválido, acesso sem token (401), MEMBER tentando escrever (403), ADMIN listando clientes.

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Framework web | Spring Boot / Spring Security (filter chain) | `net/http` com `http.ServeMux` (Go 1.22+, roteamento nativo por método + path params) |
| JWT | `io.jsonwebtoken` (jjwt) | [`github.com/golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) |
| Rate limiting | Bucket4j | Token bucket implementado à mão (`middleware/ratelimit.go`) — mesmo algoritmo, sem dependência externa |
| RBAC | `@EnableWebSecurity` + `SecurityFilterChain` (`.hasRole`/`.hasAnyRole`) | `middleware.RequireRole(handler, roles...)` — wrapper explícito por rota |
| Estado (usuários, clientes, service tokens) | `Map`/`ConcurrentHashMap`/`LinkedHashMap` em memória | mapas Go protegidos por `sync.Mutex`/`sync.RWMutex` (mesma necessidade de thread-safety, já que o servidor Go atende requisições concorrentemente) |

## Como executar

```bash
cp .env.example .env
go run .
```

```bash
# Login
curl -X POST http://localhost:3000/auth/login -d '{"username":"admin","password":"password123"}'

# Listar clientes (com o token retornado acima)
curl http://localhost:3000/customers -H "Authorization: Bearer <token>"
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem `AuthService` (login, service tokens, validação de tokens), `CustomerService` (CRUD completo), o rate limiter (limite de 90/min, buckets isolados por token, bypass para requisições não autenticadas) e testes de integração ponta-a-ponta via `httptest` replicando os cenários de `CustomerApiTest.java`.
