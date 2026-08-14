package com.iadeva.cfp;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.hamcrest.Matchers.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Testes de integração — equivalente a api-e2e/src/api/api.spec.ts, mais
 * cobertura adicional para os endpoints de events/speakers.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
class CfpPlatformApiTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    // Nota: server.servlet.context-path=/api já é aplicado automaticamente
    // pelo MockMvc (via @AutoConfigureMockMvc), então os paths aqui NÃO
    // repetem o prefixo "/api" — equivalente a testar contra
    // http://localhost:PORT/api/... na versão NestJS (api-e2e), onde o
    // cliente axios já aponta para a baseURL com o prefixo incluído.

    @Test
    void getApi_retornaMensagemDeSaudacao() throws Exception {
        mockMvc.perform(get("/"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.message", is("Hello API")));
    }

    @Test
    void postEvents_criaEventoComIdGerado() throws Exception {
        String body = """
                {"nome": "DevFest", "endereco": "Centro de Convencoes", "capacidade": 500, "data": "2026-08-15"}
                """;

        mockMvc.perform(post("/events")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.id", not(emptyOrNullString())))
                .andExpect(jsonPath("$.nome", is("DevFest")))
                .andExpect(jsonPath("$.capacidade", is(500)));
    }

    @Test
    void postEvents_semNomeRetorna400() throws Exception {
        String body = """
                {"nome": "", "endereco": "Centro", "capacidade": 100, "data": "2026-08-15"}
                """;

        mockMvc.perform(post("/events")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isBadRequest());
    }

    @Test
    void getEvents_listaEventosCriados() throws Exception {
        String body = """
                {"nome": "TDC", "endereco": "Anhembi", "capacidade": 2000, "data": "2026-10-01"}
                """;
        mockMvc.perform(post("/events").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk());

        mockMvc.perform(get("/events"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$", hasSize(greaterThanOrEqualTo(1))));
    }

    @Test
    void postSpeakers_criaPalestranteComIdGerado() throws Exception {
        String body = """
                {"name": "Ana Souza", "email": "ana@example.com", "talkTitle": "IA na pratica", "isGDE": true}
                """;

        mockMvc.perform(post("/speakers")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.id", not(emptyOrNullString())))
                .andExpect(jsonPath("$.name", is("Ana Souza")))
                .andExpect(jsonPath("$.isGDE", is(true)));
    }

    @Test
    void postSpeakers_emailInvalidoRetorna400() throws Exception {
        String body = """
                {"name": "Bruno", "email": "nao-e-um-email", "talkTitle": "Talk", "isGDE": false}
                """;

        mockMvc.perform(post("/speakers")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isBadRequest());
    }
}
