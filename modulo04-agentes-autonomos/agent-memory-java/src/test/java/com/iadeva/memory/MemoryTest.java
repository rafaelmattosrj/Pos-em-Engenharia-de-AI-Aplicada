package com.iadeva.memory;

import com.iadeva.memory.memory.ContextualMemory;
import com.iadeva.memory.memory.EpisodicMemory;
import com.iadeva.memory.memory.LongTermMemory;
import com.iadeva.memory.model.Episode;
import com.iadeva.memory.model.MemoryFact;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.Mockito;
import org.springframework.ai.embedding.EmbeddingModel;

import java.io.File;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

/**
 * Testes unitários para os 3 tipos de memória persistida.
 * Usa diretório temporário para isolar cada teste — sem efeitos colaterais.
 */
class MemoryTest {

    @TempDir
    File tempDir;

    private LongTermMemory longTermMemory;
    private EpisodicMemory episodicMemory;
    private ContextualMemory contextualMemory;
    private EmbeddingModel embeddingModelMock;

    @BeforeEach
    void setUp() {
        String path = tempDir.getAbsolutePath();

        longTermMemory = new LongTermMemory(path);
        longTermMemory.init();

        episodicMemory = new EpisodicMemory(path);
        episodicMemory.init();

        // Mock do EmbeddingModel — evita chamadas reais à API da OpenAI nos testes
        embeddingModelMock = Mockito.mock(EmbeddingModel.class);
    }

    // -------------------------------------------------------------------------
    // CENÁRIO 1: LongTermMemory não deve duplicar fatos com mesmo content
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("LongTermMemory: não deve duplicar fatos com mesmo content")
    void longTermMemory_naoDeveDuplicar_fatosComMesmoContent() {
        MemoryFact fato = new MemoryFact(
                UUID.randomUUID().toString(),
                "O usuário prefere respostas em português",
                "usuario",
                Instant.now(),
                null
        );

        // Adiciona o mesmo fato duas vezes (mesmo content, id diferente)
        longTermMemory.addFact(fato);
        longTermMemory.addFact(new MemoryFact(
                UUID.randomUUID().toString(), // id diferente
                "O usuário prefere respostas em português", // mesmo content
                "sistema",
                Instant.now(),
                null
        ));

        List<MemoryFact> fatos = longTermMemory.getFacts();
        assertThat(fatos).hasSize(1);
        assertThat(fatos.get(0).content()).isEqualTo("O usuário prefere respostas em português");
    }

    // -------------------------------------------------------------------------
    // CENÁRIO 2: ContextualMemory retorna resultados acima do threshold e ignora abaixo
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("ContextualMemory: retorna resultados acima do threshold e ignora abaixo")
    void contextualMemory_filtraPorThreshold() {
        // Vetores controlados: alta similaridade (cosseno ~1.0) e baixa (~0.0)
        float[] vetorBase = {1.0f, 0.0f, 0.0f};
        float[] vetorSimilar = {0.95f, 0.1f, 0.0f}; // cosseno alto
        float[] vetorDistante = {0.0f, 0.0f, 1.0f};  // cosseno ~ 0.0

        // Mock: primeira chamada = query, segunda = vetorSimilar, terceira = vetorDistante
        when(embeddingModelMock.embed(anyString()))
                .thenReturn(toDoubleList(vetorBase))    // embed da query
                .thenReturn(toDoubleList(vetorSimilar)) // store conteúdo 1
                .thenReturn(toDoubleList(vetorDistante)); // store conteúdo 2

        // Threshold 0.7 — só vetorSimilar deve passar
        ContextualMemory memory = new ContextualMemory(embeddingModelMock, 0.7);

        // Simula stores já realizados (precisamos que o mock forneça embeddings ao store)
        // Estratégia: resetar mock e configurar para store → search
        Mockito.reset(embeddingModelMock);
        when(embeddingModelMock.embed("conteúdo relevante")).thenReturn(toDoubleList(vetorSimilar));
        when(embeddingModelMock.embed("conteúdo irrelevante")).thenReturn(toDoubleList(vetorDistante));
        when(embeddingModelMock.embed("query de busca")).thenReturn(toDoubleList(vetorBase));

        memory.store("conteúdo relevante", Map.of("tipo", "teste"));
        memory.store("conteúdo irrelevante", Map.of("tipo", "teste"));

        List<ContextualMemory.EmbeddingEntry> resultados = memory.search("query de busca", 10);

        // Apenas o conteúdo com alta similaridade deve retornar
        assertThat(resultados).hasSize(1);
        assertThat(resultados.get(0).content()).isEqualTo("conteúdo relevante");
    }

    // -------------------------------------------------------------------------
    // CENÁRIO 3: EpisodicMemory getRecentEpisodes respeita limit
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("EpisodicMemory: getRecentEpisodes deve respeitar o limit informado")
    void episodicMemory_getRecentEpisodes_respeitaLimit() {
        // Adiciona 5 episódios com timestamps distintos
        for (int i = 1; i <= 5; i++) {
            episodicMemory.addEpisode(new Episode(
                    UUID.randomUUID().toString(),
                    "input " + i,
                    List.of("passo 1", "passo 2"),
                    "resultado " + i,
                    List.of(),
                    Instant.now().plusSeconds(i) // timestamps crescentes
            ));
        }

        List<Episode> recentes = episodicMemory.getRecentEpisodes(3);

        assertThat(recentes).hasSize(3);
        // Verifica que os mais recentes vêm primeiro (maior timestamp)
        assertThat(recentes.get(0).input()).isEqualTo("input 5");
        assertThat(recentes.get(1).input()).isEqualTo("input 4");
        assertThat(recentes.get(2).input()).isEqualTo("input 3");
    }

    // Utilitário: converte float[] para List<Double> (formato retornado pelo EmbeddingModel)
    private List<Double> toDoubleList(float[] arr) {
        List<Double> list = new java.util.ArrayList<>();
        for (float v : arr) list.add((double) v);
        return list;
    }
}
