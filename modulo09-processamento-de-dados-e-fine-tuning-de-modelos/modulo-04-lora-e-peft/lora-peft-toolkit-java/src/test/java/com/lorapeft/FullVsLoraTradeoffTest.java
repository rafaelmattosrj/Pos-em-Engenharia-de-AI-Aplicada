package com.lorapeft;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class FullVsLoraTradeoffTest {

    private static final List<FullVsLoraTradeoff.Configuracao> CONFIGS = FullVsLoraTradeoff.CONFIGURACOES_REAIS;

    private FullVsLoraTradeoff.Configuracao acha(String tipo) {
        return CONFIGS.stream().filter(c -> c.tipo().equals(tipo)).findFirst().orElseThrow();
    }

    @Test
    void fullFineTuningMelhoraValLossFrenteALoraRank8Em31Ponto6PorCento() {
        double ganho = FullVsLoraTradeoff.calcularGanhoValLoss(acha("LoRA rank 8"), acha("Full fine-tuning"));
        assertThat(ganho).isCloseTo(31.62, org.assertj.core.data.Offset.offset(0.5));
    }

    @Test
    void fullFineTuningUsa153VezesMaisParametrosTreinaveisQueLoraRank8() {
        FullVsLoraTradeoff.RazaoCusto razao = FullVsLoraTradeoff.calcularRazaoCusto(acha("LoRA rank 8"), acha("Full fine-tuning"));
        assertThat(razao.parametro()).isCloseTo(153.5, org.assertj.core.data.Offset.offset(1.0));
    }

    @Test
    void checkpointDoFullFineTuningE73Ponto8VezesMaiorQueOAdaptadorLoraRank8() {
        FullVsLoraTradeoff.RazaoCusto razao = FullVsLoraTradeoff.calcularRazaoCusto(acha("LoRA rank 8"), acha("Full fine-tuning"));
        assertThat(razao.checkpoint()).isEqualTo(73.8);
    }

    @Test
    void fullFineTuningUsaMaisMemoriaDePicoQueQualquerConfiguracaoLora() {
        FullVsLoraTradeoff.Configuracao full = acha("Full fine-tuning");
        CONFIGS.stream().filter(c -> c.tipo().startsWith("LoRA"))
                .forEach(l -> assertThat(full.picoMemGB()).isGreaterThan(l.picoMemGB()));
    }

    @Test
    void aplicaLimiar20PorCentoSobreOGanhoDeFullFineTuningVsMelhorLora() {
        FullVsLoraTradeoff.DecisaoFullFineTuning r = FullVsLoraTradeoff.valeAPenaFullFineTuning(CONFIGS, 20);
        assertThat(r.melhorLora()).isEqualTo("LoRA rank 16");
        assertThat(r.valeAPena()).isFalse();
    }

    @Test
    void aplicaLimiar10PorCentoSobreOMesmoCriterioDeDecisao() {
        FullVsLoraTradeoff.DecisaoFullFineTuning r = FullVsLoraTradeoff.valeAPenaFullFineTuning(CONFIGS, 10);
        assertThat(r.valeAPena()).isTrue();
    }

    @Test
    void validaResultadoDeInferenciaDosDoisExemplos() {
        assertThat(FullVsLoraTradeoff.EXEMPLO_FACIL.acertouTodosCampos()).isTrue();
        assertThat(FullVsLoraTradeoff.EXEMPLO_COM_DISTRATORES.acertouTodosCampos()).isTrue();
    }
}
