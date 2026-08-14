# Model Eval Gate Prototype — Java

Porte Java de **`model-eval-gate-prototype.js`** e **`model_eval_gate_prototype.py`**
(pasta `..\`, ambos leitura obrigatória como fonte de verdade — os dois são idênticos em
comportamento, esta versão Java replica os dois).

## O que o protótipo faz

Eval Gate de modelo: antes de promover um candidato para processar tráfego real, roda ele contra
um **golden set** (perguntas com cláusula regulatória esperada já conhecida) e só promove se o
score médio de fidelidade (groundedness) não regredir contra o baseline atual além de uma
**tolerância** (2%). Mesmo espírito do canary do KServe (Módulo 5.1) — só que o canary avisa
*depois* que a versão nova já está em produção; o eval gate avisa *antes* de promover.

O score usa o mesmo mecanismo de groundedness do Módulo 5.4: embedding da resposta gerada vs.
embedding da cláusula esperada (similaridade de cosseno).

O demo roda 2 cenários:
1. **Candidato real** (`gemma4:e2b-mlx`, variante MLX do baseline `gemma4:e2b`) — caso limite,
   dois modelos reais e parecidos.
2. **Candidato regredido por bug de config**, não por modelo pior: o mesmo baseline, mas sem a
   cláusula no contexto (simula RAG/prompt template quebrado) — regressão determinística, sempre
   detectada pelo gate.

## O que foi mantido 1:1

- Os modelos (`nomic-embed-text` para embeddings, `gemma4:e2b` como baseline, `gemma4:e2b-mlx`
  como candidato), sem chave de API — Ollama local.
- O golden set: as 3 mesmas perguntas, cláusulas e fontes regulatórias (RDC ANVISA 466/2012 e o
  protocolo TrialForge).
- `TOLERANCIA_REGRESSAO = 0.02`.
- A matemática de similaridade de cosseno (`similaridadeCosseno`).
- A lógica de decisão: promove se `scoreCandidato - scoreBaseline >= -tolerancia`.
- Os prompts de sistema e de usuário enviados ao modelo em `gerarResposta` (com cláusula) e
  `gerarRespostaSemContexto` (sem cláusula, simulando a regressão de config).
- Os dois cenários de demonstração e as mensagens de decisão (`PROMOVE`/`BLOQUEIA`).

## O que foi adaptado (e por quê)

- **Chamada ao Ollama**: os originais usam o SDK `ollama-js`/`ollama-python`. Em Java,
  `OllamaHttpGateway` chama `POST {OLLAMA_BASE_URL}/api/chat` (sem streaming, já que o original
  não usa streaming aqui) e `POST {OLLAMA_BASE_URL}/api/embeddings` diretamente via
  `java.net.http.HttpClient` + Jackson, sem SDK/dependência extra além do parser JSON.
- **Interface `OllamaGateway`**: extraída para permitir testar `EvalGate` (scoring, decisão de
  promoção) com um dublê determinístico (`FakeOllamaGateway`), sem depender de um Ollama local
  rodando durante `mvn test`. O original não tem testes automatizados no sentido JUnit; os testes
  Java cobrem a matemática de cosseno, a lógica de decisão do gate (incluindo o caso-limite exato
  no valor da tolerância) e o fluxo ponta a ponta de avaliação com scores controlados.
- **Sem Spring Boot**: script CLI standalone, sem servidor HTTP — Maven puro é suficiente,
  consistente com os demais protótipos irmãos do módulo.

## Como rodar

Pré-requisitos: [Ollama](https://ollama.com) instalado e rodando localmente, com os modelos
puxados:

```bash
ollama pull nomic-embed-text
ollama pull gemma4:e2b
ollama pull gemma4:e2b-mlx
ollama serve
```

Compilar e rodar:

```bash
mvn compile
mvn exec:java
```

Configuração opcional via `.env`/variável de ambiente (ver `.env.example`):

```bash
OLLAMA_BASE_URL=http://localhost:11434   # padrão, pode omitir
```

## Como testar

```bash
mvn test
```

Os testes usam `FakeOllamaGateway` para simular chat e embeddings sem depender de um Ollama
local — cobrem: similaridade de cosseno (vetores idênticos, ortogonais, opostos), a decisão de
promoção (`decidirPromocao`, incluindo o limite exato da tolerância), e os fluxos
`avaliarCandidato`/`avaliarCandidatoSemContexto` de ponta a ponta com scores controlados.
