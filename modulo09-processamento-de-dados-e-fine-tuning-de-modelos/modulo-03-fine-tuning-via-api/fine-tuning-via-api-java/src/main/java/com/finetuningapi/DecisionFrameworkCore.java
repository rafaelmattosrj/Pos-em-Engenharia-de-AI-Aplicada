package com.finetuningapi;

import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Subconjunto do gate de 4 perguntas + AHP (Modulo 1.2/1.3), necessario so
 * pra reavaliar o caso Amplitude Saude Empresarial (Modulo 3.2).
 *
 * ADAPTACAO: porte AUTOCONTIDO de modulo-01-decision-framework/decision-framework-tool.js
 * (funcoes derivarPesosAHP, avaliarFramework, RECOMENDACAO, PERGUNTAS,
 * CHAVES_PERGUNTAS, carregarConfiguracao). O original em JS importa essas
 * funcoes diretamente do arquivo do Modulo 1 via require(); como este repo
 * nao tem um mecanismo de modulo compartilhado entre projetos Maven/Go
 * independentes, a logica foi duplicada aqui (mesmo dado de entrada,
 * amplitude-seguros-casos.json, copiado como recurso deste projeto).
 */
public final class DecisionFrameworkCore {

    public static final class Recomendacao {
        public static final String FINE_TUNING = "Fine-tuning vale a pena";
        public static final String CONTINUAR_PROMPT_RAG = "Continue com prompt + RAG. Fine-tuning ainda não.";
        public static final String ESPERAR = "Espere acumular dado, depois treine.";
        public static final String BLOQUEADO_POR_GOVERNANCA = "Bloqueado: resolva a governança do dado antes de reavaliar.";

        private Recomendacao() {
        }
    }

    public static final List<String> CHAVES_PERGUNTAS = List.of("p1", "p2", "p3", "p4");

    public static final Map<String, String> PERGUNTAS = Map.of(
            "p1", "A tarefa é estreita e repetida, ou aberta e variável?",
            "p2", "Já esgotou prompt engineering + RAG + roteamento, sem chegar na qualidade/custo/latência necessários?",
            "p3", "Tem dado de exemplo suficiente, diverso e de qualidade pra treinar?",
            "p4", "A tarefa é estável o bastante pra não virar esteira de retreino constante?");

    private static final double RANDOM_INDEX_N4 = 0.90;

    private DecisionFrameworkCore() {
    }

    public static double[] derivarPesosAHP(double[][] matriz) {
        int n = matriz.length;
        double[] mediasGeometricas = new double[n];
        for (int i = 0; i < n; i++) {
            double produto = 1;
            for (double v : matriz[i]) produto *= v;
            mediasGeometricas[i] = Math.pow(produto, 1.0 / n);
        }
        double soma = 0;
        for (double v : mediasGeometricas) soma += v;
        double[] pesos = new double[n];
        for (int i = 0; i < n; i++) pesos[i] = mediasGeometricas[i] / soma;
        return pesos;
    }

    public record Sinal(double score, String sinal) {
    }

    public record ResultadoAvaliacao(
            boolean aprovado, boolean falhaSoDado, String recomendacao,
            List<Integer> perguntasFalhas, double scoreComposto, Map<String, Sinal> sinaisPorPergunta) {
    }

    public static ResultadoAvaliacao avaliarFramework(Map<String, Double> scores, double[] pesos, double limiarVerde) {
        Map<String, Sinal> sinaisPorPergunta = new LinkedHashMap<>();
        List<Integer> perguntasFalhas = new ArrayList<>();

        for (int i = 0; i < CHAVES_PERGUNTAS.size(); i++) {
            String chave = CHAVES_PERGUNTAS.get(i);
            double score = scores.get(chave);
            boolean verde = score >= limiarVerde;
            sinaisPorPergunta.put(chave, new Sinal(score, verde ? "VERDE" : "VERMELHO"));
            if (!verde) perguntasFalhas.add(i + 1);
        }

        double scoreComposto = 0;
        for (int i = 0; i < CHAVES_PERGUNTAS.size(); i++) {
            scoreComposto += scores.get(CHAVES_PERGUNTAS.get(i)) * pesos[i];
        }

        boolean aprovado = perguntasFalhas.isEmpty();
        boolean falhaSoDado = perguntasFalhas.size() == 1 && perguntasFalhas.get(0) == 3;

        return new ResultadoAvaliacao(
                aprovado, falhaSoDado,
                aprovado ? Recomendacao.FINE_TUNING : Recomendacao.CONTINUAR_PROMPT_RAG,
                perguntasFalhas, Math.round(scoreComposto * 10000) / 10000.0, sinaisPorPergunta);
    }

