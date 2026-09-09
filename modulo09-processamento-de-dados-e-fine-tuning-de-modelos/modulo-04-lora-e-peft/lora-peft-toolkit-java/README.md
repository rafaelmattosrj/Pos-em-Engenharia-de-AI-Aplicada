# LoRA/PEFT Toolkit (Java)

Porte Java das 6 ferramentas de cálculo/comparação portáveis do Módulo 4 — LoRA e PEFT. Ver `../README.md` para o que foi pulado e por quê.

## Origem (par oficial `.js` / espelho `.py`)

| Classe | Original |
|---|---|
| `AdapterComparison` | `adapter-comparison-tool.js` / `.py` |
| `FullVsLoraTradeoff` | `full-vs-lora-tradeoff-tool.js` / `.py` |
| `LoraRankTradeoff` | `lora-rank-tradeoff-tool.js` / `.py` |
| `RankAdapterComparison` | `rank-adapter-comparison-tool.js` / `.py` |
| `RegionalLoraVsCloudNpv` (+ `NpvCalculator`) | `regional-lora-vs-cloud-npv.js` / `.py` |
| `LoraManagedApiPreview` | `lora-managed-api-preview-tool.js` / `.py` |

## Mantido 1:1

Todos os dados reais (execuções de treino, saídas capturadas do `mlx_lm generate`, config LoRA) e toda a lógica de cálculo/parsing/decisão foram portados com os mesmos valores e os mesmos testes (cenário a cenário) do original.

## Adaptado

- **JSON**: o original usa `JSON.parse` nativo do JS; aqui, Gson (`JsonObject`), já que Java precisa de uma biblioteca para parsear um schema que não controla (gabarito de extração).
- **`calcularNPV`**: reimplementado em `NpvCalculator` (não há módulo Java compartilhado entre pastas de módulos diferentes neste repositório — cada módulo é portado para uma pasta `-java`/`-go` própria). Mudanças na fórmula original em `modulo-01-decision-framework/decision-framework-tool.js` precisam ser replicadas aqui manualmente.
- **Subprocesso MLX** (`AdapterComparison.rodarGenerateReal`, `RankAdapterComparison.rodarReal`): usa `ProcessBuilder` para chamar `python3 -m mlx_lm generate`, igual ao `child_process.spawnSync` original. Requer MLX + modelo/adaptador instalados localmente para rodar de verdade; os testes cobrem só o parsing (contra saída real fixada, capturada em 2026-09-04), sem precisar rodar o modelo.
- **HTTP real** (`LoraManagedApiPreview.enviarOuPrever`): usa `java.net.http.HttpClient` em vez de `fetch`; só envia de verdade se `TOGETHER_API_KEY` estiver no ambiente, senão devolve o modo preview.
- **Testes vs. demo**: mesma convenção do resto do repositório — `mvn test` roda os testes JUnit5; `mvn exec:java` roda a demo (`Main`), que não inclui os subprocessos MLX nem a chamada HTTP real.

## Rodar

```sh
mvn compile test    # 44 testes
mvn exec:java       # demo no console
```
