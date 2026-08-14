# Recomendação de Músicas em Go

Porte em Go do projeto Spring Boot [`recomendacao-musicas-java`](../recomendacao-musicas-java) — chat de recomendação musical com memória de conversa e preferências persistidas, incluindo extração de preferências e sumarização automática via LLM.

## Fluxo

```
POST /chat → carrega preferências + histórico → gera resposta (LLM) → persiste mensagens
           → threshold atingido? sumariza + extrai preferências
           → senão, mensagem contém preferência musical? extrai preferências
```

## O que foi mantido 1:1

- Mesmo endpoint `POST /chat`, mesmo contrato (`userId`/`sessionId` opcionais com defaults `"default-user"`/`"default-session"`, `message` com mínimo 3 caracteres) → `{"reply": "..."}`.
- Mesma montagem de contexto: resumo anterior + preferências + histórico da sessão, no mesmo formato de prompt.
- Mesma detecção de "contém preferência musical" (`gosto`, `adoro`, `prefiro`, `curto`, `favorit`, `ouço`, case-insensitive).
- Mesma lógica de sumarização automática ao atingir o threshold de mensagens (`app.summarize-after-messages`, default 10), incluindo o fato de que sumarizar e extrair preferências são mutuamente exclusivos com a extração "leve" por palavra-chave.
- Mesmos defaults: porta 3000, `temperature=0.7`, `max_tokens=500`, mesma lista de 5 modelos de fallback.

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Persistência | H2 file-based (JPA/Hibernate, `ddl-auto=update`) | SQLite via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (driver puro-Go, sem cgo) — mesmo conceito de banco embarcado em um único arquivo (`./data/musicdb.sqlite`) |
| Repositórios | Spring Data JPA (`ConversationRepository`, `PreferencesRepository`) | `database/sql` com queries manuais (`repository/`) |
| Servidor HTTP | Spring Boot / `@RestController` | `net/http`/`http.ServeMux` |
| Cliente LLM | Spring AI `ChatClient` | Cliente HTTP próprio (`openrouter/`, `llm/`), igual aos demais portes deste repositório |

Como o servidor Go atende requisições concorrentemente por padrão, a concorrência de escrita no SQLite é resolvida pelo próprio `database/sql` (pool de conexões); não há necessidade de mutex adicional como em `agendamento-medico-go`, pois não há estado mutável em memória fora do banco.

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY
go run .
```

```bash
curl -X POST http://localhost:3000/chat -d '{"userId":"rafael","sessionId":"s1","message":"eu gosto muito de rock progressivo"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem os repositórios (SQLite `:memory:`), os services (`MemoryService`, `PreferencesService`), o orquestrador (resposta básica, extração por palavra-chave, sumarização por threshold) e o handler HTTP — todos via `httptest`/SQLite em memória, sem depender da API real do OpenRouter.
