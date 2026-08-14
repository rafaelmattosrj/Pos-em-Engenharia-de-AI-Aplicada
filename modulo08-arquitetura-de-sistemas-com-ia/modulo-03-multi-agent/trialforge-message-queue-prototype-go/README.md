# TrialForge Message Queue Prototype — Go

Porte Go do protótipo de comunicação assíncrona entre os quatro agentes do
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
   contra um temporizador via `select`/`time.After`, não uma falha simulada
   por flag), retry com limite (até 3 tentativas, desiste e segue em frente
   — Disponibilidade sobre Consistência), e idempotência (o mesmo documento
   nunca duplica, não importa quantas tentativas rodem).
3. **Verificação de consistência** (Módulo 3.2, parágrafo 88): o Protocolo é
   um recurso VERSIONADO e mutável — uma emenda pode chegar enquanto ICF/CSR
   ainda trabalham com uma cópia mais antiga. Cada agente registra a versão
   que usou e a versão vigente quando ele próprio terminou.
4. **Compensação Saga de verdade** (Módulo 3.4): só o agente que ficou
   defasado é regenerado — o outro nunca é retrabalhado.

## Por que sem framework web (`net/http`, `chi`, `gin`, ...)

Este protótipo não expõe nenhuma rota HTTP — é um script de demonstração de
concorrência entre "agentes" (funções que simulam trabalho assíncrono e
publicam/consomem eventos). Um framework de rotas não teria nada para
rotear aqui; o que o original de fato demonstra é **concorrência real**
(EventEmitter do Node.js / `asyncio` do Python), e o equivalente idiomático
disso em Go são **goroutines e channels da biblioteca padrão** — não um
framework HTTP. Por isso o porte usa só `sync`, `time` e generics da
biblioteca padrão, sem nenhuma dependência externa (`go.mod` sem `require`),
igual ao original ("Sem dependência externa... só o EventEmitter nativo" /
"só asyncio, da biblioteca padrão").

## O que foi mantido 1:1

- As três constantes de tempo (`TempoProtocoloMs=300`, `TempoICFMs=450`,
  `TempoCSRMs=600`), a margem de segurança do timeout (`×1.5`), o atraso da
  emenda ética (`TempoEmendaMs=500`) e `MaxTentativas=3`.
- O fluxo Sequential → Parallel → Supervisor, na mesma ordem de passos.
- O evento rico `protocolo:pronto` (versão + critérios estruturados, cópia
  do estado no momento da publicação, não uma referência).
- O registro idempotente de documentos do ICF: a mesma chave
  (`"{versao}:icf"`) nunca cria um segundo documento, só incrementa o
  contador de tentativas.
- As duas estratégias de reação paralela: `EstrategiaAll` (replica o bug dos
  parágrafos 68-71 — falha do ICF derruba o resultado bom do CSR) e
  `EstrategiaAllSettled`, a **estratégia default** — mesmo default do
  original (`'promise.allSettled'` / `"gather_com_excecoes"`), obtido de
  graça porque é o valor zero do tipo `Estrategia` em Go.
- A verificação de consistência por agente (versão usada vs. versão ao
  concluir) e a compensação Saga que regenera só quem ficou defasado.
- Os mesmos 9 cenários de teste do `rodarTestes()`/`rodar_testes()` original,
  um a um, em `trialforge/fluxo_test.go`.

## O que foi adaptado (e por quê)

