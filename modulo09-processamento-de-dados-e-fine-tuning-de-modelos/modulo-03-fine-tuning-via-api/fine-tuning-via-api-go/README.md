# Fine-Tuning via API Toolkit (Go)

Porte Go de parte das ferramentas de `modulo-03-fine-tuning-via-api/`
(originais em JS, com espelho `.py`): conversão para o formato Gemini, gate
de confiança de OCR, validação/comparação de hiperparâmetro, automação de
fine-tuning via Vertex AI (upload, criação de job com trava de confirmação,
acompanhamento com backoff e retry), versionamento de modelo (hash de
dataset, ficha, model card), reavaliação do caso Amplitude Saúde
Empresarial (reabre o gate de decisão do Módulo 1), escala de dataset
(305 brutos -> dedup -> 200 balanceados) e o extra Dolly-15k (dataset real
alternativo). Sem framework web: coleção de pacotes-biblioteca, mais um
`cmd/demo` que orquestra o fluxo de ponta a ponta -- equivalente ao
`Main.java` do porte Java.

## Origem de cada pacote

| Pacote | Original | Módulo |
|---|---|---|
| `gemini` | `dataset-upload-and-tracking-tool.js` + `finetuning-automation-tool.js` | 3.2 / 3.4 |
| `ocrgate` | `dataset-upload-and-tracking-tool.js` | 3.2 |
| `hyperparam` | `hyperparameter-and-monitoring-tool.js` | 3.3 / 3.4 |
| `automation` | `finetuning-automation-tool.js` | 3.4 |
| `versioning` | `model-versioning-tool.js` | 3.5 |
| `reavaliacao` | `reavaliacao-saude-empresarial.js` | 3.2 |
| `decisionframework` | `modulo-01-decision-framework/decision-framework-tool.js` (subconjunto: só AHP + gate de 4 perguntas, sem NPV/Monte Carlo/Real Options) | 1.2 / 1.3, reusado no 3.2 |
| `minhash` | `modulo-02-preparacao-datasets/dataset-cleaning-balancing-tool.js` (MinHash+LSH, balanceamento por temperatura, entropia) | reusado no 3.2 |
| `vertexai` | implementação HTTP real do cliente de job da Vertex AI (`obterTokenAcesso`, consulta/criação de job), usada por `automation`/`versioning` | 3.2 / 3.4 / 3.5 |
| `datasetscaling` | `m3-dataset-scaling-tool.js` / `DatasetScaling.java` (305 brutos -> dedup -> 200 balanceados, 120 Auto + 80 Saúde Empresarial) | 3.2 |
| `dolly` | `dolly-dataset-real-starter.js` + `dolly-vertex-pipeline.js` / `DollyDatasetStarter.java` + `DollyVertexPipeline.java` (extra: dataset real Dolly-15k) | 3.2 / 3.4 / 3.5 |
| `cmd/demo` | `Main.java` — demo de ponta a ponta orquestrando os pacotes acima | 3.2 / 3.3 / 3.4 / 3.5 |

## Mantido 1:1

- Constantes e parâmetros de negócio: limiar padrão do gate de OCR (0.85),
  faixas válidas de hiperparâmetro, fórmula de backoff exponencial com teto,
  parâmetros de MinHash (k=32, semente=42, LSH bandas=8/linhas=4, alpha=0.3
  de temperatura) e a trava de confirmação explícita do `automation`
  (`ExigirConfirmacao`) — mesmo incidente real do Módulo 3.3 que motivou a
  trava no original.
- O subconjunto do framework de decisão (`decisionframework.AvaliarFramework`
  / `DerivarPesosAHP`) usado por `reavaliacao.ConstruirCasoNoveMesesDepois`
  para reabrir o caso Amplitude Saúde Empresarial 9 meses depois, com a
  mesma taxa de crescimento de score já projetada no caso original (não
  inventa número novo).
