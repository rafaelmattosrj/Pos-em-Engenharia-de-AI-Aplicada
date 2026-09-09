package com.finetuningapi;

import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class FineTuningAutomationTest {

    @Test
    void montaComandoGsutilCpCorreto() {
        String comando = FineTuningAutomation.montarComandoUpload("dataset.jsonl", "gs://bucket/dataset.jsonl");
        assertThat(comando).isEqualTo("gsutil cp \"dataset.jsonl\" \"gs://bucket/dataset.jsonl\"");
    }

    @Test
    void rejeitaDatasetQueNaoEhJsonl() {
        assertThatThrownBy(() -> FineTuningAutomation.montarComandoUpload("dataset.json", "gs://bucket/x.jsonl"))
                .hasMessageContaining("jsonl");
    }

    @Test
    void rejeitaDestinoQueNaoEhGs() {
        assertThatThrownBy(() -> FineTuningAutomation.montarComandoUpload("dataset.jsonl", "/local/path"))
                .hasMessageContaining("gs://");
    }

    @Test
    void bloqueiaCriacaoDeJobSemConfirmarTrue() {
        assertThatThrownBy(() -> FineTuningAutomation.exigirConfirmacao(null)).hasMessageContaining("confirmar");
        assertThatThrownBy(() -> FineTuningAutomation.exigirConfirmacao(false)).hasMessageContaining("confirmar");
    }

    @Test
    void permitePassarQuandoConfirmarTrueEhExplicito() {
        FineTuningAutomation.exigirConfirmacao(true);
    }

    @Test
    void primeiroBackoffAplicaOFator() {
        assertThat(FineTuningAutomation.calcularProximoIntervalo(5000, 1.5, 60000)).isEqualTo(7500);
    }

    @Test
    void backoffNuncaUltrapassaOTeto() {
        assertThat(FineTuningAutomation.calcularProximoIntervalo(50000, 1.5, 60000)).isEqualTo(60000);
    }

    @Test
    void reconheceEstadoJaTerminalNaPrimeiraConsultaSemEsperar() throws Exception {
        AtomicInteger chamadasConsulta = new AtomicInteger();
        AtomicInteger chamadasEspera = new AtomicInteger();
        var job = FineTuningAutomation.acompanharAteFinalizar("job-falso",
                nomeJob -> {
                    chamadasConsulta.incrementAndGet();
                    return Map.of("state", "JOB_STATE_SUCCEEDED");
                },
                ms -> chamadasEspera.incrementAndGet(),
                FineTuningAutomation.OpcoesAcompanhamento.padrao());
        assertThat(job.get("state")).isEqualTo("JOB_STATE_SUCCEEDED");
        assertThat(chamadasConsulta.get()).isEqualTo(1);
        assertThat(chamadasEspera.get()).isZero();
    }

    @Test
    void esperaEConsultaDeNovoEnquantoOJobEstaRunning() throws Exception {
        String[] sequencia = {"JOB_STATE_PENDING", "JOB_STATE_RUNNING", "JOB_STATE_RUNNING", "JOB_STATE_SUCCEEDED"};
        AtomicInteger indice = new AtomicInteger();
        AtomicInteger chamadasEspera = new AtomicInteger();
        var job = FineTuningAutomation.acompanharAteFinalizar("job-falso",
                nomeJob -> Map.of("state", sequencia[indice.getAndIncrement()]),
                ms -> chamadasEspera.incrementAndGet(),
                FineTuningAutomation.OpcoesAcompanhamento.padrao());
        assertThat(job.get("state")).isEqualTo("JOB_STATE_SUCCEEDED");
        assertThat(indice.get()).isEqualTo(sequencia.length);
        assertThat(chamadasEspera.get()).isEqualTo(sequencia.length - 1);
    }

    @Test
    void consultaComRetryAbsorveFalhaTransiente() throws Exception {
        AtomicInteger tentativas = new AtomicInteger();
        Map<String, Object> resultado = FineTuningAutomation.consultarComRetry(nomeJob -> {
            if (tentativas.getAndIncrement() < 2) throw new RuntimeException("falha transiente");
            return Map.of("state", "JOB_STATE_SUCCEEDED");
        }, "job-falso", 3, 1, ms -> {
        });
        assertThat(resultado.get("state")).isEqualTo("JOB_STATE_SUCCEEDED");
        assertThat(tentativas.get()).isEqualTo(3);
    }

    @Test
    void hiperparametroInvalidoFalhaAntesDeCriarJob() {
        assertThatThrownBy(() -> HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(0, 5.0)))
                .hasMessageContaining("epochCount");
    }
}
