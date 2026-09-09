package com.finetuningapi;

import org.assertj.core.data.Offset;
import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class DatasetScalingTest {

    @Test
    void datasetBrutoTem305Exemplos183Auto122Saude() {
        List<DatasetScaling.ExemploDataset> bruto = DatasetScaling.gerarDatasetBruto();
        assertThat(bruto).hasSize(305);
        assertThat(bruto.stream().filter(e -> e.caso().equals("amplitude-auto")).count()).isEqualTo(183);
        assertThat(bruto.stream().filter(e -> e.caso().equals("amplitude-saude-empresarial")).count()).isEqualTo(122);
    }

    @Test
    void nenhumIdSeRepeteNoBruto() {
        List<DatasetScaling.ExemploDataset> bruto = DatasetScaling.gerarDatasetBruto();
        long idsUnicos = bruto.stream().map(DatasetScaling.ExemploDataset::id).distinct().count();
        assertThat(idsUnicos).isEqualTo(bruto.size());
    }

    @Test
    void pipelineReduz305Para300Dedup300Para200Balanceado() {
        List<DatasetScaling.ExemploDataset> bruto = DatasetScaling.gerarDatasetBruto();
        Map<String, Integer> alvos = new LinkedHashMap<>();
        alvos.put("amplitude-auto", 120);
        alvos.put("amplitude-saude-empresarial", 80);
        var resultado = DatasetScaling.limparEBalancear(bruto, alvos);

        assertThat(resultado.original()).isEqualTo(305);
        assertThat(resultado.aposDedup()).isEqualTo(300);
        assertThat(resultado.total()).isEqualTo(200);

        assertThat(resultado.exemplosFinal().stream().filter(e -> e.caso().equals("amplitude-auto")).count()).isEqualTo(120);
        assertThat(resultado.exemplosFinal().stream().filter(e -> e.caso().equals("amplitude-saude-empresarial")).count()).isEqualTo(80);

        var rAuto = resultado.relatorioPorCaso().get("amplitude-auto");
        var rSaude = resultado.relatorioPorCaso().get("amplitude-saude-empresarial");
        assertThat(rAuto.nEfetivoDepois()).isGreaterThan(rAuto.nEfetivoAntes());
        assertThat(rSaude.nEfetivoDepois()).isGreaterThan(rSaude.nEfetivoAntes());
        assertThat(rAuto.nEfetivoAntes()).isCloseTo(5.160, Offset.offset(0.01));
        assertThat(rAuto.nEfetivoDepois()).isCloseTo(5.723, Offset.offset(0.01));
        assertThat(rSaude.nEfetivoAntes()).isCloseTo(4.140, Offset.offset(0.01));
        assertThat(rSaude.nEfetivoDepois()).isCloseTo(4.706, Offset.offset(0.01));
    }

    @Test
    void nenhumaFonteAlemDoNecessarioAlemDaCapacidadeReal() {
        List<DatasetScaling.ExemploDataset> bruto = DatasetScaling.gerarDatasetBruto();
        Map<String, Integer> alvos = new LinkedHashMap<>();
        alvos.put("amplitude-auto", 120);
        alvos.put("amplitude-saude-empresarial", 80);
        var resultado = DatasetScaling.limparEBalancear(bruto, alvos);
        var rAuto = resultado.relatorioPorCaso().get("amplitude-auto");
        rAuto.contagensDepois().forEach((fonte, n) -> assertThat(n).isLessThanOrEqualTo(rAuto.contagensAntes().get(fonte)));
    }
}