    public record OpcaoReal(double custoDeErroEsperadoPorChamada, double taxaCrescimentoScorePorMes, double scoreAlvo) {
    }

    public record Financeiro(int volumeInicialMensal, double crescimentoMensalModa, OpcaoReal opcaoReal) {
    }

    public record Caso(String id, Map<String, Double> scores, Financeiro financeiro) {
    }

    public record Configuracao(double limiarVerde, double[][] matrizAhp, List<Caso> casos) {
        public Caso caso(String id) {
            return casos.stream().filter(c -> c.id().equals(id)).findFirst()
                    .orElseThrow(() -> new IllegalArgumentException("caso nao encontrado: " + id));
        }
    }

    @SuppressWarnings("unchecked")
    public static Configuracao carregarConfiguracao() throws IOException {
        try (InputStream in = DecisionFrameworkCore.class.getResourceAsStream("/amplitude-seguros-casos.json")) {
            if (in == null) throw new IOException("recurso amplitude-seguros-casos.json não encontrado no classpath");
            String bruto = new String(in.readAllBytes(), StandardCharsets.UTF_8);
            Map<String, Object> raiz = JsonUtil.parseObjeto(bruto);

            double limiarVerde = ((Number) raiz.get("limiarVerde")).doubleValue();
            List<Object> matrizBruta = (List<Object>) ((Map<String, Object>) raiz.get("ahp")).get("matriz");
            double[][] matriz = new double[matrizBruta.size()][];
            for (int i = 0; i < matrizBruta.size(); i++) {
                List<Object> linha = (List<Object>) matrizBruta.get(i);
                matriz[i] = new double[linha.size()];
                for (int j = 0; j < linha.size(); j++) matriz[i][j] = ((Number) linha.get(j)).doubleValue();
            }

            List<Object> casosBrutos = (List<Object>) raiz.get("casos");
            List<Caso> casos = new ArrayList<>();
            for (Object casoObj : casosBrutos) {
                Map<String, Object> casoMap = (Map<String, Object>) casoObj;
                Map<String, Object> scoresMap = (Map<String, Object>) casoMap.get("scores");
                Map<String, Double> scores = new LinkedHashMap<>();
                scoresMap.forEach((k, v) -> scores.put(k, ((Number) v).doubleValue()));

                Financeiro financeiro = null;
                Object financeiroObj = casoMap.get("financeiro");
                if (financeiroObj != null) {
                    Map<String, Object> financeiroMap = (Map<String, Object>) financeiroObj;
                    int volumeInicialMensal = ((Number) financeiroMap.get("volumeInicialMensal")).intValue();
                    double crescimentoMensalModa = ((Number) ((Map<String, Object>) financeiroMap.get("crescimentoMensal")).get("moda")).doubleValue();
                    OpcaoReal opcaoReal = null;
                    Object opcaoRealObj = financeiroMap.get("opcaoReal");
                    if (opcaoRealObj != null) {
                        Map<String, Object> opcaoRealMap = (Map<String, Object>) opcaoRealObj;
                        opcaoReal = new OpcaoReal(
                                ((Number) opcaoRealMap.get("custoDeErroEsperadoPorChamada")).doubleValue(),
                                ((Number) opcaoRealMap.get("taxaCrescimentoScorePorMes")).doubleValue(),
                                ((Number) opcaoRealMap.get("scoreAlvo")).doubleValue());
                    }
                    financeiro = new Financeiro(volumeInicialMensal, crescimentoMensalModa, opcaoReal);
                }

                casos.add(new Caso((String) casoMap.get("id"), scores, financeiro));
            }

            return new Configuracao(limiarVerde, matriz, casos);
        }
    }
}
