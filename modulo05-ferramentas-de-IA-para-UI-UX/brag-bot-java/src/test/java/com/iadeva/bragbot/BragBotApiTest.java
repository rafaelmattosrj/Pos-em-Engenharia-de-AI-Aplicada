package com.iadeva.bragbot;

import com.iadeva.bragbot.gemini.GeminiClient;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.anyDouble;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;
import static org.hamcrest.Matchers.*;

/**
 * Testes de integração da rota POST /api/brag — o GeminiClient é mockado
 * para evitar chamadas reais à API do Gemini.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
class BragBotApiTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private GeminiClient geminiClient;

    @Test
    void postBrag_semDefinitionRetorna400() throws Exception {
        mockMvc.perform(post("/api/brag")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.error", is("Definition is required")));
    }

    @Test
    void postBrag_comDefinitionValidaRetornaBragDocument() throws Exception {
        String geminiJson = """
                {
                  "title": "Reduziu latencia da API em 50%",
                  "context": "API de pagamentos com timeouts frequentes",
                  "actionTaken": "Refatorou o connection pool e adicionou cache",
                  "businessImpact": "Reducao de 80% nos tickets de suporte",
                  "metrics": ["50% reduction", "10ms latency"],
                  "technologiesUsed": ["Java", "Redis"]
                }
                """;
        when(geminiClient.generateJson(anyString(), anyDouble())).thenReturn(geminiJson);

        mockMvc.perform(post("/api/brag")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"definition\": \"otimizei a api de pagamentos\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.id", not(emptyOrNullString())))
                .andExpect(jsonPath("$.title", is("Reduziu latencia da API em 50%")))
                .andExpect(jsonPath("$.metrics", hasSize(2)))
                .andExpect(jsonPath("$.technologiesUsed", hasSize(2)));
    }

    @Test
    void postBrag_quandoGeminiFalhaRetorna500() throws Exception {
        when(geminiClient.generateJson(anyString(), anyDouble()))
                .thenThrow(new RuntimeException("falha na API do Gemini"));

        mockMvc.perform(post("/api/brag")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"definition\": \"algo qualquer\"}"))
                .andExpect(status().isInternalServerError())
                .andExpect(jsonPath("$.error", is("Failed to generate brag")));
    }

    @Test
    void postBrag_quandoRespostaNaoEJsonValidoRetorna500() throws Exception {
        when(geminiClient.generateJson(anyString(), anyDouble())).thenReturn("isso nao e JSON");

        mockMvc.perform(post("/api/brag")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"definition\": \"algo qualquer\"}"))
                .andExpect(status().isInternalServerError());
    }
}
