# ReAct Agent Prototype — Java

Porte Java de [`react-agent-prototype.js`](../react-agent-prototype.js) (fonte primária —
a Missão Prática #02 pede o protótipo em JavaScript, conforme a ementa) e
[`react_agent_prototype.py`](../react_agent_prototype.py) (referência idêntica em
Python), do Módulo 2.5 — UNIPDS: Arquitetura de Sistemas com IA.

Protótipo do agente único do TrialForge (geração da seção de assentimento do ICF a
partir do protocolo do estudo): loop ReAct (Pensamento → Ação → Observação → Resposta
Final) + ferramenta com schema tipado. Sem fila de mensagens — isso fica para o módulo de
multi-agente.

Inclui também [`ProvedoresPagos.java`](src/main/java/com/trialforge/reactagent/ProvedoresPagos.java),
porte de referência de [`provedores-pagos.js`](../provedores-pagos.js) /
[`provedores_pagos.py`](../provedores_pagos.py) — ver seção própria abaixo.

## O que foi mantido 1:1

- **Loop ReAct** (`AgenteICF.agenteICF`): mesma sequência Pensamento → Ação → Observação,
  critério de parada explícito no orquestrador (`maxIteracoes`, default 4), nunca deixado
  para o modelo. Retorna a mesma "trilha" de desenvolvimento (uma entrada por volta).
- **Ferramenta tipada** (`ToolSchema.buscarClausulaRegulatoria`): mesmo schema JSON
  (`tema`: string, `jurisdicao`: enum `ANVISA`/`FDA`, ambos obrigatórios).
- **Execução determinística da ferramenta** (`ExecutorFerramenta.executarBuscaClausula`):
  mesmo match por palavra-chave (`menor`, `adolescente`, `pediátric`) em vez de string
  exata, e mesmo tratamento de parâmetro ausente/mal formado como falha própria da
  ferramenta (nunca uma exceção que estoura pro chamador) — reproduzindo o mesmo bug real
  documentado no original (`"jurisdicicao"` grafado errado).
- **Os mesmos 6 casos de teste da ferramenta + 1 caso de parâmetro inválido**
  (`CasosTesteFerramenta`), rodados tanto em `Main.rodarTestesFerramenta()` (console, como
  no original) quanto em `ExecutorFerramentaTest` (JUnit).
- **Retry com backoff exponencial** (`RetryingChatClient`): 500ms, 1s, 2s — mesma
  progressão, mesma distinção entre erro transitório (rede/timeout, vale retry) e erro
  terminal (ex.: modelo não encontrado, retry não resolve).
- **Os três cenários de `simularInteracao`** (`Simulador.simularInteracao`): convergência
  normal com cláusula encontrada; ferramenta sem resultado (população adulta); e
  não-convergência com escalonamento (`maxIteracoes` forçado a 1 só nesse cenário, como no
  original — em produção o limiar calibrado continua sendo 4).
- Mesmo texto de saída no console, incluindo os comentários pedagógicos (`->`).

## O que foi adaptado (e por quê)

- **Chamada ao Ollama via `java.net.http.HttpClient` + Jackson, direto na API nativa
  (`POST /api/chat`, com `tools`), sem SDK.** Mesmo padrão do projeto irmão
  [`manipulation-guardrail-prototype-java`](../../modulo-05-arquitetura-enterprise/manipulation-guardrail-prototype-java)
  deste módulo, e o mesmo usado em
  [`agent-components-demo-java`](../agent-components-demo-java). `ChatClient` é uma
  interface (`OllamaReactClient` é a implementação real) especificamente para permitir
  testar `AgenteICF` e `RetryingChatClient` com um dublê (`FakeChatClient`), sem depender
  de um Ollama local rodando durante `mvn test` — o original não tem esse problema porque
  JS/Python são interpretados e o teste da ferramenta não passa pelo modelo.
- **Histórico de mensagens vira `List<ObjectNode>` (Jackson), não um tipo de mensagem
  totalmente tipado.** O loop precisa reempurrar a mensagem bruta do assistente (com
  `tool_calls`) de volta no histórico, exatamente como veio do Ollama — igual ao original,
  que empurra `resposta.message` (um objeto dinâmico) direto na lista. Modelar isso com um
  `record` estritamente tipado por papel (system/user/assistant/tool) exigiria uma união
  de tipos mais pesada só para reproduzir o mesmo comportamento observável; `ObjectNode`
  captura a mesma flexibilidade que JS/Python têm nativamente com objetos/dicionários.
- **`ehErroTransitorio`: qualquer resposta HTTP não-200 é tratada como erro terminal**, não
  só o `status_code === 404` do original. No original, qualquer status diferente de 404
  cairia no teste genérico de mensagem (que checa por `timeout`/`econnrefused`/etc.) — que
  nunca bate para um erro puramente de status HTTP. Ou seja, o comportamento observável já
  era esse (nenhum status HTTP tratado como transitório); a versão Java só torna essa regra
  explícita em vez de dependê-la implicitamente de um teste de string que nunca casa.
- **Retorno do loop vira dois `record`s com "modo" explícito** (`ResultadoAgente` com
  `escalarParaAprovacaoHumana` sempre presente, `PassoTrilha` com `acao` como string) em
  vez dos objetos de formato variável do original (`{rascunho, iteracoes, trilha}` ou
  `{rascunho: null, escalarParaAprovacaoHumana: true, motivo, trilha}`) — mesmo conjunto de
  informações, tipado.

## `ProvedoresPagos.java` — arquivo de referência (não executado pelo fluxo principal)

Porte de referência de `provedores-pagos.js` / `provedores_pagos.py`: mostra como trocar o
Ollama local por três alternativas pagas (Claude/Anthropic, Gemini/Google, GPT/OpenAI) no
mesmo loop ReAct — o modelo é a única peça que muda; `AgenteICF`, `ExecutorFerramenta` e o
critério de parada continuam os mesmos.

- `schemaClaude`, `schemaGemini`, `schemaGpt`: **executáveis de verdade** (usam só
  Jackson), constroem o schema real que cada provedor exige para a mesma ferramenta —
  mostrando concretamente a fragmentação de formato entre provedores que motiva um
  protocolo como o MCP (Claude e Gemini usam forma achatada; GPT usa forma aninhada, igual
  ao formato nativo do Ollama já usado em `ToolSchema`).
- `chamarClaude`, `chamarGemini`, `chamarGpt`: **não executáveis** — lançam
  `UnsupportedOperationException` se chamados. Nenhum dos três SDKs pagos
  (`com.anthropic:anthropic-java`, `com.google.genai:google-genai`, `com.openai:openai-java`)
  está no `pom.xml` deste projeto (só `jackson-databind` + JUnit/AssertJ de teste) —
  adicionar um deles exigiria rede e credenciais pagas no momento do build, o que quebraria
  `mvn compile` offline para quem só quer rodar a demo local com Ollama. O Javadoc de cada
  método documenta, em bloco de código, exatamente como o corpo real ficaria com a
  dependência instalada.
- Para usar de verdade: adicione a dependência do SDK escolhido ao `pom.xml`, configure a
  variável de ambiente da chave (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY` ou `OPENAI_API_KEY`
  — ver `.env.example`) e substitua o corpo do método pela chamada real documentada no
  Javadoc.

## Como rodar

Pré-requisito: Java 17+ e Maven. O fluxo principal (`Main`) precisa do
[Ollama](https://ollama.com) rodando localmente com `ollama pull gemma4:e2b`.

```bash
# Compilar e rodar o protótipo completo (testes da ferramenta + os 3 cenários simulados)
mvn compile exec:java

# Rodar os testes automatizados (ferramenta, retry e loop ReAct com dublê — sem rede)
mvn test
```

Copie `.env.example` para `.env` se quiser customizar `OLLAMA_BASE_URL` (o valor padrão já
é `http://localhost:11434`). As variáveis de provedores pagos (`ANTHROPIC_API_KEY` etc.)
só são relevantes se você adaptar `ProvedoresPagos.java` para chamar de verdade.

## Estrutura

```
src/main/java/com/trialforge/reactagent/
  ToolSchema.java              — schema JSON da ferramenta (formato nativo Ollama/OpenAI)
  ExecutorFerramenta.java      — execução determinística da ferramenta
  CasosTesteFerramenta.java    — casos de teste compartilhados entre Main e JUnit
  ChatClient.java               — abstração da chamada "Pensamento" do loop
  ModeloResposta.java            — resposta do modelo (conteúdo ou tool_calls)
  ChamadaFerramenta.java          — uma tool_call decodificada
  ModeloHttpException.java         — erro HTTP do modelo, carrega o status code
  OllamaReactClient.java            — implementação real via HTTP nativo do Ollama
  RetryingChatClient.java            — decorador de retry com backoff exponencial
  AgenteICF.java                      — o loop ReAct em si
  Simulador.java                       — os 3 cenários de simularInteracao()
  ProvedoresPagos.java                  — referência: Claude/Gemini/GPT (não executado)
  Main.java                              — orquestra testes da ferramenta + simulação
src/test/java/com/trialforge/reactagent/
  ExecutorFerramentaTest.java   — os 6 casos + 1 inválido, parametrizados
  RetryingChatClientTest.java   — classificação de erro transitório/terminal + retry
  AgenteICFTest.java            — loop ReAct com FakeChatClient (convergência/escalonamento)
  FakeChatClient.java           — dublê de ChatClient usado nos testes acima
```
