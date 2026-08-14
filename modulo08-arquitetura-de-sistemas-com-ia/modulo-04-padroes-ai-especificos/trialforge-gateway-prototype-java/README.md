# TrialForge Gateway Prototype — Java

Porte em Java de [`trialforge-gateway-prototype.js`](../trialforge-gateway-prototype.js) (fonte de verdade, conforme a ementa da Missão Prática #04 — a versão [`trialforge_gateway_prototype.py`](../trialforge_gateway_prototype.py) é referência idêntica e também foi usada pra conferir comportamento). Artefato de demo do Módulo 4.5, combinando os 4 grupos de padrão do Módulo 4:

- **4.1 — RAG completo**: Multi-Index (um índice por domínio: ICF, Protocolo, CSR) + Hybrid Search (BM25 léxico + embedding denso, fundidos por Reciprocal Rank Fusion) + Agentic RAG (até 3 iterações ampliando a estratégia de busca).
- **4.2 — Intent-Based Routing + Model Router**: classificador determinístico decide o índice e o modelo (barato para rotina, caro para síntese de CSR).
- **4.3 — Semantic Cache + Response Streaming**: cache por similaridade de cosseno (perguntas de rotina); geração token a token.
- **4.4 — Confidence Threshold + Approval Gate + Audit Trail**: escalonamento para aprovação humana quando a confiança do RAG é baixa ou quando é síntese de CSR (sempre); trilha append-only em `audit-trail.jsonl`.

## Por que sem Spring Boot

O original não expõe nenhum servidor HTTP — não há Express/Fastify nem qualquer `listen()`/bind de porta. É um script que roda uma simulação de 5 requisições em sequência contra um Ollama local e termina. Por isso este porte segue a convenção de **CLI/demo standalone** já usada em `manipulation-guardrail-prototype-java` (mesmo Módulo 8): Maven puro, sem `spring-boot-starter-web`. Não há rotas HTTP reais que justifiquem Spring — a orquestração aqui (RAG → cache → modelo → gate → auditoria) é sequencial dentro de um único `Processor`, não uma superfície de API.

## Estrutura

```
src/main/java/com/trialforge/gateway/
  Main.java              — monta o Processor (OllamaClient real, cache vazio, ApprovalGate sobre
                            System.in/System.out) e roda o roteiro de 5 perguntas + verificação
                            da trilha, igual ao main() dos originais
  Clausula.java, ChatMessage.java, Indices.java  — modelo de dados + os 3 índices (ICF/Protocolo/CSR)
  IntentClassifier.java   — classificarIntencao (Intent-Based Routing)
  Tokenizer.java           — tokenizar (NFD + remoção de acentos + tokenização)
  Bm25.java, RankUtils.java, CosineSimilarity.java — BM25, ordenação/RRF, similaridade de cosseno
  IndicePreparado.java, ResultadoBusca.java, RagSearch.java, AgenticRag.java — RAG completo
  SemanticCache.java       — cache semântico
  ApprovalGate.java, Approver.java — Approval Gate
  AuditTrail.java           — registrar + verificarTrilhaAuditoria
  Embedder.java, ChatStreamer.java, DemoLogger.java — contratos de I/O do Processor (testáveis com dublês)
  Processor.java            — processarRequisicao — o gateway propriamente dito
  ollama/
    OllamaClient.java        — cliente HTTP (java.net.http.HttpClient + Jackson) para a API NATIVA
                                do Ollama (/api/embeddings, /api/chat) — diferente de
                                manipulation-guardrail-prototype-java, que só usa /api/chat sem
                                streaming; aqui precisamos de embeddings reais + streaming NDJSON
    RetrySupport.java, OllamaTimeoutException.java — retry com limite + timeout (Módulo 3.5)
```

## O que foi mantido 1:1

- Os 3 índices (ICF, Protocolo, CSR) com o mesmo texto/fonte regulatória das cláusulas.
- Fórmula do BM25 (k1=1.5, b=0.75), Reciprocal Rank Fusion (k=60) e o algoritmo do Agentic RAG (até 3 iterações: tema → texto → todos os índices).
- Limiares `LIMIAR_CACHE=0.75` e `LIMIAR_CONFIANCA=0.7`.
- Regras do Model Router (barato para rotina, caro para síntese de CSR) e do Approval Gate (síntese de CSR sempre passa pelo gate; demais só quando a confiança do RAG fica abaixo do limiar).
- O mesmo roteiro de demo: 5 perguntas em sequência (rotina → paráfrase que bate no cache → síntese de CSR → tema fora dos índices que esgota o Agentic RAG → critério de protocolo que converge na 1ª iteração) e a verificação pós-execução da trilha de auditoria (`AuditTrail.verificarTrilhaAuditoria`), incluindo o cuidado de excluir registros `aguardando_aprovacao` antes de olhar as "últimas 5" requisições concluídas.
- Retry com limite (3 tentativas) + timeout (20s) nas duas chamadas de rede (embedding e geração) — mesma receita do Módulo 3.5.
- Trilha de auditoria append-only em JSONL, um registro por decisão, timestamp incluído.

## O que foi adaptado (e por quê)

- **API do Ollama chamada diretamente via `java.net.http.HttpClient` + Jackson**: os originais usam os pacotes `ollama` do npm/PyPI. Não existe um SDK Java oficial equivalente, então `OllamaClient` fala HTTP diretamente com a API nativa (`/api/embeddings`, `/api/chat` com `stream:true` em NDJSON, lido via `HttpResponse.BodyHandlers.ofLines()`) — os mesmos dois endpoints que os pacotes oficiais chamam por baixo dos panos. Mesmo padrão já usado em `OllamaClassifierClient` de `manipulation-guardrail-prototype-java`, estendido pra embeddings e streaming.
- **Retry só protege o estabelecimento da chamada, não a iteração dos chunks**: replica fielmente o alcance do `comRetry(() => ollama.chat(...))` do JS (que também só protege a promise inicial, não o `for await` de consumo) — uma linha corrompida no meio do stream propaga erro, sem retry por chunk, igual aos dois originais.
- **Timeout via `ExecutorService` + `Future.get(timeout)`, não `Thread.interrupt()`**: mesma limitação honesta documentada no comentário da versão Python (que usa `ThreadPoolExecutor(...).result(timeout=...)`) — Java não mata uma thread à força de forma segura; `RetrySupport.comRetry` desiste de ESPERAR a resposta, a chamada em segundo plano pode continuar rodando até o Ollama eventualmente responder.
- **Approval Gate sem detecção de TTY**: os originais em JS precisam de um workaround real (ver comentário em `ApprovalGate.java`) porque o `readline` do Node fecha sozinho com stdin não-interativo lido aos poucos. Em Java, um único `BufferedReader` criado uma vez sobre o `InputStream` (nunca recriado) já resolve isso sem precisar checar TTY — o mesmo motivo pelo qual a versão Python (que usa `input()` simples) também não precisou do workaround.
- **`Embedder`, `ChatStreamer` e `Approver` como interfaces funcionais**: os originais não têm essa separação explícita (é JS/Python dinâmico); em Java, isolar os 3 contratos de I/O permite testar `Processor.processarRequisicao` — a peça mais importante do gateway — com dublês determinísticos (`FakeEmbedder`, `FakeChatStreamer`, `FakeApprover`), sem precisar de um Ollama real nem de stdin real.
- **Sem dotenv**: seguindo a convenção do sibling `manipulation-guardrail-prototype-java` (mesmo módulo), as variáveis de ambiente são lidas via `System.getenv()` puro, sem biblioteca de `.env`. O `.env.example` documenta os defaults, mas precisa ser exportado manualmente no shell se você quiser sobrescrevê-los.

## Como executar a demo completa

Pré-requisito: [Ollama](https://ollama.com) rodando localmente com os 3 modelos baixados:

```bash
ollama pull nomic-embed-text
ollama pull gemma4:e2b
ollama pull gemma4
```

```bash
mvn compile exec:java
```

A demo roda 5 perguntas em sequência, imprime o raciocínio de cada padrão (roteamento, cache, modelo, RAG agêntico, gate) e termina verificando a trilha de auditoria gravada em `audit-trail.jsonl` (mesmo arquivo append-only usado pelos originais).

## Testes

```bash
mvn compile
mvn test
```

Cobrem os mesmos cenários que os testes puros dos originais (classificação de intenção, cosseno, BM25, RRF) mais os que só fazem sentido com a estrutura em classes do Java: `OllamaClient` contra um `com.sun.net.httpserver.HttpServer` local (embedding e chat streaming, sucesso/retry/timeout/erro), o Approval Gate com um `ByteArrayInputStream` fake (aprovação, rejeição, EOF, múltiplas chamadas em sequência — o mesmo cenário de entrada não-interativa que motivou a correção nos originais), a trilha de auditoria (gravação append-only e as mesmas checagens de `verificarTrilhaAuditoria`), e os 5 caminhos do roteiro de demo executados via `Processor.processarRequisicao` com embeddings/chat/approval controlados (sem rede).
