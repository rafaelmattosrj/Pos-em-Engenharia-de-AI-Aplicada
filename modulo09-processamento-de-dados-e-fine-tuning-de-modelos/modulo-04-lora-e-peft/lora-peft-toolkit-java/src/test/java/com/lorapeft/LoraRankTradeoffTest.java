package com.lorapeft;

import org.assertj.core.data.Offset;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class LoraRankTradeoffTest {

    @Test
    void rank8ReduzValLossEmAproximadamente81PorCento() {
        LoraRankTradeoff.Execucao rank8 = LoraRankTradeoff.EXECUCOES_REAIS.get(1);
        double r = LoraRankTradeoff.calcularReducaoValLoss(rank8);
        assertThat(r).isCloseTo(81.17, Offset.offset(0.5));
    }

    @Test
    void rank4Para8DobraOsParametrosTreinaveis() {
        List<LoraRankTradeoff.ComparacaoSucessiva> comparacoes = LoraRankTradeoff.compararExecucoesSucessivas(LoraRankTradeoff.EXECUCOES_REAIS);
        assertThat(comparacoes.get(0).razaoParametros()).isEqualTo(2.0);
    }

    @Test
    void rank8Para16TambemDobraOsParametrosTreinaveis() {
        List<LoraRankTradeoff.ComparacaoSucessiva> comparacoes = LoraRankTradeoff.compararExecucoesSucessivas(LoraRankTradeoff.EXECUCOES_REAIS);
        assertThat(comparacoes.get(1).razaoParametros()).isEqualTo(2.0);
    }

    @Test
    void maisRankSempreMelhoraOValLossFinal() {
        List<LoraRankTradeoff.ComparacaoSucessiva> comparacoes = LoraRankTradeoff.compararExecucoesSucessivas(LoraRankTradeoff.EXECUCOES_REAIS);
        comparacoes.forEach(c -> assertThat(c.melhoriaValLoss()).isGreaterThan(0));
    }

    @Test
    void custoDeMemoriaExtraPorDobraDeRankEPequeno() {
        List<LoraRankTradeoff.ComparacaoSucessiva> comparacoes = LoraRankTradeoff.compararExecucoesSucessivas(LoraRankTradeoff.EXECUCOES_REAIS);
        comparacoes.forEach(c -> assertThat(c.custoMemoriaExtraGB()).isLessThan(0.2));
    }

    @Test
    void comMargemDe10PorCentoRecomendaRank16() {
        LoraRankTradeoff.Execucao r = LoraRankTradeoff.recomendarRankMinimo(LoraRankTradeoff.EXECUCOES_REAIS, 0.10);
        assertThat(r.rank()).isEqualTo(16);
    }

    @Test
    void comMargemDe60PorCentoRecomendaRank8() {
        LoraRankTradeoff.Execucao r = LoraRankTradeoff.recomendarRankMinimo(LoraRankTradeoff.EXECUCOES_REAIS, 0.60);
        assertThat(r.rank()).isEqualTo(8);
    }

    @Test
    void quatroBitOcupaCerca65PorCentoMenosEspacoEmDiscoQueBf16() {
        LoraRankTradeoff.ResultadoQuantizacao r = LoraRankTradeoff.compararQuantizacao(LoraRankTradeoff.BF16, LoraRankTradeoff.QUATRO_BIT);
        assertThat(r.reducaoDiscoPct()).isCloseTo(65.0, Offset.offset(1.0));
    }

    @Test
    void quatroBitReduzPicoDeMemoriaDeTreinoEmCerca61PorCento() {
        LoraRankTradeoff.ResultadoQuantizacao r = LoraRankTradeoff.compararQuantizacao(LoraRankTradeoff.BF16, LoraRankTradeoff.QUATRO_BIT);
        assertThat(r.reducaoMemTreinoPct()).isCloseTo(61.3, Offset.offset(1.0));
    }

    @Test
    void custoDeValLossDaQuantizacaoEPequeno() {
        LoraRankTradeoff.ResultadoQuantizacao r = LoraRankTradeoff.compararQuantizacao(LoraRankTradeoff.BF16, LoraRankTradeoff.QUATRO_BIT);
        assertThat(r.custoValLossPct()).isLessThan(10);
    }

    @Test
    void doraUsaCerca7Ponto5PorCentoMaisParametrosTreinaveisQueLoraNoMesmoRank() {
        LoraRankTradeoff.ResultadoAdaptacao r = LoraRankTradeoff.compararTipoAdaptacao(LoraRankTradeoff.LORA, LoraRankTradeoff.DORA);
        assertThat(r.razaoParametros()).isCloseTo(1.075, Offset.offset(0.02));
    }

    @Test
    void doraCustaMemoriaExtraDeTreinoEntre0Ponto2E0Ponto35GbAMaisQueLora() {
        LoraRankTradeoff.ResultadoAdaptacao r = LoraRankTradeoff.compararTipoAdaptacao(LoraRankTradeoff.LORA, LoraRankTradeoff.DORA);
        assertThat(r.custoMemoriaExtraGB()).isGreaterThan(0.2).isLessThan(0.35);
    }

    @Test
    void doraEmpataComLoraEmValLossFinalSemGanho() {
        LoraRankTradeoff.ResultadoAdaptacao r = LoraRankTradeoff.compararTipoAdaptacao(LoraRankTradeoff.LORA, LoraRankTradeoff.DORA);
        assertThat(r.diferencaValLoss()).isEqualTo(0);
    }
}
