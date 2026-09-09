package com.finetuningapi;

import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class ModelVersioningTest {

    private static Map<String, Object> jobFalso() {
        Map<String, Object> job = new LinkedHashMap<>();
        job.put("name", "projects/x/locations/y/tuningJobs/123");
        job.put("baseModel", "gemini-2.5-flash");
        job.put("tunedModelDisplayName", "teste");
        job.put("state", "JOB_STATE_SUCCEEDED");
        job.put("createTime", "2026-08-08T00:00:00Z");
        job.put("endTime", "2026-08-08T01:00:00Z");

        Map<String, Object> hiper = new LinkedHashMap<>();
        hiper.put("epochCount", 3);
        hiper.put("learningRateMultiplier", 5.0);
        hiper.put("adapterSize", "ADAPTER_SIZE_FOUR");
        Map<String, Object> tuningSpec = new LinkedHashMap<>();
        tuningSpec.put("trainingDatasetUri", "gs://bucket/dataset.jsonl");
        tuningSpec.put("hyperParameters", hiper);
        job.put("supervisedTuningSpec", tuningSpec);

        Map<String, Object> stats = new LinkedHashMap<>();
        stats.put("tuningDatasetExampleCount", 200);
        stats.put("totalBillableTokenCount", 27353);
        Map<String, Object> tuningDataStats = new LinkedHashMap<>();
        tuningDataStats.put("supervisedTuningDataStats", stats);
        job.put("tuningDataStats", tuningDataStats);

        Map<String, Object> tunedModel = new LinkedHashMap<>();
        tunedModel.put("model", "projects/x/locations/y/models/999");
        tunedModel.put("endpoint", "projects/x/locations/y/endpoints/888");
        job.put("tunedModel", tunedModel);
        return job;
    }

    @Test
    void geraFichaCompletaAPartirDeUmJobBemFormado() {
        var ficha = ModelVersioning.gerarFichaVersionamento(jobFalso(), "hash-de-teste");
        assertThat(ficha.jobId()).isEqualTo("projects/x/locations/y/tuningJobs/123");
        assertThat(ficha.epochCount()).isEqualTo(3);
        assertThat(ficha.datasetHashSha256()).isEqualTo("hash-de-teste");
    }

    @Test
    void rejeitaJobSemName() {
        assertThatThrownBy(() -> ModelVersioning.gerarFichaVersionamento(Map.of(), "hash"))
                .hasMessageContaining("job inválido");
    }

    @Test
    void validacaoAceitaFichaCompleta() {
        var ficha = ModelVersioning.gerarFichaVersionamento(jobFalso(), "hash-de-teste");
        ModelVersioning.validarFichaCompleta(ficha);
    }

    @Test
    void custoEstimadoUsaADuracaoRealDoJob() {
        var ficha = ModelVersioning.gerarFichaVersionamento(jobFalso(), "hash-de-teste");
        assertThat(ficha.custoEstimado().duracaoFormatada()).isEqualTo("60min 0s");
        assertThat(ficha.custoEstimado().consumerMinUsd()).isEqualTo(0.40);
        assertThat(ficha.custoEstimado().consumerMaxUsd()).isEqualTo(0.80);
        assertThat(ficha.custoEstimado().h100MinUsd()).isEqualTo(2.50);
        assertThat(ficha.custoEstimado().h100MaxUsd()).isEqualTo(4.00);
    }

    @Test
    void formataDuracaoRealDe45min42sDoJobDeProducao() {
        double segundos = TempoUtil.duracaoSegundos("2026-08-08T02:38:12.307201Z", "2026-08-08T03:23:54.310390Z");
        assertThat(TempoUtil.formatarDuracaoComEspaco(segundos)).isEqualTo("45min 42s");
    }

    @Test
    void modelCardIncluiHashDoDatasetEEndpoint() {
        var ficha = ModelVersioning.gerarFichaVersionamento(jobFalso(), "hash-de-teste-abc123");
        String markdown = ModelVersioning.gerarModelCardMarkdown(ficha);
        assertThat(markdown).contains("hash-de-teste-abc123")
                .contains("projects/x/locations/y/endpoints/888")
                .contains("projects/x/locations/y/models/999");
    }

    @Test
    void modelCardIncluiCustoRealEFaixaDeGpu() {
        var ficha = ModelVersioning.gerarFichaVersionamento(jobFalso(), "hash-de-teste");
        String markdown = ModelVersioning.gerarModelCardMarkdown(ficha);
        assertThat(markdown).contains("## Custo real")
                .contains("R$2,39")
                .contains("82.059 unidades cobradas")
                .contains("US$ 0,40-0,80")
                .contains("US$ 2,50-4,00");
    }

    @Test
    void custoRealUsaTokensFaturaveisXEpocasXTaxaReal() {
        var ficha = ModelVersioning.gerarFichaVersionamento(jobFalso(), "hash-de-teste");
        assertThat(ficha.custoReal().unidades()).isEqualTo(82059);
        assertThat(ficha.custoReal().custoReais()).isEqualTo(2.39);
    }

    @Test
    void secaoDpoCitaOJobRealDePreferenceTuning() {
        String secao = ModelVersioning.gerarSecaoDpo();
        assertThat(secao).contains("## Continuação: preference tuning (DPO)")
                .contains("tuningJobs/3733013646142341120")
                .contains("17min 32s")
                .contains("40 pares de preferência")
                .contains("preference-dataset-amplitude.jsonl");
    }

    @Test
    void hashDoMesmoArquivoCalculadoDuasVezesEIdentico() throws Exception {
        var caminho = java.nio.file.Path.of(System.getProperty("java.io.tmpdir"), "_teste_hash_versioning.jsonl");
        java.nio.file.Files.writeString(caminho, "{\"a\":1}\n");
        try {
            String hash1 = ModelVersioning.calcularHashDataset(caminho);
            String hash2 = ModelVersioning.calcularHashDataset(caminho);
            assertThat(hash1).isEqualTo(hash2).hasSize(64).matches("^[0-9a-f]+$");
        } finally {
            java.nio.file.Files.deleteIfExists(caminho);
        }
    }
}
