package com.trialforge.gateway;

import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class RagSearchTest {

    @Test
    void buscarClausulaHibridaEscolheMelhorPorFusao() {
        List<Clausula> clausulas = List.of(
                new Clausula("tema um", "texto sobre gatos e felinos", "fonte1"),
                new Clausula("tema dois", "texto sobre carros e motos", "fonte2"));
        IndicePreparado indice = new IndicePreparado(
                clausulas,
                Bm25.construirEstatisticasBM25(clausulas.stream().map(Clausula::texto).toList()),
                List.of(List.of(1.0, 0.0), List.of(0.0, 1.0)),
                List.of(List.of(1.0, 0.0), List.of(0.0, 1.0)));
        Map<String, IndicePreparado> preparados = Map.of("teste", indice);

        ResultadoBusca resultado = RagSearch.buscarClausulaHibrida(preparados, "gatos", List.of(1.0, 0.0), "teste", "tema");

        assertThat(resultado.clausula().tema()).isEqualTo("tema um");
        assertThat(resultado.similaridadeCosseno()).isEqualTo(1.0);
        assertThat(resultado.indice()).isEqualTo("teste");
    }

    @Test
    void buscarEmTodosIndicesEscolheOMaiorCosseno() {
        Map<String, IndicePreparado> preparados = new LinkedHashMap<>();
        preparados.put("icf", indiceDeUmaClausula("icf", List.of(0.0, 1.0)));
        preparados.put("protocolo", indiceDeUmaClausula("protocolo", List.of(1.0, 0.0)));
        preparados.put("csr", indiceDeUmaClausula("csr", List.of(0.0, -1.0)));

        ResultadoBusca resultado = RagSearch.buscarEmTodosIndices(preparados, "pergunta qualquer", List.of(1.0, 0.0));

        assertThat(resultado.indice()).isEqualTo("protocolo");
    }

    private static IndicePreparado indiceDeUmaClausula(String nome, List<Double> vetor) {
        Clausula clausula = new Clausula(nome + " tema", nome + " texto", nome + " fonte");
        return new IndicePreparado(
                List.of(clausula),
                Bm25.construirEstatisticasBM25(List.of(clausula.texto())),
                List.of(vetor),
                List.of(vetor));
    }

    @Test
    void prepararIndicesChamaEmbedderParaCadaTemaETexto() throws Exception {
        FakeEmbedder fake = new FakeEmbedder();

        Map<String, IndicePreparado> preparados = RagSearch.prepararIndices(fake, null);

        assertThat(preparados).hasSize(Indices.NOMES_INDICES.size());
        for (String nome : Indices.NOMES_INDICES) {
            IndicePreparado indice = preparados.get(nome);
            assertThat(indice.embeddingsTema()).hasSameSizeAs(indice.clausulas());
            assertThat(indice.embeddingsTexto()).hasSameSizeAs(indice.clausulas());
        }
        // 2 cláusulas por índice x 3 índices x 2 embeddings (tema + texto) = 12 chamadas
        assertThat(fake.chamadas).isEqualTo(12);
    }
}
