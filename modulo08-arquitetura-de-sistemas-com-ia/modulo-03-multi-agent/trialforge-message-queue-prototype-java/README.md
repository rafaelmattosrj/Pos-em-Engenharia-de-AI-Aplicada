# TrialForge Message Queue Prototype — Java

Porte Java do protótipo de comunicação assíncrona entre os quatro agentes do
TrialForge (Protocolo, ICF, CSR e Supervisor), do módulo 08 —
`modulo-03-multi-agent`.

**Fonte de verdade (portado de, ambos lidos por completo):**
- `../trialforge-message-queue-prototype.js` (original pedido pela Missão Prática #03)
- `../trialforge_message_queue_prototype.py` (versão de referência em Python, comportamento idêntico)

Não chama nenhum LLM — nem o original chama. O objetivo é demonstrar, com
mecanismos reais (não narrados/simulados por flag), tudo que o módulo 3
ensina sobre comunicação entre agentes:

1. **Sequential + Parallel**: o Agente Protocolo roda primeiro; ICF e CSR
   reagem ao mesmo evento `protocolo:pronto` em paralelo.
2. **Teorema CAP completo** (Módulo 3.4): timeout explícito (corrida real
   contra um temporizador, não uma falha simulada por flag), retry com
   limite (até 3 tentativas, desiste e segue em frente — Disponibilidade
   sobre Consistência), e idempotência (o mesmo documento nunca duplica,
   não importa quantas tentativas rodem).
3. **Verificação de consistência** (Módulo 3.2, parágrafo 88): o Protocolo é
   um recurso VERSIONADO e mutável — uma emenda pode chegar enquanto ICF/CSR
   ainda trabalham com uma cópia mais antiga. Cada agente registra a versão
   que usou e a versão vigente quando ele próprio terminou.
4. **Compensação Saga de verdade** (Módulo 3.4): só o agente que ficou
   defasado é regenerado — o outro nunca é retrabalhado.

## Por que sem Spring Boot

Este protótipo não expõe nenhum endpoint HTTP e não tem nenhum grafo de
dependências que justifique injeção de dependência — são poucas classes
pequenas (agentes, estado, barramento, supervisor) chamadas diretamente por
uma única classe orquestradora. Adicionar Spring Boot aqui significaria
carregar um container de DI, auto-configuração e um ciclo de vida de
aplicação inteiros só para rodar `main()` uma vez — sem nenhum ganho de
legibilidade ou testabilidade em troca. Por isso o porte usa **Maven puro**,
igual aos protótipos CLI do módulo 01. Todo o poder de concorrência real
(a parte que de fato precisava de uma ferramenta madura) vem do
`ExecutorService`/`Future` da JDK, que é o equivalente idiomático do
`EventEmitter`/`asyncio` do original.

## O que foi mantido 1:1

- As três constantes de tempo (`TEMPO_PROTOCOLO_MS=300`, `TEMPO_ICF_MS=450`,
  `TEMPO_CSR_MS=600`), a margem de segurança do timeout (`×1.5`), o atraso da
  emenda ética (`500ms`) e `MAX_TENTATIVAS=3`.
- O fluxo Sequential → Parallel → Supervisor, na mesma ordem de passos.
- O evento rico `protocolo:pronto` (versão + critérios estruturados, cópia
  do estado no momento da publicação, não uma referência).
- O registro idempotente de documentos do ICF: a mesma chave
  (`"{versao}:icf"`) nunca cria um segundo documento, só incrementa o
  contador de tentativas.
- As duas estratégias de reação paralela: `PROMISE_ALL` (replica o bug dos
  parágrafos 68-71 — falha do ICF derruba o resultado bom do CSR) e
  `PROMISE_ALL_SETTLED` (a correção — cada resultado é preservado
  independente do outro).
- A verificação de consistência por agente (versão usada vs. versão ao
  concluir) e a compensação Saga que regenera só quem ficou defasado.
- Os mesmos 9 cenários de teste do `rodarTestes()`/`rodar_testes()` original,
  um a um.

## O que foi adaptado (e por quê)

| Original (JS/Python) | Java | Por quê |
|---|---|---|
| `EventEmitter` nativo (JS) / classe `Barramento` sobre `asyncio` (Python) | `Barramento` própria, com `once`/`emit`, despachando o ouvinte via `ExecutorService` | Java não tem um event loop single-thread nativo; o equivalente idiomático de "não-bloqueante" com threads reais é submeter o ouvinte ao executor em vez de bloquear quem chamou `emit`. |
| `Promise.race` contra `setTimeout` (JS) / `asyncio.wait_for` (Python) | `Future#get(timeout, TimeUnit)` sobre um `Callable` submetido ao `ExecutorService` | É o mecanismo real de timeout mais idiomático da JDK — mesma semântica (corrida real contra um prazo), sem reimplementar manualmente `Promise.race`. |
| Objeto de opções posicional (`{estrategia, forcarFalhaICF, ...}`) | Classe imutável `OpcoesFluxo` com métodos fluentes (`comEstrategia`, `forcarFalhaICF`, ...) | Java não tem parâmetros nomeados/default; o padrão builder imutável é o equivalente idiomático mais próximo, sem exigir uma cadeia de overloads. |
| `Promise.all`/`Promise.allSettled` (JS) e `asyncio.gather`/`gather(return_exceptions=True)` (Python) | Enum `Estrategia` (`PROMISE_ALL`, `PROMISE_ALL_SETTLED`) com a lógica equivalente escrita a mão sobre `Future#get` | A JDK não tem um combinador de duas Promises pronto com a mesma semântica; a lógica (iniciar os dois concorrentemente, então decidir o que preservar em caso de falha) foi replicada manualmente, preservando o comportamento observável. |
| Objeto retornado com *spread* (`{...resultado, tentativas}`) | `ResultadoICFBase` + `ResultadoICF` (dois `record`s, o segundo "estende" o primeiro via `comTentativas(int)`) | Java não tem spread de objetos; a divisão em dois `record`s imutáveis reproduz exatamente a mesma composição em dois passos (grava o resultado sem tentativas, depois monta o valor final com o contador). |
| Suíte de testes manual (`rodarTestes()`, contador de passou/total, `console.log`) | Classes JUnit 5 (`TrialForgeFluxoTest`, `EstadoProtocoloTest`) com AssertJ | É o padrão de testes já usado nos demais projetos Java do repositório; substitui o contador manual por relatórios estruturados (`mvn test`). |

Todo o resto — nomes de método, mensagens de erro, estrutura dos dados, os
9 cenários de teste, os 5 cenários da demonstração narrada — é fiel ao
original.

## Estrutura

```
src/main/java/com/trialforge/messagequeue/
  Barramento.java            fila de mensagens (pub/sub once/emit)
  EstadoProtocolo.java       protocolo versionado e mutável + histórico
  Agentes.java                agenteProtocolo / agenteICF / agenteCSR / registrarDocumento
  Supervisor.java             decisão+execução de retry (CAP) e consistência+compensação (Saga)
  TrialForgeFluxo.java        orquestração do fluxo completo + comTimeout
  OpcoesFluxo.java, Estrategia.java, DadoProtocolo.java, ResultadoICF(Base).java,
  ResultadoCSR.java, DocumentoRegistro.java, Reacao.java, Verificacao.java,
  ResultadoRetry.java, CompensacaoResultado.java, ResultadoFluxo.java, RevisaoHistorico.java
  ErroDeTimeoutException.java
  Main.java                   demonstração narrada (5 cenários)
src/test/java/com/trialforge/messagequeue/
  TrialForgeFluxoTest.java    os 9 cenários do rodarTestes() original + 1 teste de escopo isolado
  EstadoProtocoloTest.java    unitário do estado versionado
```

## Como rodar

```bash
mvn compile          # compila
mvn test             # roda os 13 testes JUnit 5 (os 9 cenários do original + extras)
mvn exec:java         # roda a demonstração narrada (5 cenários), equivalente a `node trialforge-message-queue-prototype.js`
```

Requer Java 17+ e Maven. Os testes levam ~15s no total porque os tempos de
simulação (300/450/600ms por agente, multiplicados pelos cenários de
retry/timeout) são executados de verdade — igual ao original.

> Nota: se o console do Windows não estiver em UTF-8, a saída de `mvn
> exec:java` pode exibir acentos incorretamente (mojibake) — isso é uma
> limitação do codepage do terminal (`chcp 65001` resolve), não um bug no
> código; os valores retornados pela aplicação estão corretos (conferido
> pelos testes JUnit/AssertJ, que não dependem do console).
