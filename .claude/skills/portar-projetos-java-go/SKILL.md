---
name: portar-projetos-java-go
description: Cria (ou completa) as versões Java e Go de um projeto prático de um módulo do curso, a partir do projeto original (geralmente TypeScript/Node ou Python). Use quando o usuário pedir para "portar esse projeto para Java/Go", "criar a versão Java/Go do módulo X", "fazer o mesmo que fizemos em Java para Go", ou apontar um módulo cujos projetos ainda não têm par Java e/ou Go.
---

# Portar Projetos para Java e Go

Cada projeto prático de aula (TypeScript/Node, Python) deve ganhar dois projetos-irmãos com paridade funcional: um em Java, um em Go. O objetivo não é traduzir sintaxe linha a linha — é reimplementar o mesmo comportamento observável (mesmos endpoints, mesmas regras de negócio, mesmo fluxo de agente) usando as ferramentas idiomáticas de cada ecossistema.

## Passo 0 — Verificar se o projeto é portável

Nem toda pasta de módulo tem código de aplicação. Antes de tentar portar, confirme que existe lógica real (servidor, agente, script executável) e não apenas artefatos de prompt/documentação.

- **Portável:** há `package.json`/`requirements.txt`/`pyproject.toml` com dependências de execução, e um `src`/entrypoint real.
- **Não portável (pule e avise o usuário):** a pasta só contém `.md` de System Prompt, transcrições, PDFs, CSVs de exemplo — típico do módulo de Gestão de Projetos (`modulo07-*`), onde o "projeto" é um prompt colado manualmente no AI Studio, não uma aplicação.
- **Parcialmente portável:** projetos front-end puros (ex.: Angular do módulo de UX/UI) não têm equivalente 1:1 em Java/Go — porte a **lógica de negócio/backend** que sustentaria aquela tela (regras de validação, contratos de API, orquestração de agente), não a interface visual em si. Deixe isso explícito no README do projeto portado.

## Passo 1 — Mapear a convenção de nomenclatura já em uso

Os módulos 01–04 já têm o padrão estabelecido — sempre confira antes de criar algo novo, para não duplicar:

```
modulo0X-.../
  nome-do-exemplo-z/          ← projeto original (TS/Python)
  nome-equivalente-java/      ← porte Java, pasta irmã no mesmo nível do módulo
  nome-equivalente-go/        ← porte Go, mesmo nível
```

O nome da pasta `-java`/`-go` não precisa ser idêntico ao nome do exemplo original — os pares existentes usam nomes descritivos do que o projeto faz (ex.: `03-medical-appointment-z` → `agendamento-medico-java`). Ao criar a versão Go, reaproveite esse mesmo nome descritivo trocando o sufixo (`agendamento-medico-go`), para manter os três projetos (original, Java, Go) fáceis de associar visualmente.

## Passo 2 — Escolher a stack idiomática por tipo de projeto

O padrão já validado nos módulos existentes:

| Natureza do projeto original | Convenção Java | Convenção Go |
|---|---|---|
| Script/CLI standalone (sem servidor HTTP) — ex. módulo 01 | Maven puro + **LangChain4j**, sem Spring Boot, por padrão | Módulo Go simples (`go.mod`), chamadas HTTP diretas ao provedor (OpenRouter/OpenAI-compatible) via `net/http`, sem framework web, por padrão |
| Servidor HTTP / API (Fastify, Express) — ex. módulos 02, 03, 04 | **Spring Boot** (`spring-boot-starter-parent`, `spring-boot-starter-web`) + Spring AI ou LangChain4j | `net/http` (ou `chi`/`gin` se o projeto tiver múltiplas rotas complexas) + cliente HTTP idiomático para o provedor de LLM |
| Agente com estado/checkpointing (LangGraph) | Spring AI `ChatClient` + orquestração manual via `switch`/State pattern (mesmo padrão de `WorkflowOrchestrator` já usado) | `struct` de estado + função de transição explícita — Go não tem um LangGraph maduro; prefira state machine simples e legível a tentar forçar uma abstração pesada |
| Integração com Neo4j | Spring Data Neo4j | Driver oficial `neo4j-go-driver` |
| Agentes Python (CrewAI, módulo 06) | Spring AI multi-`ChatClient` com papéis definidos por `@Service`, replicando responsabilidade de cada agente do CrewAI | Múltiplas `struct`s de agente com método `Run(ctx) (Result, error)`, orquestradas por uma função coordenadora (equivalente ao "agente manager") |

