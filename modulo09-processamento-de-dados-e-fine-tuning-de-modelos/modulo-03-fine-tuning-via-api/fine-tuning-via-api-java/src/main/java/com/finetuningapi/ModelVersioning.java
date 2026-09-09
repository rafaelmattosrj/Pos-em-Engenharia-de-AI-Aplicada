package com.finetuningapi;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.LinkedHashMap;
import java.util.Locale;
import java.util.Map;

/**
 * Versionamento e documentacao de modelo fine-tunado (Modulo 3.5): hash de
 * conteudo do dataset (SHA-256, mesmo principio de git/Docker), ficha de
 * versionamento a partir do job real, e model card em Markdown.
 */
public final class ModelVersioning {

    public record FaixaCustoGpu(double consumerMin, double consumerMax, double h100Min, double h100Max) {
        public static final FaixaCustoGpu PADRAO = new FaixaCustoGpu(0.40, 0.80, 2.50, 4.00);
    }

    /** Taxa real apurada no relatorio de billing por SKU de agosto/2026 (R$0,00002909/unidade). */
    public static final double TAXA_REAL_POR_UNIDADE = 0.00002909;

    private ModelVersioning() {
    }

    public static String calcularHashDataset(Path caminhoArquivo) throws IOException {
        byte[] conteudo = Files.readAllBytes(caminhoArquivo);
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(conteudo);
            return String.format("%064x", new BigInteger(1, hash));
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException(e);
        }
    }

    public record CustoEstimado(double duracaoSegundos, String duracaoFormatada,
                                 double consumerMinUsd, double consumerMaxUsd,
                                 double h100MinUsd, double h100MaxUsd) {
    }

    public static CustoEstimado calcularCustoEstimado(double duracaoSegundos, FaixaCustoGpu faixa) {
        double duracaoHoras = duracaoSegundos / 3600.0;
        return new CustoEstimado(
                Math.round(duracaoSegundos * 1000) / 1000.0,
                TempoUtil.formatarDuracaoComEspaco(duracaoSegundos),
                arredondar2(duracaoHoras * faixa.consumerMin()),
                arredondar2(duracaoHoras * faixa.consumerMax()),
                arredondar2(duracaoHoras * faixa.h100Min()),
                arredondar2(duracaoHoras * faixa.h100Max()));
    }

    public record CustoReal(long unidades, double custoReais) {
    }

    public static CustoReal calcularCustoReal(Long tokensCobraveis, Integer epochCount) {
        if (tokensCobraveis == null || epochCount == null) return null;
        long unidades = tokensCobraveis * epochCount;
        return new CustoReal(unidades, Math.round(unidades * TAXA_REAL_POR_UNIDADE * 100) / 100.0);
    }

    private static double arredondar2(double v) {
        return Math.round(v * 100) / 100.0;
    }

    public record FichaVersionamento(
            String jobId, String modeloBase, String nomeExibicao, String datasetUri, String datasetHashSha256,
            Integer epochCount, Double learningRateMultiplier, String adapterSize,
            Long exemplosDataset, Long tokensCobraveis,
            String modeloAjustado, String endpoint, String criadoEm, String concluidoEm, String estado,
            CustoEstimado custoEstimado, CustoReal custoReal) {
    }

    @SuppressWarnings("unchecked")
    public static FichaVersionamento gerarFichaVersionamento(Map<String, Object> job, String hashDataset) {
        String jobId = (String) job.get("name");
        if (jobId == null) throw new IllegalArgumentException("job inválido: precisa ter ao menos \"name\"");

        Map<String, Object> tuningSpec = (Map<String, Object>) job.getOrDefault("supervisedTuningSpec", Map.of());
        Map<String, Object> hiper = (Map<String, Object>) tuningSpec.getOrDefault("hyperParameters", Map.of());
        Map<String, Object> stats = (Map<String, Object>) ((Map<String, Object>) job.getOrDefault("tuningDataStats", Map.of()))
                .getOrDefault("supervisedTuningDataStats", Map.of());
        Map<String, Object> tunedModel = (Map<String, Object>) job.getOrDefault("tunedModel", Map.of());

        CustoEstimado custoEstimado = null;
        String createTime = (String) job.get("createTime");
        String endTime = (String) job.get("endTime");
        if (createTime != null && endTime != null) {
            custoEstimado = calcularCustoEstimado(TempoUtil.duracaoSegundos(createTime, endTime), FaixaCustoGpu.PADRAO);
        }

        Long tokensCobraveis = numeroOuNulo(stats.get("totalBillableTokenCount"));
        Integer epochCount = numeroOuNulo(hiper.get("epochCount")) == null ? null : numeroOuNulo(hiper.get("epochCount")).intValue();
        CustoReal custoReal = calcularCustoReal(tokensCobraveis, epochCount);

        return new FichaVersionamento(
                jobId,
                (String) job.get("baseModel"),
                (String) job.get("tunedModelDisplayName"),
                (String) tuningSpec.get("trainingDatasetUri"),
                hashDataset,
                epochCount,
                hiper.get("learningRateMultiplier") == null ? null : ((Number) hiper.get("learningRateMultiplier")).doubleValue(),
                (String) hiper.get("adapterSize"),
                numeroOuNulo(stats.get("tuningDatasetExampleCount")),
                tokensCobraveis,
                (String) tunedModel.get("model"),
                (String) tunedModel.get("endpoint"),
                createTime,
                endTime,
                (String) job.get("state"),
                custoEstimado,
                custoReal);
    }

    private static Long numeroOuNulo(Object valor) {
        return valor == null ? null : ((Number) valor).longValue();
    }

    public static void validarFichaCompleta(FichaVersionamento ficha) {
        Map<String, Object> campos = new LinkedHashMap<>();
        campos.put("jobId", ficha.jobId());
        campos.put("modeloBase", ficha.modeloBase());
        campos.put("datasetUri", ficha.datasetUri());
        campos.put("datasetHashSha256", ficha.datasetHashSha256());
        campos.put("modeloAjustado", ficha.modeloAjustado());
        campos.put("endpoint", ficha.endpoint());
        StringBuilder faltando = new StringBuilder();
        campos.forEach((chave, valor) -> {
            if (valor == null) {
                if (faltando.length() > 0) faltando.append(", ");
                faltando.append(chave);
            }
        });
        if (faltando.length() > 0) {
            throw new IllegalStateException("Ficha de versionamento incompleta, faltam: " + faltando);
        }
    }

    public record JobDpoReal(String jobId, String estado, String duracaoFormatada, int exemplos) {
        public static final JobDpoReal REAL = new JobDpoReal(
                "tuningJobs/3733013646142341120", "JOB_STATE_SUCCEEDED", "17min 32s", 40);
    }

    public static String gerarSecaoDpo() {
        JobDpoReal j = JobDpoReal.REAL;
        return String.format(Locale.US, """

                ## Continuação: preference tuning (DPO)

                O Módulo 3.5 vai além do fine-tuning supervisionado acima e testa preference tuning (DPO) sobre o mesmo modelo base: 40 dos 200 exemplos deste job foram convertidos em pares de preferência (`chosen`/`rejected`) - a extração correta de sempre (`chosen`) contra uma resposta real gerada por um prompt deliberadamente mais fraco (`rejected`, sem exigir JSON estrito). O dataset de preferência resultante está em `preference-dataset-amplitude.jsonl`, nesta mesma pasta.

                - Job: `%s`
                - Estado: %s
                - Duração real: %s (mais rápido que o SFT acima, dataset 5x menor: %d exemplos contra 200)
                - Exemplos: %d pares de preferência""",
                j.jobId(), j.estado(), j.duracaoFormatada(), j.exemplos(), j.exemplos());
    }

    public static String gerarModelCardMarkdown(FichaVersionamento ficha) {
        StringBuilder sb = new StringBuilder();
        sb.append("# Model Card, modelo fine-tunado\n\n");
        sb.append("## Identificação\n");
        sb.append("- Job: ").append(ficha.jobId()).append('\n');
        sb.append("- Modelo ajustado: ").append(ficha.modeloAjustado()).append('\n');
        sb.append("- Endpoint: ").append(ficha.endpoint()).append('\n');
        sb.append("- Estado: ").append(ficha.estado()).append("\n\n");
        sb.append("## Linhagem\n");
        sb.append("- Modelo base: ").append(ficha.modeloBase()).append('\n');
        sb.append("- Dataset de treino: ").append(ficha.datasetUri()).append('\n');
        sb.append("- Hash SHA-256 do dataset: ").append(ficha.datasetHashSha256()).append("\n\n");
        sb.append("## Hiperparâmetros\n");
        sb.append("- Épocas: ").append(ficha.epochCount()).append('\n');
        sb.append("- Taxa de aprendizado (multiplicador): ").append(ficha.learningRateMultiplier()).append('\n');
        sb.append("- Rank do adaptador (LoRA): ").append(ficha.adapterSize()).append("\n\n");
        sb.append("## Estatística do dataset\n");
        sb.append("- Exemplos de treino: ").append(ficha.exemplosDataset()).append('\n');
        sb.append("- Tokens cobráveis no total: ").append(ficha.tokensCobraveis()).append("\n\n");
        sb.append("## Linha do tempo\n");
        sb.append("- Criado em: ").append(ficha.criadoEm()).append('\n');
        sb.append("- Concluído em: ").append(ficha.concluidoEm()).append("\n\n");

        if (ficha.custoEstimado() != null) {
            CustoEstimado custo = ficha.custoEstimado();
            sb.append("## Custo real\n");
            sb.append("- Duração real do job: ").append(custo.duracaoFormatada()).append('\n');
            String notaFinal = "Nota: a Vertex AI cobra por token de treino, não por hora de GPU alugada; "
                    + "a faixa de GPU acima é referência de mercado pra comparar com o custo de rodar o mesmo "
                    + "tipo de treino (LoRA) em infraestrutura própria, não a fatura real deste job.";
            if (ficha.custoReal() != null) {
                sb.append(String.format(Locale.US,
                        "- **Custo real, conferido no billing do Google Cloud (28/08/2026)**: R$%s (%s tokens faturáveis × %d épocas = %s unidades cobradas, à taxa real de R$0,00002909/unidade apurada no relatório de billing por SKU de agosto/2026)\n",
                        formatarBrl(ficha.custoReal().custoReais()), formatarMilhar(ficha.tokensCobraveis()),
                        ficha.epochCount(), formatarMilhar(ficha.custoReal().unidades())));
                notaFinal = "Nota: a Vertex AI cobra por token de treino, não por hora de GPU alugada; a faixa de "
                        + "GPU acima é referência de mercado pra comparar com o custo de rodar o mesmo tipo de treino "
                        + "(LoRA) em infraestrutura própria - o valor real deste job específico é o R$"
                        + formatarBrl(ficha.custoReal().custoReais()) + " conferido no billing, acima.";
            }
            sb.append(String.format(Locale.US, "- Faixa GPU cloud consumer (US$ 0,40-0,80/hora, cheatsheet do Módulo 1.3): US$ %s-%s\n",
                    formatarUsd(custo.consumerMinUsd()), formatarUsd(custo.consumerMaxUsd())));
            sb.append(String.format(Locale.US, "- Faixa GPU cloud H100 (US$ 2,50-4,00/hora, cheatsheet do Módulo 1.3): US$ %s-%s\n",
                    formatarUsd(custo.h100MinUsd()), formatarUsd(custo.h100MaxUsd())));
            sb.append(notaFinal).append("\n\n");
        }

        sb.append("## Nota de validade (ago/2026)\n");
        sb.append("Este model card documenta um job real, rodado com ").append(ficha.modeloBase())
                .append(". O processo -- upload, hiperparâmetro, versionamento -- é o mesmo independente da "
                        + "versão exata do modelo-base. A Google aposenta versões do Gemini com aviso prévio "
                        + "(a família 2.5 tem retirement anunciado pra 16/out/2026); antes de treinar você mesmo, "
                        + "confira em [Vertex AI release notes]"
                        + "(https://docs.cloud.google.com/vertex-ai/generative-ai/docs/release-notes) quais modelos "
                        + "têm suporte a fine-tuning supervisionado no momento.");
        return sb.toString();
    }

    private static String formatarUsd(double valor) {
        return String.format(Locale.US, "%.2f", valor).replace('.', ',');
    }

    private static String formatarBrl(double valor) {
        return String.format(Locale.US, "%.2f", valor).replace('.', ',');
    }

    private static String formatarMilhar(long numero) {
        return String.format(Locale.US, "%,d", numero).replace(',', '.');
    }
}
