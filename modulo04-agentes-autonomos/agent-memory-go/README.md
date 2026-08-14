# Agent Memory em Go

Porte em Go do projeto Spring Boot [`agent-memory-java`](../agent-memory-java) — agente com consciência de memória, integrando os **4 tipos de memória** (curto prazo, longo prazo, episódica, contextual) e um motor de **reflexão evolutiva** que extrai lições de execuções passadas.

Equivalente ao `memory_aware_agent.py` (aula 13) e `reflection_engine.py` (aula 14) do curso Python.

## Fluxo (`POST /agent-memory/run`)

```
contexto semântico (embeddings) → fatos (longo prazo) → episódios recentes
  → execução enriquecida (LLM) → persiste episódio → reflexão (extrai lições)
```

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| POST | `/agent-memory/run` | Executa com os 4 tipos de memória ativos |
| POST | `/agent-memory/run-without` | Executa sem memória — LLM direto, para comparação |
| GET | `/agent-memory/episodes` | Lista episódios, do mais recente ao mais antigo |
| GET | `/agent-memory/lessons` | Lista lições extraídas pela reflexão |
| DELETE | `/agent-memory/clear` | Remove os arquivos de memória persistida |

## O que foi mantido 1:1

- Mesmos 4 tipos de memória e mesma responsabilidade de cada um:
  - **Curto prazo** (`memory.ShortTermMemory`): estado volátil da execução atual, isolado por chamada (equivalente ao escopo *prototype* do Spring).
  - **Longo prazo** (`memory.LongTermMemory`): fatos persistidos em `long-term.json`, deduplicados por conteúdo (case-insensitive), com suporte a expiração.
  - **Episódica** (`memory.EpisodicMemory`): histórico de execuções em `episodes.json`, ordenado do mais recente ao mais antigo.
  - **Contextual** (`memory.ContextualMemory`): busca semântica via embeddings + similaridade de cosseno, com threshold configurável (default 0.7).
- Mesmo motor de reflexão (`engine.ReflectionEngine`): só reflete sobre episódios cujo outcome contém `erro`/`error`/`falha`/`notável`/`aprendizado`; mesmo prompt estruturado, mesmo parsing de campos (`SITUAÇÃO:`, `AÇÃO:`, `RESULTADO:`, `APRENDIZADO:`, `GENERALIZABILIDADE:`), mesmo filtro (só persiste lições com aprendizado E generalizabilidade preenchidos), persistido em `lessons.json`.
- Mesmo prompt enriquecido (contexto semântico + fatos + histórico + tarefa atual) montado por `MemoryAwareAgent`.
- Mesmos defaults: porta 8080, modelo `gpt-4o-mini`, embeddings `text-embedding-3-small`, `MEMORY_PATH=./data`, `SIMILARITY_THRESHOLD=0.7`.

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| Framework web | Spring Boot / `@RestController` | `net/http` com `http.ServeMux` |
| LLM + Embeddings | Spring AI `ChatClient`/`EmbeddingModel` (OpenAI) | Cliente HTTP próprio (`openai/client.go`) com `Chat` e `Embed` |
| `ShortTermMemory` escopo *prototype* | Bean Spring recriado por injeção (`ApplicationContext.getBean`) | `memory.NewShortTermMemory()` chamado diretamente no início de `Run` — mesmo efeito (isolamento total por execução), sem um container DI |
| Persistência | Jackson (`ObjectMapper` + `JavaTimeModule`) em JSON | `encoding/json` da stdlib (`time.Time` já serializa em RFC 3339 nativamente) |

## Como executar

```bash
cp .env.example .env   # preencha OPENAI_API_KEY
go run .
```

```bash
curl -X POST http://localhost:8080/agent-memory/run -d '{"input":"qual o status do servico de pagamentos?"}'
curl http://localhost:8080/agent-memory/episodes
curl http://localhost:8080/agent-memory/lessons
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem os 3 cenários de `MemoryTest.java` (deduplicação em `LongTermMemory`, filtro por threshold em `ContextualMemory` via um duplo de teste no lugar do mock Mockito de `EmbeddingModel`, `GetRecentEpisodes` respeitando o limite), além de `ReflectionEngine` (episódios sem outcome notável, extração bem-sucedida, resposta `SEM_LICAO`, generalizabilidade ausente), `MemoryAwareAgent` (fluxo completo, persistência de episódios, uso do histórico entre chamadas) e os handlers HTTP — tudo com duplos de teste simples, sem depender da API real da OpenAI.
