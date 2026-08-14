# TrialForge Model Tiering Prototype — Go

Porte Go de **`trialforge-model-tiering-prototype.js`** e **`trialforge_model_tiering_prototype.py`**
(pasta `..\`, ambos leitura obrigatória como fonte de verdade — os dois são idênticos em
comportamento; a Missão Prática pede a versão JS como entregável, a versão Python é material de
referência; esta versão Go replica os dois).

## O que o protótipo faz

Cascata de Model Tiering (mecanismo do FrugalGPT — Chen, Zaharia, Zou, Stanford TMLR 2024) +
orçamento por estudo, estendendo o Gateway do Módulo 4.5: tenta o tier mais barato primeiro
(`gemma4:e2b`), só escala pro tier mais caro (`gemma4:latest`) se um de **dois sinais de
confiança** ficar abaixo do limiar (0.75):

- **Confiança de BUSCA**: o RAG achou a cláusula certa pra pergunta?
- **Confiança de RESPOSTA** — `g(pergunta, resposta)`, como no paper original: a resposta gerada
  ficou fiel à cláusula que recebeu?

Os dois sinais pegam falhas diferentes (achado real testando contra Ollama de verdade, documentado
no original): busca ruim não é sempre detectada pelo groundedness da resposta sozinho (o modelo
pode responder fielmente a uma cláusula ERRADA que recebeu), e groundedness ruim não é sempre
detectado pela confiança de busca sozinha.

Além da cascata:
- **Síntese de CSR** é regra fixa (Módulo 1.3): erro caro e irreversível, pula direto pro Tier 2 e
  aciona o **Approval Gate** (Módulo 4.4) antes de considerar a resposta oficial.
- **Orçamento por estudo**: `ReservarOrcamento` reserva o pior caso (Tier 1 + Tier 2, se escalar)
  ANTES de qualquer chamada de modelo, e devolve a sobra depois (`LiberarSobra`) — não um
  check-then-act de duas etapas, que teria uma janela de corrida sob concorrência.
- **Trilha de auditoria** (`audit-trail-tiering.jsonl`, JSON Lines, append-only): cada requisição
  processada vira um registro com as decisões determinísticas (tier usado, se escalou, se
  bloqueou por orçamento, se foi aprovada).
- **Volume concorrente** (extra, Missão Prática): dispara dezenas de requisições ao mesmo tempo
  contra o mesmo estudo pra confirmar que `ReservarOrcamento` segura o orçamento sob concorrência
  real.

## Estrutura do projeto

```
main.go                      — orquestra a demo gravada e o modo --volume
ollama/client.go              — cliente HTTP para POST /api/chat (streaming NDJSON) e /api/embeddings
tiering/clausula.go           — banco de cláusulas
tiering/vector_math.go        — similaridade de cosseno (lógica pura)
tiering/intent.go             — classificação de intenção (lógica pura)
tiering/orcamento.go          — OrcamentoManager (reserva atômica por estudo, thread-safe)
tiering/gateway.go            — interfaces Gateway e ApprovalPrompt
tiering/approval_stdin.go     — ApprovalPrompt real, lendo stdin
tiering/audit.go              — AuditTrail (append-only JSON Lines)
tiering/audit_verifier.go     — verificação das 7 checagens sobre a trilha
tiering/rag_index.go          — indexação e busca do banco de cláusulas
tiering/cascade.go            — CascadeGateway.ProcessarComCascata (orquestrador principal)
tiering/volume.go             — SimularVolumeConcorrente (goroutines reais)
```

## O que foi mantido 1:1

- Os modelos (`nomic-embed-text`, `gemma4:e2b` = Tier 1, `gemma4:latest` = Tier 2), sem chave de
  API — Ollama local.
- `LimiarCascataBusca = LimiarCascataResposta = 0.75`, `CustoTier1 = 0.001`, `CustoTier2 = 0.01`.
- O banco de cláusulas (2 itens, RDC ANVISA 466/2012 Art. 4º e 5º) e os dois índices de embedding
  (tema para busca, texto para confiança de resposta).
- `ClassificarIntencao`: síntese de CSR se a pergunta contém "csr", "relatório final" ou "síntese";
  consulta de cláusula caso contrário.
- A lógica completa de `ProcessarComCascata`: reserva de orçamento antes de qualquer chamada,
  regra fixa para CSR, cascata de dois sinais (OR) para o resto, liberação da sobra, Approval Gate
  só para CSR, registro de auditoria com os mesmos campos.
- Os 4 cenários de `main()` (Tier 1 resolve sozinho / escala por confiança baixa / CSR com regra
  fixa e aprovação / bloqueio por orçamento) e a verificação das 7 checagens sobre as últimas 4
  entradas da trilha.
- O cenário de volume concorrente (`estudo-F` com orçamento apertado para caber exatamente 2
  reservas, 5 requisições simultâneas disputando o mesmo estudo).

## O que foi adaptado (e por quê)

- **Chamada ao Ollama**: os originais usam o SDK `ollama-js`/`ollama-python`. Em Go,
  `ollama.Client.ChatStream` chama `POST {OLLAMA_BASE_URL}/api/chat` **com streaming** (NDJSON,
  uma linha JSON por pedaço, replicando `for await (const parte of stream)` /
  `for parte in stream`) e `POST {OLLAMA_BASE_URL}/api/embeddings`, via `net/http`, sem SDK.
- **Interfaces `Gateway` e `ApprovalPrompt`**: extraídas para permitir testar `CascadeGateway`
  (RAG, cascata, orçamento, Approval Gate, auditoria) com dublês determinísticos (`fakeGateway`,
  `fakeApprovalPrompt`), sem depender de um Ollama local nem de stdin interativo durante
  `go test`.
- **Concorrência real, não cooperativa**: o original em JS conta com o event loop de Node (nunca
  há `await` entre checar e debitar dentro de `reservarOrcamento`, então nenhuma outra chamada
  "entra no meio"); a versão em Python nem reproduz a corrida (é síncrona, sem threads). Go roda
  goroutines de verdade sobre múltiplas threads do SO, então `OrcamentoManager` protege cada conta
  de estudo com seu próprio `sync.Mutex`, e `ReservarOrcamento` checa e debita sob o mesmo lock —
  sem isso, a mesma corrida do comentário original aconteceria de verdade sob paralelismo real,
  não só hipoteticamente. `SimularVolumeConcorrente` dispara todas as requisições em goroutines e
  usa `sync.WaitGroup` para esperar o fim.
- **Sem `rodarTestesPuros()` dentro do fluxo principal**: o original roda seus próprios testes
  puros (classificação de intenção, cosseno, orçamento) como parte da execução do script, porque
  não há framework de testes num script standalone. Go tem o pacote `testing` — os mesmos casos
  (e mais) estão em `*_test.go`, rodados via `go test ./...`, o que é mais idiomático que
  reimplementar um mini-framework de asserções dentro de `main()`.
- **Trilha de auditoria em arquivo local**: cada porte (JS/Python original, Java, Go) grava seu
  próprio `audit-trail-tiering.jsonl` dentro da respectiva pasta do projeto (aqui, no diretório de
  trabalho ao rodar `go run .`), em vez de todos disputarem o único arquivo compartilhado da pasta
  pai (`..\audit-trail-tiering.jsonl`, mantido como exemplo de dado, não sobrescrito). Isso evita
  que execuções em linguagens diferentes corrompam o histórico umas das outras.
- **Sem framework HTTP**: script CLI standalone, sem servidor — `net/http` puro é suficiente,
  consistente com os demais protótipos irmãos do módulo.

## Como rodar

Pré-requisitos: [Ollama](https://ollama.com) instalado e rodando localmente, com os modelos
puxados:

```bash
ollama pull nomic-embed-text
ollama pull gemma4:e2b
ollama pull gemma4:latest
ollama serve
```

Demo gravada (4 chamadas sequenciais, pede aprovação humana no caso de síntese de CSR):

```bash
go build ./...
go run .
```

Volume concorrente (Missão Prática — não pede aprovação, sem CSR na mistura):

```bash
go run . --volume
```

Configuração opcional via variável de ambiente (ver `.env.example`):

```bash
OLLAMA_BASE_URL=http://localhost:11434   # padrão, pode omitir
```

## Como testar

```bash
go vet ./...
go test ./...
```

32 testes, cobrindo:
- Lógica pura: `ClassificarIntencao`, `SimilaridadeCosseno`, `OrcamentoManager`
  (`VerificarOrcamento`, `ReservarOrcamento`, `LiberarSobra`, incluindo estudo desconhecido).
- **Concorrência real**: `TestReservarOrcamento_SobConcorrenciaReal_NuncaEstouraOLimite` dispara
  20 goroutines reais disputando a mesma conta com orçamento para exatamente 2 reservas, e
  confirma que nunca estoura.
- `AuditTrail`: gravação append-only, timestamp, leitura de arquivo inexistente.
- `VerificarTrilha`: as 7 checagens sobre a trilha (sequência correta, menos de 4 entradas,
  primeiro registro incorreto, uso apenas das últimas 4 entradas).
- `CascadeGateway` de ponta a ponta, com `fakeGateway`/`fakeApprovalPrompt`: os 4 cenários de
  `main()` (Tier 1 resolve, escalação por confiança de busca, CSR com aprovação e rejeição,
  bloqueio por orçamento) e o volume concorrente com `SimularVolumeConcorrente`.
- Cliente HTTP `ollama.Client` (chat com streaming NDJSON, embeddings) com `httptest`.
