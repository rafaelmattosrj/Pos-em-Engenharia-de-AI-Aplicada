# LoRA/PEFT Toolkit (Go)

Porte Go das 6 ferramentas de cálculo/comparação portáveis do Módulo 4 — LoRA e PEFT. Ver `../README.md` para o que foi pulado e por quê.

## Origem (par oficial `.js` / espelho `.py`)

| Pacote | Original |
|---|---|
| `adaptercomparison` | `adapter-comparison-tool.js` / `.py` |
| `fullvslora` | `full-vs-lora-tradeoff-tool.js` / `.py` |
| `lorarank` | `lora-rank-tradeoff-tool.js` / `.py` |
| `rankadapter` | `rank-adapter-comparison-tool.js` / `.py` |
| `regionalnpv` | `regional-lora-vs-cloud-npv.js` / `.py` |
| `managedapi` | `lora-managed-api-preview-tool.js` / `.py` |

## Mantido 1:1

Todos os dados reais e toda a lógica de cálculo/parsing/decisão foram portados com os mesmos valores e os mesmos testes (cenário a cenário) do original.

## Adaptado

- **JSON**: `encoding/json` padrão da stdlib, decodificando para `map[string]any` quando o schema não é controlado (gabarito de extração) e para structs tipadas quando é (corpo de requisição da Together AI).
- **`CalcularNPV`**: reimplementado em `regionalnpv` (não há módulo Go compartilhado entre pastas de módulos diferentes neste repositório). Mudanças na fórmula original em `modulo-01-decision-framework/decision-framework-tool.js` precisam ser replicadas aqui manualmente.
- **Subprocesso MLX** (`adaptercomparison.RodarGenerateReal`, `rankadapter.RodarReal`): usa `os/exec` para chamar `python3 -m mlx_lm generate`. Requer MLX + modelo/adaptador instalados localmente para rodar de verdade; os testes cobrem só o parsing (contra saída real fixada, capturada em 2026-09-04).
- **HTTP real** (`managedapi.EnviarOuPrever`): usa `net/http` em vez de `fetch`; só envia de verdade se `TOGETHER_API_KEY` estiver no ambiente, senão devolve o modo preview.

## Rodar

```sh
go build ./...
go vet ./...
go test ./...
go run .   # demo no console
```