- Os cenários de teste dos arquivos `.js`/Java originais, onde há teste: ver
  `automation_test.go`, `gemini_test.go`, `hyperparam_test.go`,
  `minhash_test.go`, `ocrgate_test.go`, `versioning_test.go`,
  `datasetscaling_test.go` (mesmos cenários de `DatasetScalingTest.java`,
  incluindo os números exatos de exemplos brutos/dedup/balanceados e as
  métricas de diversidade antes/depois) e `dolly_test.go` (carregamento e
  filtro de compatibilidade do JSONL, pipeline de preparação com dedup, e
  a orquestração de upload/criação/acompanhamento de `RodarPipeline` com
  fakes -- sem teste correspondente no Java, que também não tem
  `DollyDatasetStarterTest`/`DollyVertexPipelineTest`).

## Adaptado (sem equivalente direto)

- **`decisionframework` e `minhash` são porte AUTOCONTIDO, duplicado**: o
  `.js` original importa essas funções por `require()` direto dos arquivos
  do Módulo 1 e do Módulo 2. Como este repositório não tem um mecanismo de
  módulo compartilhado entre projetos Go/Maven independentes, a lógica foi
  duplicada aqui (mesmos dados de entrada e mesma fórmula) em vez de
  referenciada. Mudanças nos originais em
  `modulo-01-decision-framework/decision-framework-tool.js` ou
  `modulo-02-preparacao-datasets/dataset-cleaning-balancing-tool.js`
  precisam ser replicadas manualmente aqui.
- **`vertexai.HTTPClient`**: chama `gcloud auth print-access-token` via
  `os/exec` (equivalente ao `execSync` do `.js`) para obter o token, e usa
  `net/http` puro para falar com `aiplatform.googleapis.com`. Requer
  `gcloud auth login` e acesso ao projeto GCP para rodar de verdade — não
  roda em CI/teste automatizado; por isso `vertexai` não tem `_test.go`
  próprio (as funções que dependem dele, como `automation`, são testadas
  injetando um `ConsultarFn`/`JobClient` falso).
- **`dolly.CarregarDolly` requer o arquivo real baixado**: o extra Dolly-15k
  só roda de ponta a ponta com `databricks-dolly-15k.jsonl` (13MB, ~15 mil
  linhas, CC-BY-SA-3.0) baixado localmente — não incluído no repositório.
  Os testes de `dolly_test.go` usam um fixture pequeno em
  `dolly/testdata/dolly-sample.jsonl`, não o dataset real.
- **`dolly.RodarPipeline`**: os pontos de injeção (`UploadFn`, `CriarJobFn`,
  `AcompanharFn`) são tipos de função Go simples, em vez de interfaces
  funcionais Java (`FineTuningAutomation.UploadFn`/`CriarJobFn`/`AcompanharFn`)
  — mesmo padrão de injeção de `automation`/`vertexai`, só que idiomático
  pra Go (closures em vez de interface de método único).
- **JSON**: `encoding/json` da stdlib (`map[string]any`), sem biblioteca
  externa — mantém o módulo sem dependências, coerente com o padrão "Go
  simples, sem framework" da convenção de porte para CLI/biblioteca
  standalone.

## Rodar

```sh
go build ./...     # compila
go vet ./...       # lint estático
go test ./...      # roda os testes de automation, datasetscaling, dolly, gemini, hyperparam, minhash, ocrgate, versioning
go run ./cmd/demo  # roda o demo de ponta a ponta (equivalente ao Main.java)
```

`decisionframework`, `reavaliacao` e `vertexai` não têm `_test.go` próprio
nesta pasta: `decisionframework`/`reavaliacao` são exercitados
indiretamente (paridade validada manualmente contra o porte Java), e
`vertexai.HTTPClient` só tem lógica de rede real (requer `gcloud auth
login` e acesso ao GCP, fora do escopo de CI/teste automatizado). O extra
Dolly (`dolly.RodarPipeline`) segue o mesmo princípio: a lógica pura é
testada com fakes, e só a chamada real à Vertex AI fica fora do escopo de
CI/teste automatizado.
