# Agendamento Médico (Java / Spring Boot)

Chat de agendamento médico que identifica a intenção do paciente (agendar, cancelar ou desconhecida) usando um LLM com saída estruturada, e roteia para o serviço correspondente. Porte em Java/Spring Boot + Spring AI do protótipo original em TypeScript/LangGraph (`03-medical-appointment-z`), servindo também de referência para o porte em Go [`agendamento-medico-go`](../agendamento-medico-go).

## Fluxo

```
POST /chat → IntentService (LLM, JSON estruturado) → schedule / cancel / mensagem amigável → resposta
```

`AppointmentOrchestrator` substitui o `StateGraph` do LangGraph por um roteamento condicional simples em Java: identifica a intenção e despacha para `schedule`, `cancel` ou a mensagem de fallback.

## O que foi mantido 1:1

- Mesmo endpoint `POST /chat`, mesmo contrato de request/response (`{"question": "..."}` → `{"reply": "..."}`).
- Mesmos 3 profissionais fixos (`Dr. Alicio da Silva`, `Dra. Ana Pereira`, `Dra. Carol Gomes`) e os mesmos 2 agendamentos de exemplo pré-cadastrados.
- Mesmo prompt de identificação de intenção (mesmos campos: `intent`, `patientName`, `professionalId`, `professionalName`, `datetime`, `reason`).
- Mesma tolerância a JSON cercado em markdown (` ```json ... ``` `) na resposta do LLM — via `BeanOutputConverter` do Spring AI — e mesmo fallback para intent `"unknown"` quando o parse falha.
- Mesma lógica de conflito de horário (`checkAvailability`) e mesmo erro ao cancelar um agendamento inexistente.
- Mesmo tratamento de erro: falha ao agendar/cancelar gera uma mensagem de erro amigável via LLM (`generateErrorMessage`).
- Mesmos defaults: porta 3000, `temperature=0.2`, `max_tokens=500`, mesma lista de 5 modelos de fallback (via OpenRouter).

## O que foi adaptado

- **Spring AI** (`ChatModel`/`ChatClient`) no lugar de chamadas HTTP manuais ao OpenRouter — `ResilientChatClient` itera sobre a lista de modelos de fallback (`app.fallback-models`) tentando cada um até obter sucesso.
- **`BeanOutputConverter<IntentResult>`** faz o parse estruturado da resposta do LLM (equivalente ao `zod`/parse manual de JSON do TypeScript e do porte Go), já tolerando cercas markdown.
- Modelo de dados como `record`s Java (`Appointment`, `Professional`, `IntentResult`) — campos opcionais (`patientName`, `professionalId`, etc.) são `null` quando ausentes, sem necessidade de ponteiros como em Go.
- `AppointmentService` em memória com `List` simples — sem tratamento explícito de concorrência (protótipo didático/single-user, assim como o original; o porte Go usa `sync.Mutex` por atender requisições concorrentes por padrão).

## Como executar

```bash
cp .env.example .env   # preencha OPENROUTER_API_KEY
export OPENROUTER_API_KEY=...   # ou exporte a variável de outra forma
JAVA_HOME="C:/Program Files/Amazon Corretto/jdk25.0.3_9" mvn spring-boot:run
```

```bash
curl -X POST http://localhost:3000/chat \
  -H "Content-Type: application/json" \
  -d '{"question":"quero agendar uma consulta com a Dra Ana amanha as 15h"}'
```

## Como testar

```bash
JAVA_HOME="C:/Program Files/Amazon Corretto/jdk25.0.3_9" mvn compile
JAVA_HOME="C:/Program Files/Amazon Corretto/jdk25.0.3_9" mvn test
```

Os testes cobrem, sem depender da API real do OpenRouter (LLM sempre mockado via Mockito):

- `AppointmentServiceTest` — disponibilidade, conflito de horário, cancelamento (sucesso e não encontrado).
- `IntentServiceTest` — JSON puro, JSON cercado em markdown, resposta inválida (fallback para `unknown`).
- `MessageGeneratorServiceTest` — geração de mensagens de sucesso, erro e fallback de intenção desconhecida.
- `AppointmentOrchestratorTest` — agendamento com sucesso, conflito de horário, cancelamento com sucesso, cancelamento não encontrado, intenção desconhecida.
- `ChatControllerTest` (`@WebMvcTest` + `MockMvc`) — contrato HTTP do `/chat` e validação de `question` com menos de 5 caracteres.
- `ResilientChatClientTest` — fallback para o próximo modelo quando o primeiro falha, e erro quando todos falham.
