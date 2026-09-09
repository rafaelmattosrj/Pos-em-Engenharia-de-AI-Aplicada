package com.finetuningapi;

import java.io.IOException;
import java.util.Map;
import java.util.Set;
import java.util.function.Consumer;

/**
 * Automacao do processo completo de fine-tuning via API (Modulo 3.4):
 * upload, criacao de job com trava de confirmacao explicita (incidente real
 * do Modulo 3.3: hiperparametro invalido nao gera erro rapido na Vertex AI),
 * acompanhamento sozinho ate o job terminar, com backoff exponencial e retry
 * limitado por falha transiente de rede.
 */
public final class FineTuningAutomation {

    private static final Set<String> ESTADOS_TERMINAIS = Set.of(
            "JOB_STATE_SUCCEEDED", "JOB_STATE_FAILED", "JOB_STATE_CANCELLED");

    private FineTuningAutomation() {
    }

    public static String montarComandoUpload(String caminhoLocal, String uriGcs) {
        if (!caminhoLocal.endsWith(".jsonl")) throw new IllegalArgumentException("dataset precisa ser .jsonl");
        if (!uriGcs.startsWith("gs://")) throw new IllegalArgumentException("destino precisa ser um URI gs://");
        return "gsutil cp \"" + caminhoLocal + "\" \"" + uriGcs + "\"";
    }

    public static void exigirConfirmacao(Boolean confirmar) {
        if (confirmar == null || !confirmar) {
            throw new IllegalStateException(
                    "criarJobFineTuning bloqueado: passe confirmar=true explicitamente pra criar job de verdade. "
                            + "Trava adicionada depois do incidente do Módulo 3.3, onde hiperparâmetro inválido criou job real sem aviso.");
        }
    }

    public static long calcularProximoIntervalo(long atualMs, double fator, long maximoMs) {
        return Math.min(Math.round(atualMs * fator), maximoMs);
    }

    /** Consulta com retry limitado: falha transiente de rede nao derruba uma automacao de ~1h40min sozinha. */
    public interface ConsultarComRetryFn {
        Map<String, Object> consultar(String nomeJob) throws Exception;
    }

    public interface EsperarFn {
        void esperar(long ms) throws InterruptedException;
    }

    public static final EsperarFn ESPERAR_REAL = ms -> Thread.sleep(ms);

    public static Map<String, Object> consultarComRetry(ConsultarComRetryFn consultarFn, String nomeJob,
                                                          int tentativas, long atrasoMs, EsperarFn esperarFn) throws Exception {
        Exception ultimoErro = null;
        for (int tentativa = 1; tentativa <= tentativas; tentativa++) {
            try {
                return consultarFn.consultar(nomeJob);
            } catch (Exception erro) {
                ultimoErro = erro;
                if (tentativa < tentativas) esperarFn.esperar(atrasoMs);
            }
        }
        throw ultimoErro;
    }

    public record OpcoesAcompanhamento(
            long intervaloInicialMs, double fatorBackoff, long intervaloMaximoMs,
            Consumer<Map<String, Object>> aoAtualizar, int tentativasConsulta, long atrasoRetryMs) {

        public static OpcoesAcompanhamento padrao() {
            return new OpcoesAcompanhamento(5000, 1.5, 60000, job -> {
            }, 3, 3000);
        }
    }

    public static Map<String, Object> acompanharAteFinalizar(String nomeJob, ConsultarComRetryFn consultarFn,
                                                              EsperarFn esperarFn, OpcoesAcompanhamento opcoes) throws Exception {
        long intervalo = opcoes.intervaloInicialMs();
        Map<String, Object> job = consultarComRetry(consultarFn, nomeJob, opcoes.tentativasConsulta(), opcoes.atrasoRetryMs(), esperarFn);
        opcoes.aoAtualizar().accept(job);

        while (!ESTADOS_TERMINAIS.contains(job.get("state"))) {
            esperarFn.esperar(intervalo);
            intervalo = calcularProximoIntervalo(intervalo, opcoes.fatorBackoff(), opcoes.intervaloMaximoMs());
            job = consultarComRetry(consultarFn, nomeJob, opcoes.tentativasConsulta(), opcoes.atrasoRetryMs(), esperarFn);
            opcoes.aoAtualizar().accept(job);
        }
        return job;
    }

    public interface UploadFn {
        String executar(String caminhoLocal, String uriGcs);
    }

    public interface CriarJobFn {
        Map<String, Object> criar(Map<String, Object> config, boolean confirmar) throws Exception;
    }

    public interface AcompanharFn {
        Map<String, Object> acompanhar(String nomeJob, Consumer<Map<String, Object>> aoAtualizar) throws Exception;
    }

    public record ResultadoAutomacao(Map<String, Object> jobCriado, Map<String, Object> jobFinal) {
    }

    /** Encadeia validar -> subir -> criar job -> acompanhar, falhando rapido em qualquer passo. */
    public static ResultadoAutomacao automatizarFineTuning(
            HyperparameterValidator.Hiperparametros hiper, String caminhoLocal, String uriDataset,
            boolean confirmar, UploadFn uploadFn, CriarJobFn criarJobFn, AcompanharFn acompanharFn,
            Consumer<Map<String, Object>> aoAtualizar) throws Exception {
        HyperparameterValidator.validar(hiper);
        uploadFn.executar(caminhoLocal, uriDataset);
        Map<String, Object> jobCriado = criarJobFn.criar(Map.of(), confirmar);
        Map<String, Object> jobFinal = acompanharFn.acompanhar(String.valueOf(jobCriado.get("name")), aoAtualizar);
        return new ResultadoAutomacao(jobCriado, jobFinal);
    }

    public static Map<String, Object> criarJobReal(VertexAiJobClient client, Map<String, Object> corpo, Boolean confirmar)
            throws IOException, InterruptedException {
        exigirConfirmacao(confirmar);
        return client.criarJob(corpo);
    }
}
