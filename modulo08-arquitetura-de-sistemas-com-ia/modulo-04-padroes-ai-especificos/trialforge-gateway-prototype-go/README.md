# TrialForge Gateway Prototype — Go

Porte em Go de [`trialforge-gateway-prototype.js`](../trialforge-gateway-prototype.js) (fonte de verdade, conforme a ementa da Missão Prática #04 — a versão [`trialforge_gateway_prototype.py`](../trialforge_gateway_prototype.py) é referência idêntica e também foi usada pra conferir comportamento). Artefato de demo do Módulo 4.5, combinando os 4 grupos de padrão do Módulo 4:

- **4.1 — RAG completo**: Multi-Index (um índice por domínio: ICF, Protocolo, CSR) + Hybrid Search (BM25 léxico + embedding denso, fundidos por Reciprocal Rank Fusion) + Agentic RAG (até 3 iterações ampliando a estratégia de busca).
- **4.2 — Intent-Based Routing + Model Router**: classificador determinístico decide o índice e o modelo (barato para rotina, caro para síntese de CSR).
- **4.3 — Semantic Cache + Response Streaming**: cache por similaridade de cosseno (perguntas de rotina); geração token a token.
- **4.4 — Confidence Threshold + Approval Gate + Audit Trail**: escalonamento para aprovação humana quando a confiança do RAG é baixa ou quando é síntese de CSR (sempre); trilha append-only em `audit-trail.jsonl`.

## Por que sem framework web

O original não expõe nenhum servidor HTTP — não há Express/Fastify nem qualquer `listen()`/bind de porta. É um script que roda uma simulação de 5 requisições em sequência contra um Ollama local e termina. Por isso este porte segue a convenção de **CLI/demo standalone**: módulo Go simples, sem `chi`/`gin`/`echo`. Não há múltiplas rotas HTTP reais para justificar um roteador, e a orquestração aqui (RAG → cache → modelo → gate → auditoria) é sequencial dentro de um único `Processor`, não uma superfície de API — um framework web não deixaria esse fluxo mais idiomático, só adicionaria uma dependência sem função.

## Estrutura

```
main.go            — monta o Processor (client Ollama real, cache vazio, ApprovalGate sobre stdin/stdout)
                      e roda o roteiro de 5 perguntas + verificação da trilha, igual ao main() dos originais
dotenv.go           — mesmo loader minimalista usado em ollama-local-llm-chat-go
ollama/
  client.go         — cliente HTTP para a API NATIVA do Ollama (/api/embeddings, /api/chat) —
                      diferente de ollama-local-llm-chat-go, que fala com a API OpenAI-compatible
                      (/v1) só de chat; aqui precisamos de embeddings reais também
  retry.go          — comTimeout + comRetry genéricos (Go generics), mesma receita do Módulo 3.5
gateway/
  types.go, indices.go     — Clausula, Message, os 3 índices (ICF/Protocolo/CSR) e o roteamento intenção→índice
  intent.go                — ClassificarIntencao (Intent-Based Routing)
  tokenize.go               — Tokenizar (NFD + remoção de acentos + tokenização)
  bm25.go, rank.go, cosine.go — BM25, ordenação/RRF, similaridade de cosseno
  rag.go                    — PrepararIndices, BuscarClausulaHibrida, BuscarEmTodosIndices, BuscarClausulaAgentica
  cache.go                  — SemanticCache
  approval.go               — ApprovalGate (Approval Gate)
  audit.go                  — AuditTrail (registrar + VerificarTrilhaAuditoria)
  processor.go              — Processor.ProcessarRequisicao — o gateway propriamente dito
```

## O que foi mantido 1:1

- Os 3 índices (ICF, Protocolo, CSR) com o mesmo texto/fonte regulatória das cláusulas.
- Fórmula do BM25 (k1=1.5, b=0.75), Reciprocal Rank Fusion (k=60) e o algoritmo do Agentic RAG (até 3 iterações: tema → texto → todos os índices).
- Limiares `LIMIAR_CACHE=0.75` e `LIMIAR_CONFIANCA=0.7`.
- Regras do Model Router (barato para rotina, caro para síntese de CSR) e do Approval Gate (síntese de CSR sempre passa pelo gate; demais só quando a confiança do RAG fica abaixo do limiar).
- O mesmo roteiro de demo: 5 perguntas em sequência (rotina → paráfrase que bate no cache → síntese de CSR → tema fora dos índices que esgota o Agentic RAG → critério de protocolo que converge na 1ª iteração) e a verificação pós-execução da trilha de auditoria (`VerificarTrilhaAuditoria`), incluindo o cuidado de excluir registros `aguardando_aprovacao` antes de olhar as "últimas 5" requisições concluídas.
- Retry com limite (3 tentativas) + timeout (20s) nas duas chamadas de rede (embedding e geração) — mesma receita do Módulo 3.5.
- Trilha de auditoria append-only em JSONL, um registro por decisão, timestamp incluído.

## O que foi adaptado (e por quê)

- **API do Ollama usada diretamente, não um SDK**: os originais usam os pacotes `ollama` do npm/PyPI. Não existe SDK oficial Go para o Ollama no mesmo nível de maturidade, então `ollama/client.go` fala HTTP diretamente com a API nativa (`/api/embeddings`, `/api/chat` com `stream:true` em NDJSON) — os mesmos dois endpoints que os pacotes oficiais chamam por baixo dos panos.
- **Retry só protege o estabelecimento da chamada, não a iteração dos chunks**: replica fielmente o alcance do `comRetry(() => ollama.chat(...))` do JS (que também só protege a promise inicial, não o `for await` de consumo) — um chunk corrompido no meio do stream propaga erro, sem retry por chunk, igual aos dois originais.
- **Timeout "desiste de esperar, não cancela"**: mesma limitação honesta documentada no comentário da versão Python — Go não tem como matar uma goroutine à força; `comTimeout` usa `select` com `time.After` e abandona a espera, a chamada em segundo plano pode continuar rodando.
- **Approval Gate sem detecção de TTY**: os originais em JS precisam de um workaround real (ver comentário em `approval.go`) porque o `readline` do Node fecha sozinho com stdin não-interativo lido aos poucos. Em Go, um único `*bufio.Reader` criado uma vez sobre `os.Stdin` (nunca recriado) já resolve isso sem precisar checar `isTTY` — o mesmo motivo pelo qual a versão Python (que usa `input()` simples) também não precisou do workaround.
- **`Embedder`, `ChatStreamer` e `Approver` como interfaces**: os originais não têm essa separação explícita (é JS/Python dinâmico); em Go, isolar os 3 contratos de I/O permite testar `Processor.ProcessarRequisicao` — a peça mais importante do gateway — com dublês determinísticos, sem precisar de um Ollama real nem de stdin real.
- **`golang.org/x/text/unicode/norm` para NFD**: a stdlib do Go não tem normalização Unicode embutida; é a forma idiomática de replicar o `.normalize('NFD')` do JS / `unicodedata.normalize('NFD', ...)` do Python.

## Como executar a demo completa

Pré-requisito: [Ollama](https://ollama.com) rodando localmente com os 3 modelos baixados:

```bash
ollama pull nomic-embed-text
ollama pull gemma4:e2b
ollama pull gemma4
```

```bash
cp .env.example .env   # opcional, os defaults já apontam pro Ollama local
go run .
```

A demo roda 5 perguntas em sequência, imprime o raciocínio de cada padrão (roteamento, cache, modelo, RAG agêntico, gate) e termina verificando a trilha de auditoria gravada em `audit-trail.jsonl` (mesmo arquivo append-only usado pelos originais).

## Testes

```bash
go build ./...
go vet ./...
go test ./...
```

Cobrem os mesmos cenários que os testes puros dos originais (classificação de intenção, cosseno, BM25, RRF) mais os que só fazem sentido com a estrutura em pacotes do Go: `ollama.Client` contra um `httptest.Server` (embedding e chat streaming, sucesso/retry/timeout/erro), o Approval Gate com um `io.Reader` fake (aprovação, rejeição, EOF, múltiplas chamadas em sequência — o mesmo cenário de entrada não-interativa que motivou a correção nos originais), a trilha de auditoria (gravação append-only e as mesmas checagens de `VerificarTrilhaAuditoria`), e os 5 caminhos do roteiro de demo executados via `Processor.ProcessarRequisicao` com embeddings/chat/approval controlados (sem rede).
