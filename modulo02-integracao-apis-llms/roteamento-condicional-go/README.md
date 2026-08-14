# Roteamento Condicional em Go

Porte em Go do projeto Spring Boot [`roteamento-condicional-java`](../roteamento-condicional-java) — pipeline de nós com aresta condicional (substitui o `StateGraph` do LangGraph): identifica a intenção do input e roteia para transformação de texto (UPPERCASE/LOWERCASE) ou, se desconhecida, cai no fallback via LLM.

## Fluxo

```
START → IdentifyIntent → (condicional) → UpperCase / LowerCase / Fallback(LLM) → ChatResponse → END
```

## O que foi mantido 1:1

- Mesmo endpoint `POST /chat`, mesma validação (`question` com mínimo 5 caracteres) e mesma resposta em **texto puro** (não JSON — igual ao `ResponseEntity<String>` original).
- Mesma lógica de detecção de intenção (`contains("upper")` / `contains("lower")`, case-insensitive).
- Mesmo fallback em cadeia de modelos via OpenRouter quando a intenção é desconhecida, com o mesmo system prompt (`"Você é um assistente útil. Responda em português."`).
- Mesmos defaults: porta 3000, `temperature=0.2`, `max_tokens=200`, mesma lista de 5 modelos gratuitos.
- Estado imutável fluindo pelos nós (`State` em Go, `GraphState` em Java) — cada nó recebe um estado e retorna um novo, sem mutação in-place.

## O que foi adaptado

- Sem Spring Boot: servidor HTTP com `net/http`/`http.ServeMux`, validação manual do request.
- Sem Spring AI: cliente OpenRouter próprio via HTTP puro (`openrouter/`, idêntico ao usado em [`gateway-openrouter-go`](../gateway-openrouter-go)).
- Os "nós" do grafo (`IdentifyIntentNode`, `UpperCaseNode`, etc.) viram funções puras `State -> State` no pacote `graph/` em vez de `@Component`s injetados via Spring DI — mesma composição, sem o framework.

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY
go run .
```

```bash
curl -X POST http://localhost:3000/chat -d '{"question":"converte isso para UPPER case"}'
# CONVERTE ISSO PARA UPPER CASE

curl -X POST http://localhost:3000/chat -d '{"question":"qual a capital do brasil?"}'
# (resposta gerada pelo LLM via fallback)
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem o roteamento (uppercase, lowercase, fallback para LLM, todos os modelos falhando) e o handler HTTP (sucesso, validação) — via `httptest`, sem depender da API real do OpenRouter.
