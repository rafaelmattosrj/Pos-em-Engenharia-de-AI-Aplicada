package com.lorapeft;

import java.util.List;

/**
 * Porte de full-vs-lora-tradeoff-tool.js (Modulo 4.4). Dados reais de um
 * treino full fine-tuning comparado contra os tres treinos LoRA do Modulo 4.3
 * -- nenhum numero aqui e estimado, sao capturados em 2026-08-08.
 */
public final class FullVsLoraTradeoff {

    private FullVsLoraTradeoff() {
    }

    public record Configuracao(
            String tipo,
            double percentualModelo,
            double valLossFinal,
            double picoMemGB,
            int tamanhoCheckpointMB) {
    }

    public static final List<Configuracao> CONFIGURACOES_REAIS = List.of(
            new Configuracao("LoRA rank 4", 0.074, 1.246, 10.787, 13),
            new Configuracao("LoRA rank 8", 0.147, 0.895, 10.833, 27),
            new Configuracao("LoRA rank 16", 0.295, 0.725, 10.930, 52),
            new Configuracao("Full fine-tuning", 22.567, 0.612, 15.338, 1992));

    public record ResultadoInferencia(boolean acertouTodosCampos) {
    }

    public static final ResultadoInferencia EXEMPLO_FACIL = new ResultadoInferencia(true);
    public static final ResultadoInferencia EXEMPLO_COM_DISTRATORES = new ResultadoInferencia(true);

    public static double calcularGanhoValLoss(Configuracao base, Configuracao comparado) {
        double ganho = ((base.valLossFinal() - comparado.valLossFinal()) / base.valLossFinal()) * 100;
        return round2(ganho);
    }

    public record RazaoCusto(double parametro, double memoria, double checkpoint) {
    }

    public static RazaoCusto calcularRazaoCusto(Configuracao base, Configuracao comparado) {
        return new RazaoCusto(
                round1(comparado.percentualModelo() / base.percentualModelo()),
                round2(comparado.picoMemGB() / base.picoMemGB()),
                round1((double) comparado.tamanhoCheckpointMB() / base.tamanhoCheckpointMB()));
    }

    public record DecisaoFullFineTuning(String melhorLora, double ganhoPercentual, boolean valeAPena) {
    }

    public static DecisaoFullFineTuning valeAPenaFullFineTuning(List<Configuracao> configuracoes, double limiarGanhoMinimo) {
        Configuracao melhorLora = configuracoes.stream()
                .filter(c -> c.tipo().startsWith("LoRA"))
                .min((a, b) -> Double.compare(a.valLossFinal(), b.valLossFinal()))
                .orElseThrow();
        Configuracao full = configuracoes.stream()
                .filter(c -> c.tipo().equals("Full fine-tuning"))
                .findFirst()
                .orElseThrow();
        double ganho = calcularGanhoValLoss(melhorLora, full);
        return new DecisaoFullFineTuning(melhorLora.tipo(), ganho, ganho >= limiarGanhoMinimo);
    }

    private static double round1(double v) {
        return Math.round(v * 10.0) / 10.0;
    }

    private static double round2(double v) {
        return Math.round(v * 100.0) / 100.0;
    }
}
