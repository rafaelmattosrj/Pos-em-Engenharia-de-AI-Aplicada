package com.trialforge.agentcomponents;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre o mesmo cenario de testarFerramenta() em agent-components-demo.js / .py:
 * faixa etaria com menor de idade encontra clausula; faixa so de adultos, nao.
 */
class FerramentasTest {

    @Test
    void faixaComMenorDeIdade_encontraClausula() {
        Ferramentas.ResultadoClausula resultado = Ferramentas.buscarClausulaAssentimento(List.of(12, 15, 17));

        assertThat(resultado.texto()).isNotNull().contains("RDC ANVISA 466/2012");
        assertThat(resultado.fonte()).isEqualTo("RDC ANVISA 466/2012, Art. 4º");
        assertThat(resultado.aviso()).isNull();
    }

    @Test
    void faixaSoDeAdultos_naoEncontraClausula() {
        Ferramentas.ResultadoClausula resultado = Ferramentas.buscarClausulaAssentimento(List.of(25, 40, 55));

        assertThat(resultado.texto()).isNull();
        assertThat(resultado.fonte()).isNull();
        assertThat(resultado.aviso()).contains("população adulta");
    }

    @Test
    void faixaMista_comAoMenosUmMenor_encontraClausula() {
        Ferramentas.ResultadoClausula resultado = Ferramentas.buscarClausulaAssentimento(List.of(17, 30, 45));

        assertThat(resultado.texto()).isNotNull();
    }
}
