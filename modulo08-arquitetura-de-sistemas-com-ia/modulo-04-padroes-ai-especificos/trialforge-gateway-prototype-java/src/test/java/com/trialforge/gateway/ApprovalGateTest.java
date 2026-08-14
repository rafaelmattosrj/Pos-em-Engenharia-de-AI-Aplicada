package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

import static org.assertj.core.api.Assertions.assertThat;

class ApprovalGateTest {

    private static ApprovalGate criarGate(String entradaSimulada, ByteArrayOutputStream saida) {
        ByteArrayInputStream entrada = new ByteArrayInputStream(entradaSimulada.getBytes(StandardCharsets.UTF_8));
        return new ApprovalGate(entrada, new PrintStream(saida, true, StandardCharsets.UTF_8));
    }

    @Test
    void aprovaComS() throws Exception {
        ByteArrayOutputStream saida = new ByteArrayOutputStream();
        ApprovalGate gate = criarGate("s\n", saida);

        boolean aprovado = gate.pedirAprovacaoHumana("rascunho de teste");

        assertThat(aprovado).isTrue();
        assertThat(saida.toString(StandardCharsets.UTF_8)).contains("rascunho de teste");
    }

    @Test
    void rejeitaComN() throws Exception {
        ByteArrayOutputStream saida = new ByteArrayOutputStream();
        ApprovalGate gate = criarGate("n\n", saida);

        assertThat(gate.pedirAprovacaoHumana("rascunho")).isFalse();
    }

    @Test
    void eofRejeita() throws Exception {
        ByteArrayOutputStream saida = new ByteArrayOutputStream();
        ApprovalGate gate = criarGate("", saida);

        assertThat(gate.pedirAprovacaoHumana("rascunho")).isFalse();
    }

    // Mesma fila de respostas consumida em sequência por chamadas sucessivas —
    // replica o cenário de entrada não-interativa (`printf "s\ns\n" | ...`)
    // que motivou o bug corrigido nos dois originais (reader único, nunca
    // recriado a cada chamada).
    @Test
    void multiplasChamadasConsomeFilaEmOrdem() throws Exception {
        ByteArrayOutputStream saida = new ByteArrayOutputStream();
        ApprovalGate gate = criarGate("s\nn\ns\n", saida);

        assertThat(gate.pedirAprovacaoHumana("1")).isTrue();
        assertThat(gate.pedirAprovacaoHumana("2")).isFalse();
        assertThat(gate.pedirAprovacaoHumana("3")).isTrue();
    }
}
