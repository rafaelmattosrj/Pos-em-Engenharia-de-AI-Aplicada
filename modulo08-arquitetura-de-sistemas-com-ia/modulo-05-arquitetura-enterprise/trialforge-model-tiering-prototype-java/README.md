# TrialForge Model Tiering Prototype — Java

Porte Java de **`trialforge-model-tiering-prototype.js`** e **`trialforge_model_tiering_prototype.py`**
(pasta `..\`, ambos leitura obrigatória como fonte de verdade — os dois são idênticos em
comportamento; a Missão Prática pede a versão JS como entregável, a versão Python é material de
referência; esta versão Java replica os dois).

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
- **Orçamento por estudo**: `reservarOrcamento` reserva o pior caso (Tier 1 + Tier 2, se escalar)
  ANTES de qualquer chamada de modelo, e devolve a sobra depois (`liberarSobra`) — não um
  check-then-act de duas etapas, que teria uma janela de corrida sob concorrência.
- **Trilha de auditoria** (`audit-trail-tiering.jsonl`, JSON Lines, append-only): cada requisição
  processada vira um registro com as decisões determinísticas (tier usado, se escalou, se
  bloqueou por orçamento, se foi aprovada).
- **Volume concorrente** (extra, Missão Prática): dispara dezenas de requisições ao mesmo tempo
  contra o mesmo estudo pra confirmar que `reservarOrcamento` segura o orçamento sob concorrência
  real.

## O que foi mantido 1:1

- Os modelos (`nomic-embed-text`, `gemma4:e2b` = Tier 1, `gemma4:latest` = Tier 2), sem chave de
  API — Ollama local.
- `LIMIAR_CASCATA_BUSCA = LIMIAR_CASCATA_RESPOSTA = 0.75`, `CUSTO_TIER1 = 0.001`,
  `CUSTO_TIER2 = 0.01`.
- O banco de cláusulas (2 itens, RDC ANVISA 466/2012 Art. 4º e 5º) e os dois índices de embedding
  (tema para busca, texto para confiança de resposta).
- `classificarIntencao`: síntese de CSR se a pergunta contém "csr", "relatório final" ou "síntese";
  consulta de cláusula caso contrário.
- A lógica completa de `processarComCascata`: reserva de orçamento antes de qualquer chamada,
  regra fixa para CSR, cascata de dois sinais (OR) para o resto, liberação da sobra, Approval Gate
  só para CSR, registro de auditoria com os mesmos campos.
- Os 4 cenários de `main()` (Tier 1 resolve sozinho / escala por confiança baixa / CSR com regra
  fixa e aprovação / bloqueio por orçamento) e a verificação das 7 checagens sobre as últimas 4
  entradas da trilha.
- O cenário de volume concorrente (`estudo-F` com orçamento apertado para caber exatamente 2
  reservas, 5 requisições simultâneas disputando o mesmo estudo).

## O que foi adaptado (e por quê)

- **Chamada ao Ollama**: os originais usam o SDK `ollama-js`/`ollama-python`. Em Java,
  `OllamaHttpGateway` chama `POST {OLLAMA_BASE_URL}/api/chat` **com streaming** (NDJSON, uma linha
  JSON por pedaço, replicando `for await (const parte of stream)` / `for parte in stream`) e
  `POST {OLLAMA_BASE_URL}/api/embeddings`, via `java.net.http.HttpClient` + Jackson.
- **Interfaces `OllamaGateway` e `ApprovalPrompt`**: extraídas para permitir testar
  `CascadeGateway` (RAG, cascata, orçamento, Approval Gate, auditoria) com dublês determinísticos
  (`FakeOllamaGateway`, `FakeApprovalPrompt`), sem depender de um Ollama local nem de stdin
  interativo durante `mvn test`.
- **Concorrência real, não cooperativa**: o original em JS conta com o event loop de Node (nunca
  há `await` entre checar e debitar dentro de `reservarOrcamento`, então nenhuma outra chamada
  "entra no meio"); a versão em Python nem reproduz a corrida (é síncrona, sem threads). Java roda
  threads de verdade (`ExecutorService`), então `OrcamentoManager.reservarOrcamento` usa um bloco
  `synchronized` explícito por conta de estudo — sem ele, a mesma corrida aconteceria de verdade
  sob paralelismo real, não só hipoteticamente. `VolumeSimulator` roda a réplica de
  `simularVolumeConcorrente` com um `ExecutorService` de N threads.
- **`rodarTestesPuros()` não faz parte do fluxo de `main()`**: o original roda seus próprios testes
  puros (classificação de intenção, cosseno, orçamento) como parte da execução do script, porque
  não há framework de testes num script standalone. Java tem JUnit 5 — os mesmos casos (e mais)
  estão em `src/test/java`, rodados via `mvn test`, o que é mais idiomático que reimplementar um
  mini-framework de asserções dentro de `main()`.
- **Trilha de auditoria em arquivo local**: cada porte (JS/Python original, Java, Go) grava seu
  próprio `audit-trail-tiering.jsonl` dentro da respectiva pasta do projeto (aqui, no diretório de
  trabalho ao rodar `mvn exec:java`), em vez de todos disputarem o único arquivo compartilhado da
  pasta pai (`..\audit-trail-tiering.jsonl`, mantido como exemplo de dado, não sobrescrito). Isso
  evita que execuções em linguagens diferentes corrompam o histórico umas das outras.
- **Sem Spring Boot**: script CLI standalone, sem servidor HTTP — Maven puro é suficiente,
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
mvn compile
mvn exec:java
```

Volume concorrente (Missão Prática — não pede aprovação, sem CSR na mistura):

```bash
mvn exec:java -Dexec.args="--volume"
```

Configuração opcional via `.env`/variável de ambiente (ver `.env.example`):

```bash
OLLAMA_BASE_URL=http://localhost:11434   # padrão, pode omitir
```

## Como testar

```bash
mvn test
```

27 testes, cobrindo:
- Lógica pura: `IntentClassifier`, `VectorMath` (similaridade de cosseno), `OrcamentoManager`
  (`verificarOrcamento`, `reservarOrcamento`, `liberarSobra`, incluindo estudo desconhecido).
- **Concorrência real**: `OrcamentoManagerTest` dispara 20 threads reais disputando a mesma conta
  com orçamento para exatamente 2 reservas, e confirma que nunca estoura.
- `AuditTrail`: gravação append-only, timestamp, leitura de arquivo inexistente.
- `AuditTrailVerifier`: as 7 checagens sobre a trilha (sequência correta, menos de 4 entradas,
  primeiro registro incorreto, uso apenas das últimas 4 entradas).
- `CascadeGateway` de ponta a ponta, com `FakeOllamaGateway`/`FakeApprovalPrompt`: os 4 cenários
  de `main()` (Tier 1 resolve, escalação por confiança de busca, CSR com aprovação e rejeição,
  bloqueio por orçamento) e o volume concorrente com `VolumeSimulator`.
