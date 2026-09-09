package com.amplitudeseguros.decisionframework;

import com.amplitudeseguros.decisionframework.config.Caso;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Gate de 4 perguntas com score de confiança ponderado por AHP — equivalente a
 * avaliarFramework/avaliarCasoCompleto em decision-framework-tool.js.
 */
public final class FrameworkEvaluation {

    public static final List<String> CHAVES_PERGUNTAS = List.of("p1", "p2", "p3", "p4");

    private FrameworkEvaluation() {
    }

    public record SinalPergunta(double score, String sinal) {
    }

    public record Resultado(
            boolean bloqueadoPorGovernanca,
            List<String> motivosGovernanca,
            boolean aprovado,
            boolean falhaSoDado,
            boolean decisaoTecnicaEmAberto,
            Recomendacao recomendacao,
            List<Integer> perguntasFalhas,
            Double scoreComposto,
            Map<String, SinalPergunta> sinaisPorPergunta) {
    }

    /** Gate de 4 perguntas isolado, sem o gate de governança — usado pelas demos de AHP de comitê. */
    public static Resultado avaliarFramework(Map<String, Double> scores, double[] pesos, double limiarVerde) {
        Map<String, SinalPergunta> sinaisPorPergunta = new LinkedHashMap<>();
        List<Integer> perguntasFalhas = new ArrayList<>();

        for (int i = 0; i < CHAVES_PERGUNTAS.size(); i++) {
            String chave = CHAVES_PERGUNTAS.get(i);
            double score = scores.get(chave);
            boolean verde = score >= limiarVerde;
            sinaisPorPergunta.put(chave, new SinalPergunta(score, verde ? "VERDE" : "VERMELHO"));
            if (!verde) {
                perguntasFalhas.add(i + 1);
            }
        }

        double scoreComposto = 0;
        for (int i = 0; i < CHAVES_PERGUNTAS.size(); i++) {
            scoreComposto += scores.get(CHAVES_PERGUNTAS.get(i)) * pesos[i];
        }

        boolean aprovado = perguntasFalhas.isEmpty();
        // "só falha por dado" -- a única reprovação que Real Options resolve
        boolean falhaSoDado = perguntasFalhas.size() == 1 && perguntasFalhas.get(0) == 3;

        return new Resultado(
                false, List.of(),
                aprovado,
                falhaSoDado,
                aprovado, // decisaoTecnicaEmAberto: aprovado no gate abre a escolha de técnica (LoRA/full/API), Módulos 3 e 4
                aprovado ? Recomendacao.FINE_TUNING : Recomendacao.CONTINUAR_PROMPT_RAG,
                perguntasFalhas,
                Math.round(scoreComposto * 10000.0) / 10000.0,
                sinaisPorPergunta);
    }

    /**
     * Orquestra o pipeline completo: governança primeiro (bloqueador, grátis), só
     * entra no gate de 4 perguntas se a governança aprovar.
     */
    public static Resultado avaliarCasoCompleto(Caso caso, double[] pesos, double limiarVerde) {
        GovernanceGate.Resultado governanca = GovernanceGate.validar(caso.governanca());
        if (!governanca.aprovado()) {
            return new Resultado(
                    true, governanca.motivos(),
                    false, false, false,
                    Recomendacao.BLOQUEADO_POR_GOVERNANCA,
                    List.of(), null, Map.of());
        }
        Resultado semGovernanca = avaliarFramework(caso.scores(), pesos, limiarVerde);
        return new Resultado(
                false, List.of(),
                semGovernanca.aprovado(), semGovernanca.falhaSoDado(), semGovernanca.decisaoTecnicaEmAberto(),
                semGovernanca.recomendacao(), semGovernanca.perguntasFalhas(), semGovernanca.scoreComposto(),
                semGovernanca.sinaisPorPergunta());
    }
}
