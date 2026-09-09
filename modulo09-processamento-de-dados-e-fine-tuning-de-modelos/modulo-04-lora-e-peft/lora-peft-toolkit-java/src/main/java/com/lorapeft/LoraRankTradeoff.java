package com.lorapeft;

import java.util.ArrayList;
import java.util.List;

/**
 * Porte de lora-rank-tradeoff-tool.js (Modulo 4.3). Dados reais de tres
 * treinos LoRA (rank 4/8/16) mais duas comparacoes adicionais (quantizacao
 * bf16 vs. 4-bit / QLoRA, e LoRA vs. DoRA), todos capturados rodando
 * mlx_lm.lora nesta disciplina.
 */
public final class LoraRankTradeoff {

    private LoraRankTradeoff() {
    }

    public record Execucao(
            int rank,
            double parametrosTreinaveis,
            double percentualModelo,
            double valLossInicial,
            double valLossFinal,
            double picoMemGB,
            double itPorSegundoFinal,
            int tamanhoAdapterMB) {
    }

    public static final List<Execucao> EXECUCOES_REAIS = List.of(
            new Execucao(4, 3.408e6, 0.074, 4.752, 1.246, 10.787, 7.401, 13),
            new Execucao(8, 6.816e6, 0.147, 4.752, 0.895, 10.833, 7.295, 27),
            new Execucao(16, 13.631e6, 0.295, 4.752, 0.725, 10.930, 7.253, 52));

    public record ConfigQuantizacao(
            double tamanhoModeloDiscoGB, double picoMemTreinoGB, double valLossFinal, int tamanhoAdapterMB) {
    }

    public static final ConfigQuantizacao BF16 = new ConfigQuantizacao(10.241, 10.833, 0.895, 27);
    public static final ConfigQuantizacao QUATRO_BIT = new ConfigQuantizacao(3.583, 4.193, 0.932, 27);

    public record ConfigAdaptacao(double parametrosTreinaveis, double picoMemGB, double valLossFinal, int tamanhoAdapterMB) {
    }

    public static final ConfigAdaptacao LORA = new ConfigAdaptacao(6.816e6, 10.833, 0.895, 27);
    public static final ConfigAdaptacao DORA = new ConfigAdaptacao(7.328e6, 11.099, 0.895, 28);

    public static double calcularReducaoValLoss(Execucao execucao) {
        double reducao = ((execucao.valLossInicial() - execucao.valLossFinal()) / execucao.valLossInicial()) * 100;
        return round2(reducao);
    }

    public record ComparacaoSucessiva(
            int deRank, int paraRank, double razaoParametros, double melhoriaValLoss,
            double melhoriaPercentual, double custoMemoriaExtraGB) {
    }

    public static List<ComparacaoSucessiva> compararExecucoesSucessivas(List<Execucao> execucoes) {
        List<ComparacaoSucessiva> comparacoes = new ArrayList<>();
        for (int i = 1; i < execucoes.size(); i++) {
            Execucao anterior = execucoes.get(i - 1);
            Execucao atual = execucoes.get(i);
            double razaoParametros = round2(atual.parametrosTreinaveis() / anterior.parametrosTreinaveis());
            double melhoriaValLoss = round3(anterior.valLossFinal() - atual.valLossFinal());
            double melhoriaPercentual = round2((melhoriaValLoss / anterior.valLossFinal()) * 100);
            double custoMemoriaExtraGB = round3(atual.picoMemGB() - anterior.picoMemGB());
            comparacoes.add(new ComparacaoSucessiva(anterior.rank(), atual.rank(), razaoParametros,
                    melhoriaValLoss, melhoriaPercentual, custoMemoriaExtraGB));
        }
        return comparacoes;
    }

    public record ResultadoQuantizacao(double reducaoDiscoPct, double reducaoMemTreinoPct, double custoValLoss, double custoValLossPct) {
    }

    public static ResultadoQuantizacao compararQuantizacao(ConfigQuantizacao bf16, ConfigQuantizacao quatroBit) {
        double reducaoDiscoPct = round1((1 - quatroBit.tamanhoModeloDiscoGB() / bf16.tamanhoModeloDiscoGB()) * 100);
        double reducaoMemTreinoPct = round1((1 - quatroBit.picoMemTreinoGB() / bf16.picoMemTreinoGB()) * 100);
        double custoValLoss = round3(quatroBit.valLossFinal() - bf16.valLossFinal());
        double custoValLossPct = round1((custoValLoss / bf16.valLossFinal()) * 100);
        return new ResultadoQuantizacao(reducaoDiscoPct, reducaoMemTreinoPct, custoValLoss, custoValLossPct);
    }

    public record ResultadoAdaptacao(double razaoParametros, double custoMemoriaExtraGB, double diferencaValLoss) {
    }

    public static ResultadoAdaptacao compararTipoAdaptacao(ConfigAdaptacao lora, ConfigAdaptacao dora) {
        double razaoParametros = round3(dora.parametrosTreinaveis() / lora.parametrosTreinaveis());
        double custoMemoriaExtraGB = round3(dora.picoMemGB() - lora.picoMemGB());
        double diferencaValLoss = round3(dora.valLossFinal() - lora.valLossFinal());
        return new ResultadoAdaptacao(razaoParametros, custoMemoriaExtraGB, diferencaValLoss);
    }

    public static Execucao recomendarRankMinimo(List<Execucao> execucoes, double margemAceitavel) {
        double melhorValLoss = execucoes.stream().mapToDouble(Execucao::valLossFinal).min().orElseThrow();
        double limiar = melhorValLoss * (1 + margemAceitavel);
        return execucoes.stream()
                .filter(c -> c.valLossFinal() <= limiar)
                .min((a, b) -> Integer.compare(a.rank(), b.rank()))
                .orElseThrow();
    }

    private static double round1(double v) {
        return Math.round(v * 10.0) / 10.0;
    }

    private static double round2(double v) {
        return Math.round(v * 100.0) / 100.0;
    }

    private static double round3(double v) {
        return Math.round(v * 1000.0) / 1000.0;
    }
}
