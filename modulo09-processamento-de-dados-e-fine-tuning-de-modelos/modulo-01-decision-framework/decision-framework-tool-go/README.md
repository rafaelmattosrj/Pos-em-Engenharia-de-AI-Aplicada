# Decision Framework Tool (Go)

Porte Go de `decision-framework-tool-java` (que por sua vez é o porte Java de
`../decision-framework-tool.js` / `../decision_framework_tool.py`, Módulo
09.1): gate de governança LGPD, AHP (pesos + consistência de Saaty +
agregação de comitê), gate de 4 perguntas ponderado, NPV/DCF, Monte Carlo,
Real Options (árvore binomial CRR) e análise de sensibilidade, para decidir
se vale a pena fazer fine-tuning nos 3 casos reais da Amplitude Seguros.
Inclui também o porte de `../grpo-verifiable-reward-demo.js` (recompensa
verificável + vantagem relativa ao grupo do GRPO).

Sem framework web: lógica de cálculo standalone, com dois binários CLI --
`cmd/decisionframework` (demo do framework de decisão) e `cmd/grpodemo`
(demo do GRPO).

## Origem de cada pacote

| Pacote | Equivalente em `decision-framework-tool.js` |
|---|---|
| `config` | `carregarConfiguracao()` — carrega `amplitude-seguros-casos.json` |
| `ahp` | `derivarPesosAHP`, `calcularConsistenciaAHP`, `agregarMatrizesComite` |
| `governance` | `validarGovernancaDado()` — gate LGPD, roda antes de tudo |
| `framework` | `avaliarFramework` / `avaliarCasoCompleto` — gate de 4 perguntas |
| `finance` (`npv.go`) | `paramsDeterministicos` / `calcularNPV` |
| `finance` (`montecarlo.go`) | `amostrarTriangular` / `simularMonteCarlo` |
| `finance` (`sensitivity.go`) | `analisarSensibilidade` |
| `finance` (`realoptions.go`) | `calcularValorPresenteFluxos`, `derivarVolatilidade`, `precificarOpcaoDeEsperar` |
| `grpo` (`reward.go`) | `recompensaVerificavel()` / `GABARITO`, de `grpo-verifiable-reward-demo.js` |
| `grpo` (`advantage.go`) | `vantagemRelativaAoGrupo()` |
| `grpo` (`jsonextract.go`) | `extrairJson()` |
| `grpo` (`ollama.go`) | `chamarOllama()` — cliente HTTP mínimo pro `/api/generate` do Ollama local |
| `cmd/decisionframework` | fluxo principal (`main()`) do `.js`, aplicado aos 3 casos reais |
| `cmd/grpodemo` | `main()` de `grpo-verifiable-reward-demo.js` / `GrpoDemoMain.java` -- demo standalone do GRPO |

Este é o porte **Go de `decision-framework-tool-java`** (não direto do
`.js`/`.py`) — mesma nomenclatura de tipos/funções em português traduzida do
Java, não uma tradução literal do JS.

## Mantido 1:1

- Todas as constantes de negócio e os 3 casos reais da Amplitude Seguros
  (`amplitude-auto`, `amplitude-saude-empresarial`,
  `amplitude-atendimento-cliente`), embutidos via `go:embed` a partir de
  `config/amplitude-seguros-casos.json`.
- O algoritmo de AHP (média geométrica das linhas, Random Index de Saaty
  n=4, limiar CR < 0.10), o gate de 4 perguntas, NPV/DCF, Monte Carlo
  (10.000 simulações), Real Options (árvore binomial CRR, indução
  retroativa) e a análise de sensibilidade (+/-20%, estilo tornado chart).
- O gabarito e a fórmula de recompensa verificável do GRPO
  (`Gabarito`/`RecompensaVerificavel`), e o cálculo de vantagem relativa ao
  grupo (`A_i = (r_i - média(r)) / desvio_padrão(r)`).

## Testes

`ahp`, `governance`, `framework`, `finance` e `grpo` têm arquivos `_test.go`
(pacote `testing` padrão) cobrindo os mesmos cenários de
`DecisionFrameworkTest` e `GrpoRewardDemoTest` do porte Java: pesos e
consistência de AHP (solo e comitê), gate de governança (base legal, DPA),
gate de 4 perguntas (limiar exato, falha só por dado, decisão técnica em
aberto), NPV/DCF (fórmula fechada de anuidade, atraso sem economia), Monte
Carlo (convergência da amostragem triangular, caso Auto), Real Options
(valor de esperar, volatilidade, u*d≈1) e sensibilidade (ranking por
amplitude), além da recompensa verificável e vantagem relativa ao grupo do
GRPO. `config` e `cmd/*` não têm teste próprio (exercitados indiretamente
pelos pacotes acima e pela execução manual do `main`).

## Demo do GRPO

`cmd/grpodemo` é o equivalente ao `GrpoDemoMain` do Java: amostragem real de
grupo via Ollama local (modelo `gemma4:e2b`, grupo de tamanho 6, temperatura
1.0) -> recompensa verificável -> vantagem relativa ao grupo. Não treina
nada. Requer `ollama serve` rodando localmente com o modelo disponível
(`ollama list`); sem isso, o programa detecta a falha de conexão na
primeira chamada e encerra com uma mensagem, sem lançar exceção.

## Adaptado (sem equivalente direto)

- **JSON embutido no binário**: `config.go` usa `go:embed` para embutir
  `amplitude-seguros-casos.json` no executável, em vez de ler do disco em
  tempo de execução como o `.js`/Java fazem.

## Rodar

```sh
go build ./...                        # compila
go vet ./...                          # lint estático
go test ./...                         # roda os testes de ahp, governance, framework, finance e grpo
go run ./cmd/decisionframework         # roda o demo (3 casos reais + AHP + financeiro)
go run ./cmd/grpodemo                  # roda o demo do GRPO (requer 'ollama serve' local)
```
