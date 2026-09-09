package com.finetuningapi;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class HyperparameterValidatorTest {

    @Test
    void aceitaConfigRealUsadaNoJobDeProducao() {
        HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(3, 5.0));
    }

    @Test
    void rejeitaEpochCountZero() {
        assertThatThrownBy(() -> HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(0, 5.0)))
                .hasMessageContaining("epochCount deve ser inteiro entre 1 e 20");
    }

    @Test
    void rejeitaEpochsNegativas() {
        assertThatThrownBy(() -> HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(-3, 1.0)))
                .hasMessageContaining("epochCount");
    }

    @Test
    void rejeitaLearningRateMultiplierForaDaFaixa() {
        assertThatThrownBy(() -> HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(3, 50.0)))
                .hasMessageContaining("learningRateMultiplier");
    }

    @Test
    void nenhumaDivergenciaQuandoTudoBate() {
        var divergencias = HyperparameterValidator.compararHiperparametros(Map.of("epochCount", 3), Map.of("epochCount", 3));
        assertThat(divergencias).isEmpty();
    }

    @Test
    void detectaOCasoRealEpochCountAplicadoAusente() {
        Map<String, Object> aplicado = new java.util.HashMap<>();
        aplicado.put("epochCount", null);
        var divergencias = HyperparameterValidator.compararHiperparametros(Map.of("epochCount", 0), aplicado);
        assertThat(divergencias).hasSize(1);
        assertThat(divergencias.get(0)).contains("default silencioso");
    }

    @Test
    void detectaDivergenciaDeValorNumerico() {
        var divergencias = HyperparameterValidator.compararHiperparametros(Map.of("epochCount", 3), Map.of("epochCount", 5));
        assertThat(divergencias).hasSize(1);
    }

    @Test
    void naoApontaDivergenciaQuandoValorIgualMasTipoDiferente() {
        var pedido = Map.<String, Object>of("epochCount", 3, "learningRateMultiplier", 5);
        var aplicado = Map.<String, Object>of("epochCount", "3", "learningRateMultiplier", "5");
        var divergencias = HyperparameterValidator.compararHiperparametros(pedido, aplicado);
        assertThat(divergencias).isEmpty();
    }

}
