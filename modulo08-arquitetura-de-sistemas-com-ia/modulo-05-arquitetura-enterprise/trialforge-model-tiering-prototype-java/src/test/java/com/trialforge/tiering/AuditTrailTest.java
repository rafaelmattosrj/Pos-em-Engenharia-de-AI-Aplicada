package com.trialforge.tiering;

import com.fasterxml.jackson.databind.JsonNode;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.IOException;
import java.nio.file.Path;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class AuditTrailTest {

    @Test
    void registrar_gravaLinhaJsonComTimestamp(@TempDir Path tempDir) throws IOException {
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit-trail-tiering.jsonl"));

        Map<String, Object> registro = new LinkedHashMap<>();
        registro.put("estudoId", "estudo-A");
        registro.put("status_final", "aprovado");
        auditTrail.registrar(registro);

        List<JsonNode> linhas = auditTrail.lerTodas();
        assertThat(linhas).hasSize(1);
        assertThat(linhas.get(0).path("timestamp").asText()).isNotBlank();
        assertThat(linhas.get(0).path("estudoId").asText()).isEqualTo("estudo-A");
        assertThat(linhas.get(0).path("status_final").asText()).isEqualTo("aprovado");
    }

    @Test
    void registrar_ehAppendOnly(@TempDir Path tempDir) throws IOException {
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("audit-trail-tiering.jsonl"));

        auditTrail.registrar(Map.of("n", 1));
        auditTrail.registrar(Map.of("n", 2));
        auditTrail.registrar(Map.of("n", 3));

        List<JsonNode> linhas = auditTrail.lerTodas();
        assertThat(linhas).hasSize(3);
        assertThat(linhas.get(0).path("n").asInt()).isEqualTo(1);
        assertThat(linhas.get(2).path("n").asInt()).isEqualTo(3);
    }

    @Test
    void lerTodas_arquivoInexistente_retornaListaVazia(@TempDir Path tempDir) throws IOException {
        AuditTrail auditTrail = new AuditTrail(tempDir.resolve("nao-existe.jsonl"));

        assertThat(auditTrail.lerTodas()).isEmpty();
    }
}