> 🔑 **Regra fundamental:** replique o *comportamento*, não a biblioteca. Se o original usa LangChain.js com um recurso que não tem equivalente direto em Go, implemente o mesmo resultado observável com o que for idiomático em Go — não force uma tradução literal que produza código não-idiomático.

> ⚙️ **Flexibilidade de framework:** "sem Spring Boot"/"sem framework web" no CLI/standalone é o padrão *default*, não uma proibição rígida. Se o protótipo crescer em complexidade (múltiplos componentes, injeção de dependência útil, exposição futura como endpoint) e Spring Boot deixar o código mais idiomático/legível em Java, use Spring Boot mesmo sem servidor HTTP exposto. O mesmo vale para Go: `chi`, `gin` ou **`echo`** podem ser usados fora do caso "servidor HTTP com múltiplas rotas complexas" sempre que facilitarem — ex. roteamento simples, middleware de validação, ou só para manter o código mais limpo do que `net/http` puro exigiria. Decida caso a caso pelo que resulta em código mais idiomático e legível na linguagem; documente no README do porte qual framework foi escolhido e por quê.

## Passo 3 — Extrair o contrato do projeto original

Antes de escrever qualquer linha em Java/Go, leia do projeto original:
1. `package.json`/`requirements.txt` — dependências reais usadas (não instale o que não é usado).
2. Entrypoint principal e rotas/endpoints (se houver servidor).
3. Schemas de entrada/saída (Zod, Pydantic) — viram Bean Validation em Java e `struct` com validação explícita em Go.
4. Testes existentes (`node:test`, `pytest`) — definem o comportamento esperado; os testes Java/Go devem cobrir os mesmos cenários, não menos.
5. `README.md` do original — objetivo do projeto e como rodar, para replicar no novo README.

## Passo 4 — Gerar os dois projetos

Para cada projeto original portável, produza:
- `<nome>-java/` — Maven, `src/main/java/...`, testes em `src/test/java/...` (JUnit 5 + Mockito/AssertJ, seguindo o padrão dos projetos Java já existentes no repositório), `pom.xml`, `README.md`, `.env.example` se houver credencial.
- `<nome>-go/` — `go.mod`, pacote organizado por responsabilidade (não tudo em `main.go` se o projeto tiver mais de um componente), testes em `_test.go` (pacote `testing` padrão, `testify` se precisar de assertions mais expressivas), `README.md`, `.env.example` se houver credencial.

Ambos os READMEs devem indicar explicitamente: **de qual projeto original foram portados**, **o que foi mantido 1:1** e **o que foi adaptado por não ter equivalente direto na linguagem/ecossistema**.

## Passo 5 — Build → Verify → Fix antes de considerar pronto

- Java: `mvn compile` (ou `mvn test` se houver testes) precisa passar sem erro antes de entregar.
- Go: `go build ./...` e `go vet ./...` precisam passar sem erro; rode `go test ./...` se houver testes.
- Nunca entregue um projeto que não compila/builda como se estivesse pronto — isso quebra a paridade funcional que é o objetivo desta skill.

## Passo 6 — Checklist final

- [ ] O projeto original foi lido por completo (dependências, entrypoint, contratos, testes)?
- [ ] A pasta `-java` e a pasta `-go` foram criadas no mesmo nível do projeto original, dentro do módulo correto?
- [ ] A stack escolhida (Passo 2) é coerente com os projetos já existentes no mesmo módulo?
- [ ] Java compila (`mvn compile`) e Go builda (`go build ./...`) sem erro?
- [ ] Os testes cobrem pelo menos os mesmos cenários do projeto original?
- [ ] O README de cada porte explica origem, paridade e adaptações?
- [ ] Se o projeto original não for portável (sem código de aplicação, ou front-end puro), isso foi comunicado ao usuário em vez de forçar uma criação artificial?
