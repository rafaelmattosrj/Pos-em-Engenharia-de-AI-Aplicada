package com.trialforge.evalgate;

/**
 * Porte Java de model-eval-gate-prototype.js / model_eval_gate_prototype.py.
 * Ver README.md deste projeto para detalhes de paridade e adaptacoes.
 */
public class Main {

    public static void main(String[] args) {
        String baseUrl = System.getenv().getOrDefault("OLLAMA_BASE_URL", "http://localhost:11434");
        EvalGate gate = new EvalGate(new OllamaHttpGateway(baseUrl));

        try {
            System.out.printf("%n[Eval Gate] Avaliando BASELINE (%s) contra golden set de %d perguntas...%n",
                    EvalGate.MODELO_BASELINE, EvalGate.GOLDEN_SET.size());
            double scoreBaseline = gate.avaliarCandidato(EvalGate.MODELO_BASELINE, EvalGate.GOLDEN_SET);
            System.out.printf("[Eval Gate] Baseline: score médio %.3f%n", scoreBaseline);

            System.out.println("\n== Cenário 1: candidato real (variante MLX do mesmo modelo) ==");
            System.out.printf("[Eval Gate] Avaliando CANDIDATO (%s) contra o mesmo golden set...%n", EvalGate.MODELO_CANDIDATO);
            double scoreCandidato1 = gate.avaliarCandidato(EvalGate.MODELO_CANDIDATO, EvalGate.GOLDEN_SET);
            System.out.printf("[Eval Gate] Candidato: score médio %.3f%n", scoreCandidato1);
            double diferenca1 = scoreCandidato1 - scoreBaseline;
            boolean promove1 = EvalGate.decidirPromocao(scoreBaseline, scoreCandidato1, EvalGate.TOLERANCIA_REGRESSAO);
            System.out.printf("[Eval Gate] Diferença: %s%.3f | Tolerância: -%s%n",
                    diferenca1 >= 0 ? "+" : "", diferenca1, EvalGate.TOLERANCIA_REGRESSAO);
            System.out.printf("[Eval Gate] Decisão: %s — caso limite, dois modelos reais e parecidos; pode mudar entre "
                    + "execuções por variância do próprio modelo, por isso a decisão nunca deve ser no olho, sempre pelo gate.%n",
                    promove1 ? "PROMOVE" : "BLOQUEIA");

            System.out.println("\n== Cenário 2: candidato regredido por config, não por modelo pior ==");
            System.out.printf("[Eval Gate] Avaliando %s SEM a cláusula no contexto (simula bug de RAG/config, mesmo modelo)...%n",
                    EvalGate.MODELO_BASELINE);
            double scoreCandidato2 = gate.avaliarCandidatoSemContexto(EvalGate.MODELO_BASELINE, EvalGate.GOLDEN_SET);
            System.out.printf("[Eval Gate] Candidato regredido: score médio %.3f%n", scoreCandidato2);
            double diferenca2 = scoreCandidato2 - scoreBaseline;
            boolean promove2 = EvalGate.decidirPromocao(scoreBaseline, scoreCandidato2, EvalGate.TOLERANCIA_REGRESSAO);
            System.out.printf("[Eval Gate] Diferença: %s%.3f | Tolerância: -%s%n",
                    diferenca2 >= 0 ? "+" : "", diferenca2, EvalGate.TOLERANCIA_REGRESSAO);
            System.out.printf("[Eval Gate] Decisão: %s — mesmo modelo, contexto perdido; o gate pega uma regressão de "
                    + "config que nenhuma troca de modelo causou.%n", promove2 ? "PROMOVE" : "BLOQUEIA");
        } catch (Exception erro) {
            System.err.println("[Erro não tratado] " + erro.getMessage());
            System.exit(1);
        }
    }
}