| Original (JS/Python) | Go | Por quê |
|---|---|---|
| `EventEmitter` nativo (JS) / classe `Barramento` sobre `asyncio` (Python) | `Barramento` própria, com `Once`/`Emit`, despachando o ouvinte via `go ouvinte(dado)` | Go não tem um event loop; o equivalente idiomático de "publica e não bloqueia" é disparar uma goroutine no `Emit`, protegida por `sync.Mutex` para o registro do ouvinte. |
| `Promise.race` contra `setTimeout` (JS) / `asyncio.wait_for` (Python) | `select` entre um canal de resultado e `time.After(timeoutMs)` (`comTimeout`/`iniciarComTimeout`, com generics) | É a forma idiomática de "corrida contra um timer" em Go — a mesma semântica observável (quem chega primeiro "ganha"), sem `context.Context` porque o original também não cancela o agente travado, só abandona seu resultado. |
| `Promise.all`/`Promise.allSettled` (JS) e `asyncio.gather`/`gather(return_exceptions=True)` (Python) | Tipo `Estrategia` (`EstrategiaAll`, `EstrategiaAllSettled`) com a lógica equivalente escrita a mão sobre dois canais bufferizados iniciados concorrentemente | Go não tem combinador de duas Promises pronto; ICF e CSR são iniciados como duas corridas concorrentes (`iniciarComTimeout` não bloqueia) e cada branch decide o que preservar em caso de falha, preservando o comportamento observável. |
| Objeto de opções posicional/nomeado (`{estrategia, forcarFalhaICF, ...}`) | `struct OpcoesFluxo` — o valor zero (`OpcoesFluxo{}`) já reproduz "nenhuma falha simulada, estratégia allSettled", igual ao default do original | Go não tem parâmetros nomeados/default nem sobrecarga; o valor zero de uma struct é o equivalente idiomático mais direto para "opções todas desligadas por padrão". |
| Objeto retornado com *spread* (`{...resultado, tentativas}`) | `ResultadoICFBase` embutido (embedding) em `ResultadoICF`, mais o campo `Tentativas` | Composição por embedding é o equivalente idiomático de Go ao spread de objetos: `ResultadoICF` "herda" os campos de `ResultadoICFBase` e adiciona o contador. |
| Suíte de testes manual (`rodarTestes()`, contador de passou/total, `console.log`/`print`) | Funções `Test*` em `trialforge/fluxo_test.go` e `trialforge/state_test.go`, pacote `testing` padrão | É o padrão de testes idiomático em Go; substitui o contador manual por `go test`, que já reporta passou/falhou por caso. O `main.go` roda só a demonstração narrada (os 5 cenários), sem duplicar a suíte de testes dentro dele. |

Todo o resto — nomes de campos/funções (adaptados para `PascalCase`
exportado, convenção Go), mensagens de erro, estrutura dos dados, os 9
cenários de teste, os 5 cenários da demonstração narrada — é fiel ao
original.

## Estrutura

```
main.go                        demonstração narrada (5 cenários) — `go run .`
trialforge/
  bus.go                        Barramento (fila de mensagens pub/sub Once/Emit)
  state.go                      EstadoProtocolo (versionado, mutável) + DadoProtocolo + RevisaoHistorico
  registro.go                   OrdemDeExecucao e DocumentosGerados (thread-safe)
  opcoes.go                     OpcoesFluxo, Estrategia
  agentes.go                    AgenteProtocolo / AgenteICF / AgenteCSR + constantes de tempo/timeout
  reacao.go                     Reacao + InscreverReacaoParalela (Parallel: EstrategiaAll vs AllSettled)
  supervisor.go                 retry com limite (CAP), verificação de consistência, compensação Saga
  fluxo.go                      RodarFluxoTrialForge — orquestração do fluxo completo
  timeout.go                    ErroDeTimeout + comTimeout/iniciarComTimeout (corrida real via select)
  fluxo_test.go                 os 9 cenários do rodar_testes() original + 1 teste de escopo isolado
  state_test.go                 unitário do estado versionado
```

## Como rodar

```bash
go build ./...    # compila
go vet ./...      # análise estática
go test ./...     # roda os 13 testes (os 9 cenários do original + extras)
go run .           # roda a demonstração narrada (5 cenários), equivalente a `python trialforge_message_queue_prototype.py`
```

Requer Go 1.21+ (usa generics). Os testes levam ~15s no total porque os
tempos de simulação (300/450/600ms por agente, multiplicados pelos cenários
de retry/timeout) são executados de verdade — igual ao original. `go test
-race` não roda neste ambiente por falta de um compilador C (cgo), mas o
acesso a estado compartilhado (`EstadoProtocolo`, `OrdemDeExecucao`,
`DocumentosGerados`) é protegido por `sync.Mutex`/`sync.Map`-like locking em
todos os pontos de escrita/leitura entre goroutines.
