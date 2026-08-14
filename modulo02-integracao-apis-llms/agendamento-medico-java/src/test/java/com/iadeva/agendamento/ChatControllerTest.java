package com.iadeva.agendamento;

import com.iadeva.agendamento.graph.AppointmentOrchestrator;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebMvcTest(ChatController.class)
class ChatControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private AppointmentOrchestrator orchestrator;

    @Test
    void postChat_retornaReplyDoOrchestrator() throws Exception {
        when(orchestrator.process(anyString())).thenReturn("Ola! Como posso ajudar?");

        mockMvc.perform(post("/chat")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"question\":\"oi tudo bem?\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.reply").value("Ola! Como posso ajudar?"));
    }

    @Test
    void postChat_retorna400QuandoQuestionMenorQue5Caracteres() throws Exception {
        mockMvc.perform(post("/chat")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"question\":\"oi\"}"))
                .andExpect(status().isBadRequest());
    }
}
