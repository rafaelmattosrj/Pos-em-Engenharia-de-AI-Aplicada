package com.iadeva.musicas.service;

import com.iadeva.musicas.ResilientChatClient;
import com.iadeva.musicas.model.UserPreferences;
import com.iadeva.musicas.repository.PreferencesRepository;
import com.iadeva.musicas.support.MockOpenAiServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.springframework.ai.openai.OpenAiChatModel;
import org.springframework.ai.openai.OpenAiChatOptions;
import org.springframework.ai.openai.api.OpenAiApi;
import org.springframework.ai.retry.RetryUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

// Equivalente a TestPreferencesService_GetOrCreate / TestPreferencesService_ExtractAndSavePreferences (service_test.go)
@DataJpaTest
class PreferencesServiceTest {

    @Autowired
    private PreferencesRepository repository;

    private MockOpenAiServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.close();
        }
    }

    @Test
    void getOrCreate_createsOnFirstCall_returnsSameRecordOnSecondCall() {
        PreferencesService service = new PreferencesService(repository, null);

        UserPreferences first = service.getOrCreate("user-1");
        assertThat(first.getPreferences()).isEqualTo("{}");

        UserPreferences second = service.getOrCreate("user-1");
        assertThat(second.getUserId()).isEqualTo(first.getUserId());
        assertThat(repository.count()).isEqualTo(1); // segunda chamada não recria o registro
    }

    @Test
    void extractAndSavePreferences_persistsJsonExtractedByLlm() {
        server = MockOpenAiServer.fixedResponse(200, "{\"genero\":\"rock classico\"}");
        ResilientChatClient client = new ResilientChatClient(newChatModel(), "model-a");
        PreferencesService service = new PreferencesService(repository, client);

        service.extractAndSavePreferences("user-1", List.of("user: gosto de rock classico"));

        UserPreferences prefs = service.getOrCreate("user-1");
        assertThat(prefs.getPreferences()).isEqualTo("{\"genero\":\"rock classico\"}");
    }

    private OpenAiChatModel newChatModel() {
        OpenAiApi api = new OpenAiApi(server.baseUrl(), "test-key");
        return new OpenAiChatModel(api, OpenAiChatOptions.builder().model("placeholder").build(),
                null, RetryUtils.SHORT_RETRY_TEMPLATE);
    }
}
