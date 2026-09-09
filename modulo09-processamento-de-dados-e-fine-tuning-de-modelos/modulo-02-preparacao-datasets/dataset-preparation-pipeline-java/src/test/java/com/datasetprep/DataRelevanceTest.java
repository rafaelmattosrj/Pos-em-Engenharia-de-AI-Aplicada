package com.datasetprep;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class DataRelevanceTest {

    @Test
    void candidatoComOs4CriteriosVerdadeirosEAceito() {
        var r = DataRelevance.avaliarCandidato(Map.of(
                "contemGroundTruth", true, "producaoReal", true, "cobreVariacao", true, "passaCompliance", true));
        assertThat(r.aceito()).isTrue();
        assertThat(r.criteriosFalhos()).isEmpty();
    }

    @Test
    void candidatoSemGroundTruthERejeitado() {
        var r = DataRelevance.avaliarCandidato(Map.of(
                "contemGroundTruth", false, "producaoReal", true, "cobreVariacao", true, "passaCompliance", true));
        assertThat(r.aceito()).isFalse();
        assertThat(r.criteriosFalhos()).containsExactly("contemGroundTruth");
    }

    @Test
    void candidatoQueFalhaComplianceERejeitadoMesmoComOsOutros3Verdadeiros() {
        var r = DataRelevance.avaliarCandidato(Map.of(
                "contemGroundTruth", true, "producaoReal", true, "cobreVariacao", true, "passaCompliance", false));
        assertThat(r.aceito()).isFalse();
        assertThat(r.criteriosFalhos()).containsExactly("passaCompliance");
    }

    @Test
    void aplicacaoAos7CandidatosReaisDaAmplitudeSeguros() {
        Map<String, Boolean> esperadoAceito = Map.of(
                "orcamento-oficina", true,
                "boletim-ocorrencia", false,
                "foto-veiculo-danificado", false,
                "transcricao-ligacao", false,
                "recibo-medico", true,
                "prontuario-medico-completo", false,
                "cadastro-beneficiarios", false);

        for (var candidato : DataRelevance.CANDIDATOS) {
            var resultado = DataRelevance.avaliarCandidato(candidato.criterios());
            assertThat(resultado.aceito())
                    .as("candidato %s", candidato.id())
                    .isEqualTo(esperadoAceito.get(candidato.id()));
        }
    }

    @Test
    void exatamente2Dos7CandidatosSaoAceitos() {
        List<String> aceitos = DataRelevance.CANDIDATOS.stream()
                .filter(c -> DataRelevance.avaliarCandidato(c.criterios()).aceito())
                .map(DataRelevance.Candidato::id)
                .sorted()
                .toList();
        assertThat(aceitos).containsExactly("orcamento-oficina", "recibo-medico");
    }
}
