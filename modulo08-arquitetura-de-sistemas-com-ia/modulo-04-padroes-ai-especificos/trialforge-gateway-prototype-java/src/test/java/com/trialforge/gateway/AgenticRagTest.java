package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class AgenticRagTest {

    // Constrói um cenário onde a 1ª iteração (comparação com o tema) fica
    // abaixo do limiar, mas a 2ª (texto completo) atinge — replica o
    // comportamento de "amplia a estratégia até achar confiança suficiente".
    @Test
    void convergeNaSegundaIteracao() {
        Clausula clausula = new Clausula("tema fraco", "texto forte", "f1");
        IndicePreparado indice = new IndicePreparado(
                List.of(clausula),
                Bm25.construirEstatisticasBM25(List.of(clausula.texto())),
                List.of(List.of(0.5, 0.5)),
                List.of(List.of(1.0, 0.0)));
        Map<String, IndicePreparado> preparados = Map.of("icf", indice);

        ResultadoBusca resultado = AgenticRag.buscarClausulaAgentica(
                preparados, "pergunta", List.of(1.0, 0.0), "icf", 0.9, null);

        assertThat(resultado.iteracoesUsadas()).isEqualTo(2);
        assertThat(resultado.esgotouLimite()).isFalse();
    }

    // Nenhuma das 3 estratégias atinge o limiar — Agentic RAG esgota o
    // limite e devolve o melhor resultado encontrado.
    @Test
    void esgotaLimiteQuandoNenhumaEstrategiaConverge() {
        Map<String, IndicePreparado> preparados = new LinkedHashMap<>();
        for (String nome : Indices.NOMES_INDICES) {
            Clausula clausula = new Clausula(nome + " tema", nome + " texto", nome + " fonte");
            preparados.put(nome, new IndicePreparado(
                    List.of(clausula),
                    Bm25.construirEstatisticasBM25(List.of(clausula.texto())),
                    List.of(List.of(0.0, 1.0)),
                    List.of(List.of(0.0, 1.0))));
        }

        ResultadoBusca resultado = AgenticRag.buscarClausulaAgentica(
                preparados, "pergunta", List.of(1.0, 0.0), "icf", 0.9, null);

        assertThat(resultado.iteracoesUsadas()).isEqualTo(AgenticRag.MAX_ITERACOES_AGENTIC);
        assertThat(resultado.esgotouLimite()).isTrue();
    }
}
