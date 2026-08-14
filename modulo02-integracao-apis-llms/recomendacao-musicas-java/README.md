# Recomendação de Músicas em Java (Spring Boot)

Porte em Java/Spring Boot (Spring AI + Spring Data JPA) do projeto original em TypeScript [`04-song-highlights-z`](../04-song-highlights-z) — chat de recomendação musical com memória de conversa e preferências persistidas, incluindo extração de preferências e sumarização automática via LLM. Também serve de referência para o porte em Go, [`recomendacao-musicas-go`](../recomendacao-musicas-go).

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

- LangGraph + `MemorySaver`/vector store (TS) → orquestração explícita em `MusicChatOrchestrator` (equivalente ao `StateGraph`: `START → chat → (savePreferences? / summarize?) → END`), com persistência via Spring Data JPA.
- Memória de conversa e preferências: `ConversationMessage`/`UserPreferences` (JPA) persistidos em H2 file-based (`./data/musicdb`, `ddl-auto=update`) no lugar do `MemorySaver` + vector store do LangGraph.
- Cliente OpenRouter próprio (TS) → Spring AI `ChatClient`/`OpenAiChatModel`, configurados via `spring.ai.openai.*`, com fallback entre modelos em `ResilientChatClient`.

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY (ou defina spring.ai.openai.api-key)
JAVA_HOME="<seu JDK 21+>" mvn spring-boot:run
```

```bash
curl -X POST http://localhost:3000/chat -H "Content-Type: application/json" -d '{"userId":"rafael","sessionId":"s1","message":"eu gosto muito de rock progressivo"}'
```

## Testes

```bash
JAVA_HOME="<seu JDK 21+>" mvn compile
JAVA_HOME="<seu JDK 21+>" mvn test
```

Os testes não dependem da API real do OpenRouter nem do arquivo H2 de produção: usam um mock server HTTP JDK puro (`com.sun.net.httpserver.HttpServer`, ver `src/test/java/.../support/MockOpenAiServer.java`) apontado via `spring.ai.openai.base-url`, e um banco H2 em memória isolado por teste (`@DataJpaTest` ou `spring.datasource.url` sobrescrito). Cobrem:

- `ConversationRepositoryTest` / `PreferencesRepositoryTest` — persistência (ordenação por `createdAt`, contagem por sessão, upsert de preferências por `userId`).
- `MemoryServiceTest` — adicionar mensagens e recuperar histórico/contagem.
- `PreferencesServiceTest` — criação idempotente de preferências (`getOrCreate`) e extração de preferências via LLM.
- `MusicChatOrchestratorTest` — resposta básica com persistência de histórico, extração por palavra-chave de preferência musical, sumarização automática ao atingir o threshold.
- `ChatControllerTest` — ponta a ponta via `MockMvc` com contexto Spring real (`@SpringBootTest`): mensagem válida (200, `reply`), mensagem curta demais (400).
