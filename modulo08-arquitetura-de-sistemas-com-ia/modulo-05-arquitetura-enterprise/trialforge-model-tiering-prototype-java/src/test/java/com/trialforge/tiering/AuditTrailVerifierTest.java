package com.trialforge.tiering;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class AuditTrailVerifierTest {

    private final ObjectMapper mapper = new ObjectMapper();

    private JsonNode registro(Map<String, Object> campos) {
        return mapper.valueToTree(campos);
    }

    /** Réplica exata da sequência dos 4 registros esperados em main() do original. */
    @Test
    void verificar_sequenciaCorretaDosQuatroRegistros_todasChecagensPassam() {
        Map<String, Object> r1 = new LinkedHashMap<>();
        r1.put("tier_usado", "Tier 1");
        r1.put("escalou_cascata", false);
        r1.put("confianca_resposta", 1.0);

        Map<String, Object> r2 = new LinkedHashMap<>();
        r2.put("tier_usado", "Tier 2 (escalado)");
        r2.put("escalou_cascata", true);
        r2.put("confianca_resposta", 0.83);

        Map<String, Object> r3 = new LinkedHashMap<>();
        r3.put("tier_usado", "Tier 2");
        r3.put("escalou_cascata", false);
        r3.put("aprovado", true);

        Map<String, Object> r4 = new LinkedHashMap<>();
        r4.put("status_final", "bloqueado_por_orcamento");

        List<JsonNode> linhas = List.of(registro(r1), registro(r2), registro(r3), registro(r4));

        List<AuditTrailVerifier.Checagem> checagens = AuditTrailVerifier.verificar(linhas);

        assertThat(checagens).allSatisfy(c -> assertThat(c.ok()).as(c.descricao()).isTrue());
    }

    @Test
    void verificar_menosDeQuatroEntradas_falhaNaPrimeiraChecagem() {
        List<JsonNode> linhas = List.of(registro(Map.of("tier_usado", "Tier 1")));

        List<AuditTrailVerifier.Checagem> checagens = AuditTrailVerifier.verificar(linhas);

        assertThat(checagens.get(0).descricao()).contains("pelo menos 4 entradas");
        assertThat(checagens.get(0).ok()).isFalse();
    }

    @Test
    void verificar_primeiroRegistroEscalou_falhaNaChecagem1() {
        Map<String, Object> r1 = new LinkedHashMap<>();
        r1.put("tier_usado", "Tier 2 (escalado)"); // deveria ser Tier 1, sem escalar
        r1.put("escalou_cascata", true);

        Map<String, Object> vazio = Map.of();
        List<JsonNode> linhas = List.of(registro(r1), registro(vazio), registro(vazio), registro(vazio));

        List<AuditTrailVerifier.Checagem> checagens = AuditTrailVerifier.verificar(linhas);

        assertThat(checagens.get(1).ok()).isFalse();
    }

    @Test
    void verificar_usaSomenteAsUltimasQuatroEntradas() {
        // 6 entradas no total: as 2 primeiras são "lixo" de execuções anteriores,
        // só as últimas 4 devem ser verificadas (mesma disciplina do original).
        Map<String, Object> lixo = new LinkedHashMap<>();
        lixo.put("tier_usado", "qualquer coisa");

        Map<String, Object> r1 = new LinkedHashMap<>();
        r1.put("tier_usado", "Tier 1");
        r1.put("escalou_cascata", false);
        r1.put("confianca_resposta", 1.0);

        Map<String, Object> r2 = new LinkedHashMap<>();
        r2.put("tier_usado", "Tier 2 (escalado)");
        r2.put("escalou_cascata", true);
        r2.put("confianca_resposta", 0.83);

        Map<String, Object> r3 = new LinkedHashMap<>();
        r3.put("tier_usado", "Tier 2");
        r3.put("escalou_cascata", false);
        r3.put("aprovado", true);

        Map<String, Object> r4 = new LinkedHashMap<>();
        r4.put("status_final", "bloqueado_por_orcamento");

        List<JsonNode> linhas = List.of(registro(lixo), registro(lixo), registro(r1), registro(r2), registro(r3), registro(r4));

        List<AuditTrailVerifier.Checagem> checagens = AuditTrailVerifier.verificar(linhas);

        assertThat(checagens).allSatisfy(c -> assertThat(c.ok()).as(c.descricao()).isTrue());
    }
}
