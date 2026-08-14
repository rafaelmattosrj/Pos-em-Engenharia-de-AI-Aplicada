package com.trialforge.gateway;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class AuditTrailTest {

    @Test
    void registrarAcrescentaLinhaComTimestamp(@TempDir Path tempDir) throws IOException {
        Path caminho = tempDir.resolve("audit-trail.jsonl");
        AuditTrail trilha = new AuditTrail(caminho.toString());

        trilha.registrar(Map.of("id_requisicao", "req-1", "status_final", "aprovado"));
        trilha.registrar(Map.of("id_requisicao", "req-2", "status_final", "rejeitado"));

        List<String> linhas = Files.readAllLines(caminho, StandardCharsets.UTF_8);
        assertThat(linhas).hasSize(2);
        assertThat(linhas.get(0)).contains("\"id_requisicao\":\"req-1\"").contains("\"timestamp\"");
    }

    // Grava exatamente o roteiro de 5 requisições que o próprio Main.java
    // executa (rotina / paráfrase-cache-hit / síntese-csr /
    // tema-diferente-esgota-agentic / protocolo-1ª-iteração), pra exercitar
    // verificarTrilhaAuditoria sem precisar de um Ollama real.
    private void escreverTrilhaDemo(AuditTrail trilha, String modeloCaro) throws IOException {
        List<Map<String, Object>> registros = List.of(
                mapa("id_requisicao", "req-1", "cache_hit", false, "modelo_usado", "gemma4:e2b",
                        "indice_usado", "icf", "iteracoes_agentic", 1, "esgotou_agentic", false,
                        "confianca_rag", 0.9, "gate_acionado", false, "aprovado", true, "status_final", "aprovado"),
                mapa("id_requisicao", "req-2", "cache_hit", true, "similaridade_cache", 0.9,
                        "status_final", "respondido_via_cache"),
                mapa("id_requisicao", "req-3", "status_final", "aguardando_aprovacao", "motivo_gate", "síntese de CSR"),
                mapa("id_requisicao", "req-3", "cache_hit", false, "modelo_usado", modeloCaro,
                        "indice_usado", "csr", "iteracoes_agentic", 1, "esgotou_agentic", false,
                        "confianca_rag", 0.95, "gate_acionado", true, "aprovado", true, "status_final", "aprovado"),
                mapa("id_requisicao", "req-4", "status_final", "aguardando_aprovacao", "motivo_gate", "confiança baixa"),
                mapa("id_requisicao", "req-4", "cache_hit", false, "modelo_usado", "gemma4:e2b",
                        "indice_usado", "csr", "iteracoes_agentic", 3, "esgotou_agentic", true,
                        "confianca_rag", 0.5, "gate_acionado", true, "aprovado", true, "status_final", "aprovado"),
                mapa("id_requisicao", "req-5", "cache_hit", false, "modelo_usado", "gemma4:e2b",
                        "indice_usado", "protocolo", "iteracoes_agentic", 1, "esgotou_agentic", false,
                        "confianca_rag", 0.85, "gate_acionado", false, "aprovado", true, "status_final", "aprovado"));

        for (Map<String, Object> registro : registros) {
            trilha.registrar(registro);
        }
    }

    private static Map<String, Object> mapa(Object... kv) {
        Map<String, Object> mapa = new LinkedHashMap<>();
        for (int i = 0; i < kv.length; i += 2) {
            mapa.put((String) kv[i], kv[i + 1]);
        }
        return mapa;
    }

    @Test
    void verificarTrilhaAuditoriaCaminhoFeliz(@TempDir Path tempDir) throws IOException {
        AuditTrail trilha = new AuditTrail(tempDir.resolve("audit-trail.jsonl").toString());
        escreverTrilhaDemo(trilha, "gemma4:latest");

        trilha.verificarTrilhaAuditoria("gemma4:latest", 0.7, null); // não deve lançar
    }

    @Test
    void verificarTrilhaAuditoriaFalhaComTrilhaIncompleta(@TempDir Path tempDir) throws IOException {
        AuditTrail trilha = new AuditTrail(tempDir.resolve("audit-trail.jsonl").toString());
        trilha.registrar(Map.of("id_requisicao", "req-1", "status_final", "aprovado"));

        assertThatThrownBy(() -> trilha.verificarTrilhaAuditoria("gemma4:latest", 0.7, null))
                .isInstanceOf(IllegalStateException.class);
    }
}
