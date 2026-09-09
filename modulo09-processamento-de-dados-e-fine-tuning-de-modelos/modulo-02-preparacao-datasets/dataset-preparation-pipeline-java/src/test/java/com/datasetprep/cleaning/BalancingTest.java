package com.datasetprep.cleaning;

import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class BalancingTest {

    @Test
    void alpha1SemSuavizacaoReproduzADistribuicaoProporcionalOriginal() {
        Map<String, Integer> contagens = Map.of("A", 8, "B", 2);
        var pesos = Balancing.pesosAmostragemPorTemperatura(contagens, 1);
        assertThat(pesos.get("A")).isCloseTo(0.8, org.assertj.core.data.Offset.offset(1e-9));
        assertThat(pesos.get("B")).isCloseTo(0.2, org.assertj.core.data.Offset.offset(1e-9));
    }

    @Test
    void alphaMenorSuavizaADistribuicaoEmDirecaoAUniforme() {
        Map<String, Integer> contagens = Map.of("A", 8, "B", 2);
        var pesosBaixoAlpha = Balancing.pesosAmostragemPorTemperatura(contagens, 0.3);
        var pesosAltoAlpha = Balancing.pesosAmostragemPorTemperatura(contagens, 1);
        assertThat(pesosBaixoAlpha.get("A")).isLessThan(pesosAltoAlpha.get("A"));
        assertThat(pesosBaixoAlpha.get("B")).isGreaterThan(pesosAltoAlpha.get("B"));
    }

    @Test
    void alocacaoSemRestricaoDeCapacidadeSomaExatamenteAoAlvo() {
        Map<String, Double> pesos = Map.of("A", 0.5, "B", 0.3, "C", 0.2);
        var aloc = Balancing.alocarMaiorResto(pesos, 17);
        int soma = aloc.values().stream().mapToInt(Integer::intValue).sum();
        assertThat(soma).isEqualTo(17);
    }

    @Test
    void alocacaoCapacitadaNuncaAlocaMaisQueACapacidadeRealDeNenhumaFonte() {
        Map<String, Integer> contagens = Map.of(
                "Oficina Estrela", 14, "Auto Center Silva", 5, "Funilaria Rio Bonito", 4, "Oficina Nova Alianca", 3);
        for (int alvo : new int[]{26, 22, 20, 18}) {
            var aloc = Balancing.alocarComCapacidade(contagens, Balancing.ALPHA_TEMPERATURA, alvo);
            for (var f : aloc.keySet()) {
                assertThat(aloc.get(f)).as(f).isLessThanOrEqualTo(contagens.get(f));
            }
        }
    }

    @Test
    void alocacaoCapacitadaSomaExatamenteAoAlvoEmTodosOsCasosTestados() {
        Map<String, Integer> contagens = Map.of(
                "Oficina Estrela", 14, "Auto Center Silva", 5, "Funilaria Rio Bonito", 4, "Oficina Nova Alianca", 3);
        for (int alvo : new int[]{26, 22, 20, 18}) {
            var aloc = Balancing.alocarComCapacidade(contagens, Balancing.ALPHA_TEMPERATURA, alvo);
            int soma = aloc.values().stream().mapToInt(Integer::intValue).sum();
            assertThat(soma).isEqualTo(alvo);
        }
    }
}
