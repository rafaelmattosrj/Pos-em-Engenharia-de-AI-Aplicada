package com.trialforge.reactagent;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.Map;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre os mesmos 6 casos + 1 caso de parâmetro inválido de
 * rodarTestesFerramenta() em react-agent-prototype.js / react_agent_prototype.py
 * (via {@link CasosTesteFerramenta}, a mesma lista usada por {@link Main}).
 */
class ExecutorFerramentaTest {

    static Stream<CasosTesteFerramenta.Caso> casos() {
        return CasosTesteFerramenta.CASOS.stream();
    }

    @ParameterizedTest(name = "tema=\"{0}\" jurisdicao={1} -> achou esperado={2}")
    @MethodSource("casos")
    void executarBuscaClausula_seguemOsCasosDeReferencia(CasosTesteFerramenta.Caso caso) {
        ExecutorFerramenta.ResultadoBusca resultado =
                ExecutorFerramenta.executarBuscaClausula(Map.of("tema", caso.tema(), "jurisdicao", caso.jurisdicao()));

        boolean achou = resultado.texto() != null;
        assertThat(achou).as("achou clausula para tema=%s jurisdicao=%s", caso.tema(), caso.jurisdicao())
                .isEqualTo(caso.esperaAchar());
    }

    @Test
    void parametroMalFormado_naoEstouraExcecao_viraAvisoDeFalhaPropria() {
        ExecutorFerramenta.ResultadoBusca resultado =
                ExecutorFerramenta.executarBuscaClausula(Map.of("tema", "x", "jurisdicicao", "ANVISA"));

        assertThat(resultado.texto()).isNull();
        assertThat(resultado.aviso()).isNotNull().contains("parâmetro inválido ou ausente");
    }

    @Test
    void argumentosNulos_naoEstouraExcecao() {
        ExecutorFerramenta.ResultadoBusca resultado = ExecutorFerramenta.executarBuscaClausula(null);

        assertThat(resultado.texto()).isNull();
        assertThat(resultado.aviso()).isNotNull();
    }

    @Test
    void jurisdicaoCorretaSemMencaoAMenor_naoEncontraClausula() {
        ExecutorFerramenta.ResultadoBusca resultado = ExecutorFerramenta
                .executarBuscaClausula(Map.of("tema", "Consentimento informado de população adulta", "jurisdicao", "ANVISA"));

        assertThat(resultado.texto()).isNull();
        assertThat(resultado.aviso()).contains("não encontrado");
    }
}
