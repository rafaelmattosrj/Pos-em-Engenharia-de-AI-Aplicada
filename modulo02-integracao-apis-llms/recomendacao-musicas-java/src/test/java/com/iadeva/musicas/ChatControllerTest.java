package com.iadeva.musicas;

import com.iadeva.musicas.support.MockOpenAiServer;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.test.web.servlet.MockMvc;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

/**
 * Teste de integração ponta a ponta do endpoint POST /chat, subindo o contexto Spring real com
 * um banco H2 em memória isolado (spring.datasource.url sobrescrito) e um mock server HTTP local
 * (JDK puro) no lugar da API OpenRouter real. Equivalente a handler_test.go do porte Go.
 */
@SpringBootTest
@AutoConfigureMockMvc
class ChatControllerTest {

    private static MockOpenAiServer server;

    @Autowired
    private MockMvc mockMvc;

    @BeforeAll
    static void startMockServer() {
        server = MockOpenAiServer.fixedResponse(200, "resposta do assistente");
    }

    @AfterAll
    static void stopMockServer() {
        server.close();
    }

    @DynamicPropertySource
    static void overrideProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.ai.openai.base-url", () -> server.baseUrl());
        registry.add("spring.ai.openai.api-key", () -> "test-key");
        registry.add("app.fallback-models", () -> "model-a");
        // Banco em memória isolado para o teste, em vez do H2 file-based usado em runtime.
        registry.add("spring.datasource.url", () -> "jdbc:h2:mem:musicdb-test;DB_CLOSE_DELAY=-1");
    }

    @Test
    void chat_success_returns200WithReply() throws Exception {
        mockMvc.perform(post("/chat")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"message\":\"me recomenda algo\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.reply").value("resposta do assistente"));
    }

    @Test
    void chat_messageTooShort_returns400() throws Exception {
        mockMvc.perform(post("/chat")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"message\":\"oi\"}"))
                .andExpect(status().isBadRequest());
    }
}
