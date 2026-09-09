# Decision Framework Tool (Java)

Porte Java de `../decision-framework-tool.js` (Módulo 09.1, com espelho em
`../decision_framework_tool.py`): gate de governança LGPD, AHP (pesos +
consistência de Saaty + agregação de comitê), gate de 4 perguntas ponderado,
NPV/DCF, Monte Carlo, Real Options (árvore binomial CRR) e análise de
sensibilidade, para decidir se vale a pena fazer fine-tuning nos 3 casos
reais da Amplitude Seguros. Inclui também o porte de
`../grpo-verifiable-reward-demo.js` (recompensa verificável + vantagem
relativa ao grupo do GRPO).

Sem Spring Boot: lógica de cálculo standalone, Maven puro, sem servidor
HTTP. É a base a partir da qual `../decision-framework-tool-go/` foi
portado.

## Origem de cada classe

| Classe | Equivalente em `decision-framework-tool.js` |
|---|---|
| `config.ConfigLoader` | `carregarConfiguracao()` — lê `amplitude-seguros-casos.json` do classpath |
| `Ahp` | `derivarPesosAHP`, `calcularConsistenciaAHP`, `agregarMatrizesComite` |
| `GovernanceGate` | `validarGovernancaDado()` — gate LGPD, roda antes de tudo |
| `FrameworkEvaluation` | `avaliarFramework` / `avaliarCasoCompleto` — gate de 4 perguntas |
| `NpvCalculator` | `paramsDeterministicos` / `calcularNPV` |
| `MonteCarloSimulator` | `amostrarTriangular` / `simularMonteCarlo` |
| `SensitivityAnalyzer` | `analisarSensibilidade` |
| `RealOptionsPricer` | `calcularValorPresenteFluxos`, `derivarVolatilidade`, `precificarOpcaoDeEsperar` |
| `Recomendacao` | objeto `RECOMENDACAO` |
| `grpo.VerifiableReward` | `recompensaVerificavel()` / `GABARITO`, de `grpo-verifiable-reward-demo.js` |
| `grpo.GroupRelativeAdvantage` | `vantagemRelativaAoGrupo()` |
| `grpo.JsonExtractor` | `extrairJson()` |
| `grpo.OllamaClient` | `chamarOllama()` — cliente mínimo pro `/api/generate` do Ollama local |
| `grpo.GrpoDemoMain` | `main()` de `grpo-verifiable-reward-demo.js` |
| `Main` | fluxo principal (`main()`) do `.js`, aplicado aos 3 casos reais |

## Mantido 1:1

- Todas as constantes de negócio e os 3 casos reais da Amplitude Seguros
  (`amplitude-auto`, `amplitude-saude-empresarial`,
  `amplitude-atendimento-cliente`), em `amplitude-seguros-casos.json`.
- O algoritmo de AHP (média geométrica das linhas, Random Index de Saaty
  n=4, limiar CR < 0.10), o gate de 4 perguntas, NPV/DCF, Monte Carlo
  (10.000 simulações), Real Options (árvore binomial CRR, indução
  retroativa) e a análise de sensibilidade (+/-20%, estilo tornado chart).
- O gabarito e a fórmula de recompensa verificável do GRPO
  (`VerifiableReward.recompensaVerificavel`), a vantagem relativa ao grupo
  (`A_i = (r_i - média(r)) / desvio_padrão(r)`), e o modelo/parâmetros do
  demo (`gemma4:e2b`, grupo de tamanho 6, temperatura 1.0).
- Os mesmos cenários de teste do `.js` original: `DecisionFrameworkTest`
  cobre AHP (pesos e consistência, agregação de comitê), governança, os 3
  casos reais, NPV/DCF, Monte Carlo, Real Options e sensibilidade;
  `GrpoRewardDemoTest` cobre a recompensa verificável e a vantagem de grupo.

## Adaptado (sem equivalente direto)

- **Jackson em vez de `JSON.parse`**: `amplitude-seguros-casos.json` é lido
  do classpath com `jackson-databind` para um modelo tipado
  (`AmplitudeConfig`/`Caso`/`Financeiro`/`Governanca`/`OpcaoReal`), em vez
  de um objeto solto como no `.js`.
- **`GrpoDemoMain` é uma classe separada de `Main`**: o `exec-maven-plugin`
  do `pom.xml` aponta só para `Main` (o demo do framework de decisão); para
  rodar o demo do GRPO é preciso apontar explicitamente o `mainClass` (ver
  "Rodar" abaixo). Requer `ollama serve` rodando localmente com o modelo
  `gemma4:e2b` disponível (`ollama list`) — sem isso, o demo imprime a
  instrução e encerra sem treinar/chamar nada.

## Rodar

```sh
mvn compile test    # compila e roda DecisionFrameworkTest + GrpoRewardDemoTest
mvn exec:java       # roda o demo do framework de decisão (Main) nos 3 casos reais
mvn exec:java -Dexec.mainClass="com.amplitudeseguros.decisionframework.grpo.GrpoDemoMain"
                     # roda o demo do GRPO (requer Ollama local com o modelo gemma4:e2b)
```
