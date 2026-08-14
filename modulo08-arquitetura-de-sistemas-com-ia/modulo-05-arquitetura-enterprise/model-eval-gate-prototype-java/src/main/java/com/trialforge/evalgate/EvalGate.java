package com.trialforge.evalgate;

import java.io.IOException;
import java.util.List;

/**
 * Eval Gate de modelo — antes de promover um candidato pra processar tráfego
 * real, roda ele contra um golden set (perguntas com cláusula esperada
 * conhecida) e só promove se o score médio não regredir contra o baseline
 * atual além de uma tolerância.
 *
 * Porte 1:1 de model-eval-gate-prototype.js / model_eval_gate_prototype.py.
 */
public class EvalGate {

    public static final String MODELO_EMBEDDING = "nomic-embed-text";
    public static final String MODELO_BASELINE = "gemma4:e2b";
    public static final String MODELO_CANDIDATO = "gemma4:e2b-mlx";

    // Candidato pode ficar até essa fração abaixo do baseline sem bloquear a promoção —
    // tolerância, não corte exato. Num contexto farmacêutico regulado, faz sentido usar
    // a ponta mais rígida dessa faixa, não a mais frouxa.
    public static final double TOLERANCIA_REGRESSAO = 0.02;

    // Golden set: mesmo banco de cláusulas do TrialForge usado desde o Módulo 2.5/4.5/5.4.
    public static final List<GoldenItem> GOLDEN_SET = List.of(
            new GoldenItem(
                    "Quais são as regras de assentimento pra menores nesse estudo?",
                    "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, "
                            + "além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
                    "RDC ANVISA 466/2012, Art. 4º"),
            new GoldenItem(
                    "O participante pode desistir do estudo a qualquer momento?",
                    "O participante pode retirar seu consentimento a qualquer momento, sem necessidade "
                            + "de justificativa e sem prejuízo ao seu tratamento (RDC ANVISA 466/2012, Art. 5º).",
                    "RDC ANVISA 466/2012, Art. 5º"),
            new GoldenItem(
                    "Qual é o critério de idade mínima pra participar desse estudo?",
                    "A idade mínima para participação no estudo é de doze anos completos na data "
                            + "da assinatura do assentimento, conforme a versão vigente do protocolo aprovada "
                            + "pelo comitê de ética.",
                    "Protocolo Clínico TrialForge, critério de inclusão nº 2"));

    private final OllamaGateway ollama;

    public EvalGate(OllamaGateway ollama) {
        this.ollama = ollama;
    }

    /** Similaridade de cosseno entre dois vetores — lógica pura, testável sem rede. */
    public static double similaridadeCosseno(double[] a, double[] b) {
        double produto = 0, normaA = 0, normaB = 0;
        for (int i = 0; i < a.length; i++) {
            produto += a[i] * b[i];
            normaA += a[i] * a[i];
            normaB += b[i] * b[i];
        }
        return produto / (Math.sqrt(normaA) * Math.sqrt(normaB));
    }

    /**
     * Decisão pura de promoção: o candidato é promovido se a diferença de score
     * em relação ao baseline não for pior que -tolerancia. Extraída para ser
     * testável sem depender de scores reais de um modelo.
     */
    public static boolean decidirPromocao(double scoreBaseline, double scoreCandidato, double tolerancia) {
        double diferenca = scoreCandidato - scoreBaseline;
        return diferenca >= -tolerancia;
    }

    public String gerarResposta(String modelo, String pergunta, String clausula, String fonte)
            throws IOException, InterruptedException {
        String system = "Você redige respostas curtas e precisas sobre regras de estudos clínicos, "
                + "citando a fonte regulatória fornecida.";
        String user = "Pergunta: " + pergunta + "\n\nCláusula regulatória relevante: " + clausula
                + "\nFonte: " + fonte + "\n\nResponda usando essa cláusula.";
        return ollama.chat(modelo, system, user);
    }

    // Simula um candidato REGREDIDO por bug de config, não por modelo pior: o mesmo
    // baseline, mas sem a cláusula no contexto.
    public String gerarRespostaSemContexto(String modelo, String pergunta) throws IOException, InterruptedException {
        String system = "Você redige respostas curtas e precisas sobre regras de estudos clínicos.";
        String user = "Pergunta: " + pergunta;
        return ollama.chat(modelo, system, user);
    }

    public double avaliarCandidato(String modelo, List<GoldenItem> goldenSet) throws IOException, InterruptedException {
        double somaScores = 0;
        for (GoldenItem item : goldenSet) {
            String texto = gerarResposta(modelo, item.pergunta(), item.clausula(), item.fonte());
            double[] embResposta = ollama.embed(MODELO_EMBEDDING, texto);
            double[] embClausula = ollama.embed(MODELO_EMBEDDING, item.clausula());
            double score = similaridadeCosseno(embResposta, embClausula);
            somaScores += score;
            System.out.printf("  [Eval] \"%s...\" -> score %.3f%n", ellipsis(item.pergunta(), 55), score);
        }
        return somaScores / goldenSet.size();
    }

    public double avaliarCandidatoSemContexto(String modelo, List<GoldenItem> goldenSet) throws IOException, InterruptedException {
        double somaScores = 0;
        for (GoldenItem item : goldenSet) {
            String texto = gerarRespostaSemContexto(modelo, item.pergunta());
            double[] embResposta = ollama.embed(MODELO_EMBEDDING, texto);
            double[] embClausula = ollama.embed(MODELO_EMBEDDING, item.clausula());
            double score = similaridadeCosseno(embResposta, embClausula);
            somaScores += score;
            System.out.printf("  [Eval] \"%s...\" -> score %.3f%n", ellipsis(item.pergunta(), 55), score);
        }
        return somaScores / goldenSet.size();
    }

    private static String ellipsis(String texto, int tamanho) {
        return texto.length() <= tamanho ? texto : texto.substring(0, tamanho);
    }
}
