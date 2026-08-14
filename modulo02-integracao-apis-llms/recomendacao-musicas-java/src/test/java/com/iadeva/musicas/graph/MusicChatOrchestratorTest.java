package com.iadeva.musicas.graph;

import com.iadeva.musicas.ResilientChatClient;
import com.iadeva.musicas.model.UserPreferences;
import com.iadeva.musicas.repository.ConversationRepository;
import com.iadeva.musicas.repository.PreferencesRepository;
import com.iadeva.musicas.service.MemoryService;
import com.iadeva.musicas.service.PreferencesService;
import com.iadeva.musicas.support.MockOpenAiServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.springframework.ai.openai.OpenAiChatModel;
import org.springframework.ai.openai.OpenAiChatOptions;
import org.springframework.ai.openai.api.OpenAiApi;
import org.springframework.ai.retry.RetryUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.test.util.ReflectionTestUtils;

import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;

// Equivalente a graph/orchestrator_test.go — resposta básica + histórico, extração por
// palavra-chave e sumarização por threshold.
@DataJpaTest
class MusicChatOrchestratorTest {

    @Autowired
    private ConversationRepository conversationRepository;

    @Autowired
    private PreferencesRepository preferencesRepository;

    private MockOpenAiServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.close();
        }
    }

    @Test
    void chat_basicReply_persistsHistory() {
        server = MockOpenAiServer.fixedResponse(200, "Recomendo Pink Floyd!");
        MusicChatOrchestrator orchestrator = newOrchestrator(10);

        String reply = orchestrator.chat("user-1", "session-1", "me indique algo de rock progressivo");

        assertThat(reply).isEqualTo("Recomendo Pink Floyd!");
        assertThat(conversationRepository.findBySessionIdOrderByCreatedAtAsc("session-1")).hasSize(2);
    }

    @Test
    void chat_musicPreferenceKeyword_triggersExtraction() {
        AtomicInteger calls = new AtomicInteger();
        server = MockOpenAiServer.start(body -> {
            calls.incrementAndGet();
            return new MockOpenAiServer.Response(200, "Legal saber que voce gosta disso!");
        });
        MusicChatOrchestrator orchestrator = newOrchestrator(10);

        orchestrator.chat("user-1", "session-1", "eu gosto muito de jazz");

        // 1 chamada para a resposta do chat + 1 chamada para extrair preferencias.
        assertThat(calls.get()).isEqualTo(2);
        UserPreferences prefs = preferencesRepository.findByUserId("user-1").orElseThrow();
        assertThat(prefs.getPreferences()).isNotEqualTo("{}");
    }

    @Test
    void chat_summarizesAfterThreshold() {
        server = MockOpenAiServer.fixedResponse(200, "resposta padrao");
        MusicChatOrchestrator orchestrator = newOrchestrator(2);

        orchestrator.chat("user-1", "session-1", "primeira mensagem neutra");
        // Apos a 1a troca ja ha 2 mensagens (user+assistant) >= threshold=2,
        // entao a sumarizacao deve disparar ja nessa chamada.
        orchestrator.chat("user-1", "session-1", "segunda mensagem neutra");

        UserPreferences prefs = preferencesRepository.findByUserId("user-1").orElseThrow();
        assertThat(prefs.getConversationSummary()).isNotBlank();
    }

    private MusicChatOrchestrator newOrchestrator(int summarizeAfter) {
        OpenAiApi api = new OpenAiApi(server.baseUrl(), "test-key");
        OpenAiChatModel chatModel = new OpenAiChatModel(api, OpenAiChatOptions.builder().model("placeholder").build(),
                null, RetryUtils.SHORT_RETRY_TEMPLATE);
        ResilientChatClient client = new ResilientChatClient(chatModel, "model-a");

        MemoryService memoryService = new MemoryService(conversationRepository);
        PreferencesService preferencesService = new PreferencesService(preferencesRepository, client);

        MusicChatOrchestrator orchestrator = new MusicChatOrchestrator(client, memoryService, preferencesService);
        ReflectionTestUtils.setField(orchestrator, "summarizeAfterMessages", summarizeAfter);
        return orchestrator;
    }
}
