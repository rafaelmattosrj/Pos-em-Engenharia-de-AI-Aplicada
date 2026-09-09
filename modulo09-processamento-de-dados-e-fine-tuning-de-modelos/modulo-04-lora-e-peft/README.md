# Módulo 4 — LoRA e PEFT: porte Java/Go

Este módulo tem ferramentas JS/Python (par oficial `.js` + espelho `.py`) de duas naturezas bem diferentes. Só uma delas foi portada para Java e Go — a outra não tem equivalente idiomático nessas linguagens, e forçar um porte seria só teatro.

## Portado (`lora-peft-toolkit-java/` e `lora-peft-toolkit-go/`)

Seis ferramentas de cálculo/comparação, CLIs standalone, sem servidor HTTP:

- `adapter-comparison-tool` (M4.2) — parsing puro (montagem de argumentos do `mlx_lm generate`, parse da saída, comparação com gabarito) + disparo de subprocesso `python3 -m mlx_lm generate` via `ProcessBuilder`/`os/exec`.
- `full-vs-lora-tradeoff-tool` (M4.4) — cálculo puro sobre dados reais capturados.
- `lora-rank-tradeoff-tool` (M4.3) — cálculo puro (trade-off de rank, quantização bf16 vs. 4-bit, LoRA vs. DoRA).
- `rank-adapter-comparison-tool` (M4.3, companion) — mesmo padrão do adapter-comparison: parsing puro + subprocesso.
- `regional-lora-vs-cloud-npv` (M4.1) — reabre `calcularNPV` do Módulo 1 (decision-framework-tool). Como este repositório não compartilha um pacote Java/Go entre pastas de módulos diferentes, a fórmula foi **reimplementada** em `NpvCalculator`/`regionalnpv` (documentado no código; mudanças na fórmula original precisam ser replicadas manualmente aqui).
- `lora-managed-api-preview-tool` (M4.2, companion) — monta e opcionalmente envia (se `TOGETHER_API_KEY` estiver no ambiente) uma requisição real para a API de fine-tuning da Together AI.

As seis foram agrupadas num único projeto Java (`com.lorapeft`, uma classe por ferramenta) e um único projeto Go (um pacote por ferramenta), em vez de 6 mini-projetos — são calculadoras relacionadas do mesmo tema (trade-off de fine-tuning), e agrupar evita a fragmentação de 12 `pom.xml`/`go.mod` quase idênticos.

`AdapterComparison`/`RankAdapterComparison` (Java) e `adaptercomparison`/`rankadapter` (Go) dependem de `mlx-data/test.jsonl` (arquivo de dado já existente no repositório, referenciado por caminho relativo) para os testes de carregamento de exemplo — isso não é "porte" do dado, é o mesmo padrão de leitura que o JS/Python original já usa.

## Não portado (pulado de propósito)

Estes dependem de hardware específico (Apple Silicon/MLX) ou de um runtime Python de ML sem equivalente idiomático em Java/Go — não é falta de tempo, é ausência real de porte possível sem reescrever um motor de treino de LLM em Java/Go:

- `local-lora-training-tool.js`, `local-lora-training-hf-tool.py` — treino local de verdade via MLX/HF.
- `colab-lora-training-notebook.ipynb` — notebook Colab.
- `mlx-adapters/`, `mlx-adapters-rank4/`, `mlx-adapters-rank16/`, `mlx-data/` — artefatos de treino (pesos, dados), não código. (`mlx-data/test.jsonl` continua sendo *lido* pelos testes portados, como dado de fixture — ver acima.)
- `lora-rank4-config.yaml`, `lora-rank8-config.yaml`, `lora-rank16-config.yaml` — configs de treino, não lógica.

## Validação

- Java: `mvn compile test` (JUnit5 + AssertJ) — 44 testes, todos passando.
- Go: `go build ./...`, `go vet ./...`, `go test ./...` — todos os pacotes passando.
