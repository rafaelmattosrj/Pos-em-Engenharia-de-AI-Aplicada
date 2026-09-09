package com.amplitudeseguros.decisionframework.grpo;

import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.within;

class GrpoRewardDemoTest {

    @Test
    void extraiJsonCercadoDeTextoSolto() {
        var resultado = JsonExtractor.extrairJson("texto antes {\"a\": 1, \"b\": 2} texto depois");
        assertThat(resultado).isPresent();
        assertThat(resultado.get()).containsEntry("a", 1).containsEntry("b", 2);
    }

    @Test
    void retornaVazioQuandoNaoHaJson() {
        assertThat(JsonExtractor.extrairJson("isso nao tem json nenhum")).isEmpty();
    }

    @Test
    void retornaVazioParaJsonMalformadoSemLancarExcecao() {
        assertThat(JsonExtractor.extrairJson("{\"a\": invalido}")).isEmpty();
    }

    @Test
    void candidatoIdenticoAoGabaritoRecebeRecompensaUm() {
        assertThat(VerifiableReward.recompensaVerificavel(VerifiableReward.GABARITO)).isEqualTo(1.0);
    }

    @Test
    void candidatoSemJsonValidoRecebeRecompensaZero() {
        assertThat(VerifiableReward.recompensaVerificavel(null)).isEqualTo(0.0);
    }

    @Test
    void umCampoErradoEmSeisReduzRecompensaProporcionalmente() {
        Map<String, Object> parcial = new HashMap<>(VerifiableReward.GABARITO);
        parcial.put("status", "campo errado");
        assertThat(VerifiableReward.recompensaVerificavel(parcial)).isCloseTo(5.0 / 6, within(1e-9));
    }

    @Test
    void diferencaDeArredondamentoNoValorAindaContaComoAcerto() {
        Map<String, Object> candidato = new HashMap<>(VerifiableReward.GABARITO);
        candidato.put("estimated_amount_brl", 8450.001);
        assertThat(VerifiableReward.recompensaVerificavel(candidato)).isEqualTo(1.0);
    }

    @Test
    void grupoComVarianciaRealGeraVantagensPositivasENegativas() {
        var resultado = GroupRelativeAdvantage.calcular(new double[]{1.0, 0.5, 0.0});
        assertThat(resultado.media()).isCloseTo(0.5, within(1e-9));
        assertThat(resultado.vantagens()[0]).isGreaterThan(0);
        assertThat(resultado.vantagens()[2]).isLessThan(0);
        assertThat(resultado.vantagens()[1]).isCloseTo(0.0, within(1e-9));
    }

    @Test
    void grupoDegeneradoProduzVantagemZeroSemDividirPorZero() {
        var resultado = GroupRelativeAdvantage.calcular(new double[]{0.83, 0.83, 0.83, 0.83});
        assertThat(resultado.desvio()).isEqualTo(0.0);
        for (double v : resultado.vantagens()) {
            assertThat(v).isEqualTo(0.0);
        }
    }

    @Test
    void recompensasIguaisDentroDoGrupoRecebemAMesmaVantagem() {
        var resultado = GroupRelativeAdvantage.calcular(new double[]{1.0, 1.0, 0.0, 0.0});
        assertThat(resultado.vantagens()[0]).isEqualTo(resultado.vantagens()[1]);
        assertThat(resultado.vantagens()[2]).isEqualTo(resultado.vantagens()[3]);
        assertThat(resultado.vantagens()[0]).isGreaterThan(resultado.vantagens()[2]);
    }
}
