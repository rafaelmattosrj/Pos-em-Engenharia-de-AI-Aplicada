package com.iadeva.gateway;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebMvcTest(ChatController.class)
class ChatControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private OpenRouterService openRouterService;

    @Test
    void postChat_retornaModeloEConteudoDoOpenRouterService() throws Exception {
        when(openRouterService.generate(anyString())).thenReturn(new LlmResponse("model-a", "resposta"));

        mockMvc.perform(post("/chat")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"question\":\"O que e machine learning?\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.model").value("model-a"))
                .andExpect(jsonPath("$.content").value("resposta"));
    }

    @Test
    void postChat_retorna400QuandoQuestionMenorQue5Caracteres() throws Exception {
        mockMvc.perform(post("/chat")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"question\":\"oi\"}"))
                .andExpect(status().isBadRequest());
    }

    @Test
    void getChat_retorna405MetodoNaoPermitido() throws Exception {
        mockMvc.perform(get("/chat"))
                .andExpect(status().isMethodNotAllowed());
    }
}
