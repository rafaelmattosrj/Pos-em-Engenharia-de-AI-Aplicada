package com.finetuningapi;

import java.util.Map;
import java.util.function.Consumer;

/**
 * Extra parte 2/2: sobe o dataset Dolly já preparado, cria job real e
 * acompanha -- porte de dolly-vertex-pipeline.js. Reusa
 * GeminiConverter.converterTexto (saída do Dolly é texto solto, não objeto)
 * e a mesma trava de confirmação/validação de hiperparâmetro/backoff de
 * FineTuningAutomation.
 */
public final class DollyVertexPipeline {

    private DollyVertexPipeline() {
    }

    public record ConfigPipeline(String caminhoLocal, String uriDataset, String baseModel, String displayName,
                                  int epochCount, double learningRateMultiplier) {
    }

    public record ResultadoPipeline(Map<String, Object> jobCriado, Map<String, Object> jobFinal) {
    }

    public static ResultadoPipeline rodarPipeline(
            ConfigPipeline config, boolean confirmar,
            FineTuningAutomation.UploadFn uploadFn, FineTuningAutomation.CriarJobFn criarJobFn,
            FineTuningAutomation.AcompanharFn acompanharFn, Consumer<Map<String, Object>> aoAtualizar) throws Exception {
        HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(config.epochCount(), config.learningRateMultiplier()));
        uploadFn.executar(config.caminhoLocal(), config.uriDataset());
        Map<String, Object> jobCriado = criarJobFn.criar(Map.of(), confirmar);
        Map<String, Object> jobFinal = acompanharFn.acompanhar(String.valueOf(jobCriado.get("name")), aoAtualizar);
        return new ResultadoPipeline(jobCriado, jobFinal);
    }
}
