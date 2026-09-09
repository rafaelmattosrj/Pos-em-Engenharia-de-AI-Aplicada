package com.psprouting.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.psprouting.application.RoutingService;
import com.psprouting.application.SeedService;
import com.psprouting.domain.PSP;
import com.psprouting.domain.RoutingRecommendation;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.math.BigDecimal;
import java.util.List;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

/**
 * Testes de contrato dos endpoints descritos em IDEIA.md
 * ({@code POST /api/routing/recommend} e {@code POST /api/routing/seed}),
 * com {@link RoutingService}/{@link SeedService} mockados — sem depender de
 * Neo4j, do modelo de embeddings ou de uma chamada HTTP real ao OpenRouter.
 */
@WebMvcTest(RoutingController.class)
class RoutingControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private RoutingService routingService;

    @MockBean
    private SeedService seedService;

    @Test
    void recommend_comRequisicaoValidaRetorna200ComARecomendacao() throws Exception {
        RoutingRecommendation recommendation = new RoutingRecommendation(
                PSP.ADYEN, 0.87, "Alta aprovacao no Adyen para esse perfil.",
                List.of(PSP.BRASPAG, PSP.BRADESCO), List.of());
        when(routingService.recommend(any())).thenReturn(recommendation);

        String body = """
                {
                  "amount": 1500.00,
                  "method": "CREDIT_CARD",
                  "brand": "VISA",
                  "userRegion": "SP",
                  "hour": 20,
                  "merchantCategory": "STREAMING"
                }
                """;

        mockMvc.perform(post("/api/routing/recommend")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.primary").value("ADYEN"))
                .andExpect(jsonPath("$.confidence").value(0.87))
                .andExpect(jsonPath("$.fallback[0]").value("BRASPAG"));
    }

    @Test
    void recommend_semCamposObrigatoriosRetorna400() throws Exception {
        String body = """
                { "userRegion": "SP" }
                """;

        mockMvc.perform(post("/api/routing/recommend")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.error").value("VALIDATION_ERROR"));
    }

    @Test
    void recommend_horaForaDoIntervalo0a23Retorna400() throws Exception {
        String body = """
                {
                  "amount": 10.00,
                  "method": "PIX",
                  "userRegion": "SP",
                  "hour": 25,
                  "merchantCategory": "STREAMING"
                }
                """;

        mockMvc.perform(post("/api/routing/recommend")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(body))
                .andExpect(status().isBadRequest());
    }

    @Test
    void seed_retorna200ComQuantidadeDeTransacoesCarregadas() throws Exception {
        when(seedService.seed()).thenReturn(50);

        mockMvc.perform(post("/api/routing/seed"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.seeded").value(50));
    }
}
