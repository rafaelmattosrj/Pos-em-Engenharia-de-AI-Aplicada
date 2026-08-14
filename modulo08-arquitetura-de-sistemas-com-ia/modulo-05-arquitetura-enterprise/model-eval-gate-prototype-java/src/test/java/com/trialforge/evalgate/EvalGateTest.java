package com.trialforge.evalgate;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.offset;

class EvalGateTest {

    // ---------- similaridadeCosseno: lógica pura ----------

    @Test
    void similaridadeCosseno_vetoresIdenticos_dá1() {
        assertThat(EvalGate.similaridadeCosseno(new double[]{1, 0, 0}, new double[]{1, 0, 0}))
                .isCloseTo(1.0, offset(1e-9));
    }

    @Test
    void similaridadeCosseno_vetoresOrtogonais_dá0() {
        assertThat(EvalGate.similaridadeCosseno(new double[]{1, 0, 0}, new double[]{0, 1, 0}))
                .isCloseTo(0.0, offset(1e-9));
    }

    @Test
    void similaridadeCosseno_vetoresOpostos_dáMenos1() {
        assertThat(EvalGate.similaridadeCosseno(new double[]{1, 0}, new double[]{-1, 0}))
                .isCloseTo(-1.0, offset(1e-9));
    }

    // ---------- decidirPromocao: lógica pura de decisão do gate ----------

    @Test
    void decidirPromocao_scoreIgual_promove() {
        assertThat(EvalGate.decidirPromocao(0.85, 0.85, 0.02)).isTrue();
    }

    @Test
    void decidirPromocao_dentroDaTolerancia_promove() {
        // baseline 0.85, candidato 0.84 -> diferença -0.01, tolerância 0.02 -> promove
        assertThat(EvalGate.decidirPromocao(0.85, 0.84, 0.02)).isTrue();
    }

    @Test
    void decidirPromocao_exatamenteNoLimiteDaTolerancia_promove() {
        // diferença == -tolerancia é o limite inclusivo (>=), igual ao original.
        // 0.5/0.25 são exatamente representáveis em ponto flutuante binário, então
        // a diferença bate exatamente com -tolerancia, sem ruído de arredondamento.
        assertThat(EvalGate.decidirPromocao(0.5, 0.25, 0.25)).isTrue();
    }

    @Test
    void decidirPromocao_alemDaTolerancia_bloqueia() {
        // baseline 0.85, candidato 0.80 -> diferença -0.05, tolerância 0.02 -> bloqueia
        assertThat(EvalGate.decidirPromocao(0.85, 0.80, 0.02)).isFalse();
    }

    @Test
    void decidirPromocao_candidatoMelhorQueBaseline_promove() {
        assertThat(EvalGate.decidirPromocao(0.80, 0.95, 0.02)).isTrue();
    }

    // ---------- avaliarCandidato: fluxo ponta a ponta com dublê determinístico ----------

    @Test
    void avaliarCandidato_calculaMediaDosScores() throws Exception {
        GoldenItem item1 = new GoldenItem("pergunta 1", "clausula 1", "fonte 1");
        GoldenItem item2 = new GoldenItem("pergunta 2", "clausula 2", "fonte 2");

        FakeOllamaGateway fake = new FakeOllamaGateway()
                .comChat("pergunta 1", "resposta 1")
                .comChat("pergunta 2", "resposta 2")
                .comEmbedding("resposta 1", new double[]{1, 0})
                .comEmbedding("clausula 1", new double[]{1, 0})   // score 1.0
                .comEmbedding("resposta 2", new double[]{1, 0})
                .comEmbedding("clausula 2", new double[]{0, 1});  // score 0.0

        EvalGate gate = new EvalGate(fake);
        double media = gate.avaliarCandidato("qualquer-modelo", List.of(item1, item2));

        assertThat(media).isCloseTo(0.5, offset(1e-9)); // (1.0 + 0.0) / 2
    }

    @Test
    void avaliarCandidatoSemContexto_naoEnviaClausulaNoPrompt() throws Exception {
        // gerarRespostaSemContexto não deve incluir a cláusula no prompt do usuário —
        // simula o bug de RAG/config perdendo o contexto (Cenário 2 do original).
        GoldenItem item = new GoldenItem("pergunta sem contexto", "clausula que não deve aparecer", "fonte");

        FakeOllamaGateway fake = new FakeOllamaGateway() {
            @Override
            public String chat(String modelo, String systemPrompt, String userPrompt) {
                assertThat(userPrompt).doesNotContain("clausula que não deve aparecer");
                return "resposta sem contexto";
            }
        };
        fake.comEmbedding("resposta sem contexto", new double[]{1, 0})
                .comEmbedding("clausula que não deve aparecer", new double[]{0, 1});

        EvalGate gate = new EvalGate(fake);
        double media = gate.avaliarCandidatoSemContexto("gemma4:e2b", List.of(item));

        assertThat(media).isCloseTo(0.0, offset(1e-9));
    }

    @Test
    void gerarResposta_incluiClausulaEFonteNoPrompt() throws Exception {
        FakeOllamaGateway fake = new FakeOllamaGateway() {
            @Override
            public String chat(String modelo, String systemPrompt, String userPrompt) {
                assertThat(userPrompt).contains("minha-clausula").contains("minha-fonte").contains("minha-pergunta");
                return "ok";
            }
        };
        EvalGate gate = new EvalGate(fake);

        String resposta = gate.gerarResposta("modelo", "minha-pergunta", "minha-clausula", "minha-fonte");

        assertThat(resposta).isEqualTo("ok");
    }
}
