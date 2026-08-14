# ReAct Agent Prototype — Go

Porte Go de [`react-agent-prototype.js`](../react-agent-prototype.js) (fonte primária — a
Missão Prática #02 pede o protótipo em JavaScript, conforme a ementa) e
[`react_agent_prototype.py`](../react_agent_prototype.py) (referência idêntica em
Python), do Módulo 2.5 — UNIPDS: Arquitetura de Sistemas com IA.

Protótipo do agente único do TrialForge (geração da seção de assentimento do ICF a
partir do protocolo do estudo): loop ReAct (Pensamento → Ação → Observação → Resposta
Final) + ferramenta com schema tipado. Sem fila de mensagens — isso fica para o módulo de
multi-agente.

Inclui também
[`internal/reactagent/provedores_pagos.go`](internal/reactagent/provedores_pagos.go),
porte de referência de [`provedores-pagos.js`](../provedores-pagos.js) /
[`provedores_pagos.py`](../provedores_pagos.py) — ver seção própria abaixo.

## O que foi mantido 1:1

- **Loop ReAct** (`AgenteICF` / `AgenteICFComLimite`): mesma sequência Pensamento → Ação →
  Observação, critério de parada explícito no orquestrador (`maxIteracoes`, default 4 via
  `MaxIteracoesPadrao`), nunca deixado para o modelo. Retorna a mesma "trilha" de
  desenvolvimento (uma entrada por volta).
- **Ferramenta tipada** (`BuscarClausulaRegulatoria`): mesmo schema JSON (`tema`: string,
  `jurisdicao`: enum `ANVISA`/`FDA`, ambos obrigatórios).
- **Execução determinística da ferramenta** (`ExecutarBuscaClausula`): mesmo match por
  palavra-chave (`menor`, `adolescente`, `pediátric`) em vez de string exata, e mesmo
  tratamento de parâmetro ausente/mal formado como falha própria da ferramenta (nunca um
  panic que sobe pro chamador) — reproduzindo o mesmo bug real documentado no original
  (`"jurisdicicao"` grafado errado).
- **Os mesmos 6 casos de teste da ferramenta + 1 caso de parâmetro inválido**
  (`CasosTesteFerramenta`), rodados tanto em `main.rodarTestesFerramenta()` (console, como
  no original) quanto em `executor_test.go` (testes Go).
- **Retry com backoff exponencial** (`RetryingChatClient`): 500ms, 1s, 2s — mesma
  progressão, mesma distinção entre erro transitório (rede/timeout, vale retry) e erro
  terminal (ex.: modelo não encontrado, retry não resolve).
- **Os três cenários de `simularInteracao`** (`SimularInteracao`): convergência normal com
  cláusula encontrada; ferramenta sem resultado (população adulta); e não-convergência com
  escalonamento (`maxIteracoes` forçado a 1 só nesse cenário, como no original — em
  produção o limiar calibrado continua sendo 4).
- Mesmo texto de saída no console, incluindo os comentários pedagógicos (`->`).

## O que foi adaptado (e por quê)

- **Chamada ao Ollama via `net/http` + `encoding/json`, direto na API nativa
  (`POST /api/chat`, com `tools`), sem SDK.** Mesmo padrão de
  [`agent-components-demo-go`](../agent-components-demo-go) e de
  [`ollama-local-llm-chat-go`](../../../modulo01-fundamentos-de-ia-e-llms-para-programadores/ollama-local-llm-chat-go)
  (módulo 1), adaptado para o endpoint nativo de chat com suporte a `tools`. `ChatClient`
  é uma interface (`OllamaReactClient` é a implementação real) especificamente para
  permitir testar `AgenteICF` e `RetryingChatClient` com um dublê (`fakeChatClient`), sem
  depender de um Ollama local rodando durante `go test`.
- **Histórico de mensagens vira `[]Mensagem` (`map[string]interface{}`), não um tipo de
  mensagem totalmente tipado.** O loop precisa reempurrar a mensagem bruta do assistente
  (com `tool_calls`) de volta no histórico exatamente como veio do Ollama — igual ao
  original, que empurra `resposta.message` (um objeto dinâmico) direto na lista. Um mapa
  dinâmico captura a mesma flexibilidade que JS/Python têm nativamente com
  objetos/dicionários, sem forçar uma união de tipos por papel de mensagem.
- **`EhErroTransitorio`: qualquer resposta HTTP não-200 é tratada como erro terminal**, não
  só o `status_code === 404` do original. No original, qualquer status diferente de 404
  cairia no teste genérico de mensagem (que checa por `timeout`/`econnrefused`/etc.) — que
  nunca bate para um erro puramente de status HTTP. Ou seja, o comportamento observável já
  era esse (nenhum status HTTP tratado como transitório); a versão Go só torna essa regra
  explícita (`errors.As` para `*ModeloHTTPError`) em vez de dependê-la implicitamente de um
  teste de string que nunca casa.
- **Retorno do loop vira `ResultadoAgente` com "modo" explícito**
  (`EscalarParaAprovacaoHumana` sempre presente como `bool`) em vez dos objetos de formato
  variável do original (`{rascunho, iteracoes, trilha}` ou `{rascunho: null,
  escalarParaAprovacaoHumana: true, motivo, trilha}`) — mesmo conjunto de informações,
  tipado.

## `provedores_pagos.go` — arquivo de referência (não executado pelo fluxo principal)

Porte de referência de `provedores-pagos.js` / `provedores_pagos.py`: mostra como trocar o
Ollama local por três alternativas pagas (Claude/Anthropic, Gemini/Google, GPT/OpenAI) no
mesmo loop ReAct — o modelo é a única peça que muda; `AgenteICF`, `ExecutarBuscaClausula` e
o critério de parada continuam os mesmos.

- `SchemaClaude`, `SchemaGemini`, `SchemaGPT`: **executáveis de verdade** (usam só a
  biblioteca padrão), constroem o schema real que cada provedor exige para a mesma
  ferramenta — mostrando concretamente a fragmentação de formato entre provedores que
  motiva um protocolo como o MCP (Claude e Gemini usam forma achatada; GPT usa forma
  aninhada, igual ao formato nativo do Ollama já usado em `BuscarClausulaRegulatoria`).
- `ChamarClaude`, `ChamarGemini`, `ChamarGPT`: **não executáveis** — retornam erro
  `"referência não executável: ..."` se chamadas. Nenhum dos três SDKs pagos
  (`github.com/anthropics/anthropic-sdk-go`, `google.golang.org/genai`,
  `github.com/openai/openai-go`) está no `go.mod` deste projeto (só a biblioteca padrão é
  usada) — adicionar um deles exigiria `go get` com rede e credenciais pagas no momento do
  build, o que quebraria `go build` offline para quem só quer rodar a demo local com
  Ollama. O comentário de cada função documenta, em bloco de código, exatamente como o
  corpo real ficaria com a dependência instalada.
- Para usar de verdade: rode `go get` da SDK escolhida, configure a variável de ambiente da
  chave (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY` ou `OPENAI_API_KEY` — ver `.env.example`) e
  substitua o corpo da função pela chamada real documentada no comentário.

## Como rodar

Pré-requisito: Go 1.22+. O fluxo principal (`main.go`) precisa do
[Ollama](https://ollama.com) rodando localmente com `ollama pull gemma4:e2b`.

```bash
# Rodar o protótipo completo (testes da ferramenta + os 3 cenários simulados)
go run .

# Rodar os testes automatizados (ferramenta, retry e loop ReAct com dublê — sem rede)
go test ./...

# Verificação estática
go vet ./...
```

Copie `.env.example` para `.env` se quiser customizar `OLLAMA_BASE_URL` (o valor padrão já
é `http://localhost:11434`). As variáveis de provedores pagos (`ANTHROPIC_API_KEY` etc.)
só são relevantes se você adaptar `provedores_pagos.go` para chamar de verdade.

## Estrutura

```
internal/reactagent/
  tipos.go               — Mensagem, ChamadaFerramenta, ModeloResposta, ChatClient
  schema.go               — schema JSON da ferramenta (formato nativo Ollama/OpenAI)
  executor.go              — execução determinística da ferramenta
  casos_teste.go            — casos de teste compartilhados entre main.go e Go tests
  ollama_client.go           — implementação real via HTTP nativo do Ollama
  retry.go                    — RetryingChatClient (backoff exponencial) + classificação de erro
  agente.go                    — o loop ReAct em si
  simulador.go                  — os 3 cenários de SimularInteracao()
  provedores_pagos.go            — referência: Claude/Gemini/GPT (não executado)
  *_test.go                       — testes (executor, retry, loop ReAct com fakeChatClient)
main.go                            — orquestra testes da ferramenta + simulação
dotenv.go                           — leitura simples de .env
```
