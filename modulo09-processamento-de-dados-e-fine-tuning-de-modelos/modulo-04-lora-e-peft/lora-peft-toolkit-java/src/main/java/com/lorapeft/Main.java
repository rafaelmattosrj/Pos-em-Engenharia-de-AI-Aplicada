package com.lorapeft;

import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.util.List;

/**
 * Demo: imprime os resultados das calculadoras de trade-off de LoRA/PEFT com
 * dado real capturado nesta disciplina. Nao dispara os subprocessos MLX
 * (AdapterComparison/RankAdapterComparison) nem a chamada HTTP real
 * (LoraManagedApiPreview) -- isso depende de hardware/rede locais; a demo
 * mostra a logica pura (calculo, parsing de fixture, montagem de requisicao).
 */
public final class Main {

    private Main() {
    }

    public static void main(String[] args) {
        System.setOut(new PrintStream(System.out, true, StandardCharsets.UTF_8));

        System.out.println("=".repeat(72));
        System.out.println("LoRA rank trade-off (Módulo 4.3)");
        System.out.println("=".repeat(72));
        for (LoraRankTradeoff.Execucao e : LoraRankTradeoff.EXECUCOES_REAIS) {
            System.out.printf("  Rank %d: val loss %.3f -> %.3f (%.2f%% redução), pico mem %.3fGB%n",
                    e.rank(), e.valLossInicial(), e.valLossFinal(), LoraRankTradeoff.calcularReducaoValLoss(e), e.picoMemGB());
        }
        LoraRankTradeoff.Execucao recomendado = LoraRankTradeoff.recomendarRankMinimo(LoraRankTradeoff.EXECUCOES_REAIS, 0.10);
        System.out.println("  Recomendação (margem 10%): rank " + recomendado.rank());

        System.out.println();
        System.out.println("=".repeat(72));
        System.out.println("Full fine-tuning vs. LoRA (Módulo 4.4)");
        System.out.println("=".repeat(72));
        FullVsLoraTradeoff.DecisaoFullFineTuning decisao =
                FullVsLoraTradeoff.valeAPenaFullFineTuning(FullVsLoraTradeoff.CONFIGURACOES_REAIS, 20);
        System.out.printf("  Melhor LoRA: %s | Ganho do full fine-tuning: %.2f%% | Vale a pena (limiar 20%%)? %s%n",
                decisao.melhorLora(), decisao.ganhoPercentual(), decisao.valeAPena() ? "SIM" : "NÃO");

        System.out.println();
        System.out.println("=".repeat(72));
        System.out.println("NPV: parceria regional, GPU alugada vs. LoRA local (Módulo 4.1)");
        System.out.println("=".repeat(72));
        NpvCalculator.Resultado npvGerenciado = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_JOB_GERENCIADO));
        NpvCalculator.Resultado npvLora = NpvCalculator.calcularNPV(
                RegionalLoraVsCloudNpv.paramsBase(RegionalLoraVsCloudNpv.CUSTO_LORA_LOCAL));
        System.out.printf("  GPU alugada: NPV em 24 meses = R$ %.2f (breakeven: %s)%n",
                npvGerenciado.npv(), npvGerenciado.mesBreakeven() == null ? "não atinge" : "mês " + npvGerenciado.mesBreakeven());
        System.out.printf("  LoRA local:  NPV em 24 meses = R$ %.2f (breakeven: mês %s)%n",
                npvLora.npv(), npvLora.mesBreakeven());

        System.out.println();
        System.out.println("=".repeat(72));
        System.out.println("Preview de requisição LoRA gerenciada (Together AI, Módulo 4.2)");
        System.out.println("=".repeat(72));
        LoraManagedApiPreview.Requisicao requisicao =
                LoraManagedApiPreview.montarRequisicaoLoraGerenciada(null, "file-exemplo-amplitude-seguros");
        System.out.println("  URL: " + requisicao.url());
        System.out.println("  Authorization: " + requisicao.headers().get("Authorization"));
        System.out.println("  Body: " + requisicao.body());
        System.out.println("  (sem TOGETHER_API_KEY no ambiente -- modo preview, nada é enviado de verdade)");

        System.out.println();
        System.out.println("Nota: AdapterComparison e RankAdapterComparison (comparação real via");
        System.out.println("mlx_lm.generate, Módulos 4.2/4.3) não têm demo aqui -- dependem de MLX +");
        System.out.println("modelo/adaptador locais. Ver testes em AdapterComparisonTest /");
        System.out.println("RankAdapterComparisonTest para o parsing validado contra saída real.");
    }
}
