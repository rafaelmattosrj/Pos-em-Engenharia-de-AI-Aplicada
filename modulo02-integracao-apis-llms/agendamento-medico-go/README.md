# Agendamento Médico em Go

Porte em Go do projeto Spring Boot [`agendamento-medico-java`](../agendamento-medico-java) — chat de agendamento que identifica a intenção do paciente (agendar, cancelar ou desconhecida) usando um LLM com saída estruturada, e roteia para o serviço correspondente.

## Fluxo

```
POST /chat → IdentifyIntent (LLM, JSON estruturado) → schedule / cancel / mensagem amigável → resposta
```

## O que foi mantido 1:1

- Mesmo endpoint `POST /chat`, mesmo contrato de request/response (`{"question": "..."}` → `{"reply": "..."}`).
- Mesmos 3 profissionais fixos e os mesmos 2 agendamentos de exemplo pré-cadastrados.
- Mesmo prompt de identificação de intenção (mesmos campos: `intent`, `patientName`, `professionalId`, `professionalName`, `datetime`, `reason`).
- Mesma tolerância a JSON cercado em markdown (` ```json ... ``` `) na resposta do LLM, e mesmo fallback para intent `"unknown"` quando o parse falha.
- Mesma lógica de conflito de horário (`checkAvailability`) e mesmo erro ao cancelar um agendamento inexistente.
- Mesmo tratamento de erro: falha ao agendar/cancelar gera uma mensagem de erro amigável via LLM (`generateErrorMessage`); falha ao gerar a mensagem para intenção desconhecida propaga como erro HTTP 500, igual à exceção não capturada da versão Java.
- Mesmos defaults: porta 3000, `temperature=0.2`, `max_tokens=500`, mesma lista de 5 modelos de fallback.

## O que foi adaptado

- Sem Spring Boot: `net/http`/`http.ServeMux` no lugar de `@RestController`.
- Sem Spring AI/`BeanOutputConverter`: parse de JSON manual (`encoding/json`) com a mesma tolerância a cercas markdown.
- **Concorrência**: `AppointmentService` usa um `sync.Mutex` protegendo a lista de agendamentos em memória — um servidor Go atende requisições HTTP concorrentemente por padrão (o prototype Java, de uso didático/single-user, não trata esse caso).
- Campos opcionais do `IntentResult` (`patientName`, `professionalId`, etc., que em Java são `null`) são `*string`/`*int` em Go.

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY
go run .
```

```bash
curl -X POST http://localhost:3000/chat -d '{"question":"quero agendar uma consulta com a Dra Ana amanha as 15h"}'
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem `AppointmentService` (disponibilidade, conflito, cancelamento), `IntentService` (JSON puro, JSON cercado em markdown, resposta inválida) e o `Orchestrator` (agendamento bem-sucedido, conflito de horário, intenção desconhecida) — todos via `httptest`, sem depender da API real do OpenRouter.
