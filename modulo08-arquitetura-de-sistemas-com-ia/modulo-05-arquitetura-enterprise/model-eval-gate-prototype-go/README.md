# Model Eval Gate Prototype — Go

Porte Go de **`model-eval-gate-prototype.js`** e **`model_eval_gate_prototype.py`**
(pasta `..\`, ambos leitura obrigatória como fonte de verdade — os dois são idênticos em
comportamento, esta versão Go replica os dois).

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

## Estrutura do projeto

```
main.go                — orquestra os 2 cenários de demo (equivalente a main() no original)
ollama/client.go        — cliente HTTP mínimo para POST /api/chat e POST /api/embeddings
evalgate/evalgate.go    — golden set, similaridade de cosseno, decisão de promoção, avaliação
```

## O que foi mantido 1:1

- Os modelos (`nomic-embed-text` para embeddings, `gemma4:e2b` como baseline, `gemma4:e2b-mlx`
  como candidato), sem chave de API — Ollama local.
- O golden set: as 3 mesmas perguntas, cláusulas e fontes regulatórias (RDC ANVISA 466/2012 e o
  protocolo TrialForge).
- `ToleranciaRegressao = 0.02`.
- A matemática de similaridade de cosseno (`SimilaridadeCosseno`).
- A lógica de decisão: promove se `scoreCandidato - scoreBaseline >= -tolerancia`.
- Os prompts de sistema e de usuário enviados ao modelo em `GerarResposta` (com cláusula) e
  `GerarRespostaSemContexto` (sem cláusula, simulando a regressão de config).
- Os dois cenários de demonstração e as mensagens de decisão (`PROMOVE`/`BLOQUEIA`).

## O que foi adaptado (e por quê)

- **Chamada ao Ollama**: os originais usam o SDK `ollama-js`/`ollama-python`. Em Go,
  `ollama.Client` chama `POST {OLLAMA_BASE_URL}/api/chat` (sem streaming) e
  `POST {OLLAMA_BASE_URL}/api/embeddings` diretamente via `net/http`, sem framework HTTP — este
  é um script CLI standalone, não um servidor.
- **Interface `Gateway`**: extraída em `evalgate.go` para permitir testar `EvalGate` (scoring,
  decisão de promoção) com um dublê determinístico (`fakeGateway`), sem depender de um Ollama
  local rodando durante `go test`. O original não tem testes automatizados no sentido
  `testing`/`pytest`; os testes Go cobrem a matemática de cosseno, a lógica de decisão do gate
  (incluindo o caso-limite exato no valor da tolerância) e o fluxo ponta a ponta de avaliação com
  scores controlados, além do cliente HTTP com `httptest`.
- Organização em pacotes (`ollama`, `evalgate`) em vez de um único arquivo — convenção deste
  repositório para projetos Go com mais de um componente lógico.

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
go build ./...
go run .
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

Os testes usam `fakeGateway` para simular chat e embeddings sem depender de um Ollama local —
cobrem: similaridade de cosseno (vetores idênticos, ortogonais, opostos), a decisão de promoção
(`DecidirPromocao`, incluindo o limite exato da tolerância), os fluxos
`AvaliarCandidato`/`AvaliarCandidatoSemContexto` de ponta a ponta com scores controlados, e o
cliente HTTP `ollama.Client` com `httptest`.
