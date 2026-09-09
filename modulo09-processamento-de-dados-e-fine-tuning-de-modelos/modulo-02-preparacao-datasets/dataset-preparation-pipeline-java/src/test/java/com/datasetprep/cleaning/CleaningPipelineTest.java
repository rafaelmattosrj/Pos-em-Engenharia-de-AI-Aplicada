package com.datasetprep.cleaning;

import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class CleaningPipelineTest {

    @Test
    void pipelineCompletoProduzDatasetFinalMenorQueOOriginalComDiversidadeMaior() {
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        var resultado = CleaningPipeline.limparEBalancear(dataset, Balancing.ALPHA_TEMPERATURA,
                Map.of("amplitude-auto", 20, "amplitude-saude-empresarial", 14));

        assertThat(resultado.duplicatasRemovidas()).isEqualTo(3);
        assertThat(resultado.fin()).isLessThan(resultado.original());
        assertThat(resultado.relatorioPorCaso().get("amplitude-auto").nEfetivoDepois())
                .isGreaterThan(resultado.relatorioPorCaso().get("amplitude-auto").nEfetivoAntes());
        assertThat(resultado.relatorioPorCaso().get("amplitude-saude-empresarial").nEfetivoDepois())
                .isGreaterThan(resultado.relatorioPorCaso().get("amplitude-saude-empresarial").nEfetivoAntes());
        assertThat(resultado.exemplosFinal()).anyMatch(e -> e.metadata().caso().equals("amplitude-auto"));
        assertThat(resultado.exemplosFinal()).anyMatch(e -> e.metadata().caso().equals("amplitude-saude-empresarial"));
    }
}
