# Fine-Tuning via API Toolkit (Java)

Porte Java de `modulo-03-fine-tuning-via-api/` (originais em JS, com
espelho `.py`): conversão para o formato Gemini, gate de confiança de OCR,
validação/comparação de hiperparâmetro, automação de fine-tuning via Vertex
AI (upload, criação de job com trava de confirmação explícita,
acompanhamento com backoff exponencial e retry), versionamento de modelo
(hash de dataset, ficha de versionamento, model card), escala de dataset
(reusa dedup MinHash+LSH e balanceamento por temperatura do Módulo 2),
reavaliação do caso Amplitude Saúde Empresarial (reusa o gate de decisão do
Módulo 1), e o extra Dolly-15k (dataset real alternativo). Agrupadas num
único projeto Maven, organizado por classe de responsabilidade — mesma
decisão de agrupamento de `dataset-preparation-pipeline-java` (ver
`../../modulo-02-preparacao-datasets/dataset-preparation-pipeline-java/README.md`).

Sem Spring Boot: coleção de ferramentas CLI, Maven puro, cliente HTTP via
`java.net.http.HttpClient` (sem lib externa).

## Origem de cada classe

| Classe | Original | Módulo |
|---|---|---|
| `GeminiConverter` | `dataset-upload-and-tracking-tool.js` + `finetuning-automation-tool.js` | 3.2 / 3.4 |
| `OcrConfidenceGate` | `dataset-upload-and-tracking-tool.js` | 3.2 |
| `HyperparameterValidator` | `hyperparameter-and-monitoring-tool.js` | 3.3 / 3.4 |
| `FineTuningAutomation` | `finetuning-automation-tool.js` | 3.4 |
| `ModelVersioning` | `model-versioning-tool.js` | 3.5 |
| `TempoUtil` | formatação de duração real de job, porte de `formatarDuracao` | 3.2 / 3.5 |
| `DatasetScaling` | `m3-dataset-scaling-tool.js` (305 exemplos brutos -> dedup -> 200 balanceados, 120 Auto + 80 Saúde Empresarial) | 3.2 |
| `MinHashDedupBalancer` | `modulo-02-preparacao-datasets/dataset-cleaning-balancing-tool.js` (MinHash+LSH, balanceamento por temperatura, entropia) | reusado no 3.2 |
| `ReavaliacaoSaudeEmpresarial` | `reavaliacao-saude-empresarial.js` | 3.2 |
| `DecisionFrameworkCore` | `modulo-01-decision-framework/decision-framework-tool.js` (subconjunto: só AHP + gate de 4 perguntas, sem NPV/Monte Carlo/Real Options) | 1.2 / 1.3, reusado no 3.2 |
| `VertexAiHttpClient` / `VertexAiJobClient` | implementação HTTP real do cliente de job (`obterTokenAcesso` via `gcloud`, chamada REST à `aiplatform.googleapis.com`) | 3.2 / 3.4 / 3.5 |
| `DollyDatasetStarter` | `dolly-dataset-real-starter.js` (Extra: dataset real alternativo Dolly-15k, CC-BY-SA-3.0) | 3.2 |
| `DollyVertexPipeline` | `dolly-vertex-pipeline.js` (Extra parte 2/2: upload + job real + acompanhamento) | 3.2 / 3.4 / 3.5 |
| `JsonUtil` | serializador/parser JSON mínimo, próprio (não porta um arquivo específico) | — |
| `Main` | demo que encadeia conversão Gemini, gate de OCR, validação de hiperparâmetro, escala de dataset e reavaliação do caso Saúde Empresarial | 3.2 / 3.3 |

## Mantido 1:1

- Constantes e parâmetros de negócio: limiar padrão do gate de OCR (0.85),
  faixas válidas de hiperparâmetro, fórmula de backoff exponencial com teto,
  parâmetros de MinHash (k=32, semente=42, LSH bandas=8/linhas=4, alpha=0.3
  de temperatura), e a trava de confirmação explícita em
  `FineTuningAutomation` — mesmo incidente real do Módulo 3.3 (hiperparâmetro
  inválido cria job real sem aviso) que motivou a trava no original.
