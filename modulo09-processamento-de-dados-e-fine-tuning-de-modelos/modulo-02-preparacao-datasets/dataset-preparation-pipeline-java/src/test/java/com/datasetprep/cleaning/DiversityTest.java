package com.datasetprep.cleaning;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.data.Offset.offset;

class DiversityTest {

    @Test
    void distribuicaoUniformeEntreNFontesTemNumeroEfetivoDeFontesIgualAN() {
        Map<String, Double> dist = Map.of("A", 0.25, "B", 0.25, "C", 0.25, "D", 0.25);
        assertThat(Diversity.numeroEfetivoFontes(dist)).isCloseTo(4.0, offset(1e-9));
    }

    @Test
    void umaUnicaFonteConcentrandoTudoTemEntropiaZeroENumeroEfetivo1() {
        Map<String, Double> dist = Map.of("A", 1.0, "B", 0.0, "C", 0.0);
        assertThat(Diversity.entropiaShannon(dist)).isCloseTo(0.0, offset(1e-9));
        assertThat(Diversity.numeroEfetivoFontes(dist)).isCloseTo(1.0, offset(1e-9));
    }

    @Test
    void suavizarComAlphaMenorAumentaAEntropia() {
        Map<String, Integer> contagens = Map.of("A", 14, "B", 5, "C", 4, "D", 3);
        double hAlto = Diversity.entropiaShannon(Balancing.pesosAmostragemPorTemperatura(contagens, 1.0));
        double hBaixo = Diversity.entropiaShannon(Balancing.pesosAmostragemPorTemperatura(contagens, 0.3));
        assertThat(hBaixo).isGreaterThan(hAlto);
    }
}