- A seção fixa sobre o job real de preference tuning (DPO) do Módulo 3.5 em
  `ModelVersioning` (40 dos 200 exemplos convertidos em pares
  `chosen`/`rejected`, referenciando `preference-dataset-amplitude.jsonl`).
- O subconjunto do framework de decisão (`DecisionFrameworkCore`) usado por
  `ReavaliacaoSaudeEmpresarial` para reabrir o caso Amplitude Saúde
  Empresarial 9 meses depois, com a mesma taxa de crescimento de score já
  projetada no caso original.
- Os mesmos cenários de teste dos `.js` originais, onde há teste:
  `DatasetScalingTest`, `FineTuningAutomationTest`, `GeminiConverterTest`,
  `HyperparameterValidatorTest`, `ModelVersioningTest`,
  `OcrConfidenceGateTest`, `ReavaliacaoSaudeEmpresarialTest`.

## Adaptado (sem equivalente direto)

- **`DecisionFrameworkCore` e `MinHashDedupBalancer` são porte AUTOCONTIDO,
  duplicado**: o `.js` original importa essas funções por `require()`
  direto dos arquivos do Módulo 1 e do Módulo 2. Como este projeto Maven é
  independente dos projetos dos outros módulos, a lógica foi duplicada aqui
  (mesmos dados de entrada e mesma fórmula) em vez de referenciada.
  Mudanças nos originais em
  `modulo-01-decision-framework/decision-framework-tool.js` ou
  `modulo-02-preparacao-datasets/dataset-cleaning-balancing-tool.js`
  precisam ser replicadas manualmente aqui.
- **`JsonUtil`**: serializador/parser JSON recursivo mínimo, escrito à mão
  (objetos, arrays, `String`, `Number`, `Boolean`, `null`), para evitar
  puxar Jackson/Gson numa ferramenta CLI standalone — mantém o projeto
  "Maven puro" conforme a convenção da skill de porte.
- **`VertexAiHttpClient`**: obtém o token via `ProcessBuilder` chamando
  `gcloud auth print-access-token` (equivalente ao `execSync` do `.js`) e
  usa `java.net.http.HttpClient` nativo para falar com
  `aiplatform.googleapis.com`. Requer `gcloud auth login` e acesso ao
  projeto GCP para rodar de verdade — não roda em CI/teste automatizado; por
  isso não tem teste próprio (as classes que dependem dele, como
  `FineTuningAutomation`, são testadas injetando uma função/cliente falso).
- **`DecisionFrameworkCore` e `ReavaliacaoSaudeEmpresarial` sem teste
  dedicado do primeiro**: `DecisionFrameworkCore` é exercitado
  indiretamente pelos testes de `ReavaliacaoSaudeEmpresarialTest`, sem
  suíte própria equivalente a `DecisionFrameworkTest` do porte do Módulo 1.
- Sem Spring Boot, sem LangChain4j: nenhuma das ferramentas expõe servidor
  HTTP nem usa abstração de LLM além de chamada HTTP direta.

## Rodar

```sh
mvn compile test   # compila e roda os testes (dataset scaling, automação, Gemini,
                    # hiperparâmetro, versionamento, gate de OCR, reavaliação)
mvn exec:java      # roda o demo (Main): conversão Gemini, gate de OCR, validação de
                    # hiperparâmetro, escala do dataset e reavaliação Saúde Empresarial
```

`ModelVersioning` (ficha/model card real) e `DollyVertexPipeline` dependem
de um job real consultado via `VertexAiHttpClient` (requer `gcloud auth
login` e acesso ao GCP) — não são chamados pelo `Main` padrão; use-os a
partir do seu próprio código com a credencial disponível.
