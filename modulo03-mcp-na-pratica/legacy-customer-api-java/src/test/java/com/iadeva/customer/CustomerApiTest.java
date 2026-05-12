package com.iadeva.customer;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.customer.model.AuthRequest;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;
import static org.hamcrest.Matchers.*;

/**
 * Testes de integração da API de clientes.
 * Equivalente aos testes de integração do módulo 07 do curso JS/TS.
 *
 * Cenários cobertos:
 * 1. Login com credenciais válidas → retorna JWT
 * 2. Acesso sem token → 401 Unauthorized
 * 3. MEMBER tentando POST /customers → 403 Forbidden
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
class CustomerApiTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    // -------------------------------------------------------------------------
    // Cenário 1: Login com credenciais válidas retorna JWT
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("POST /auth/login com credenciais válidas retorna 200 e token JWT")
    void loginComCredenciaisValidas_retornaJwt() throws Exception {
        AuthRequest request = new AuthRequest("admin", "password123");

        mockMvc.perform(post("/auth/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.token", not(emptyOrNullString())))
                .andExpect(jsonPath("$.role", is("ADMIN")));
    }

    @Test
    @DisplayName("POST /auth/login com credenciais inválidas retorna 401")
    void loginComCredenciaisInvalidas_retorna401() throws Exception {
        AuthRequest request = new AuthRequest("admin", "senhaErrada");

        mockMvc.perform(post("/auth/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isUnauthorized());
    }

    // -------------------------------------------------------------------------
    // Cenário 2: Acesso sem token retorna 401
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("GET /customers sem token retorna 401 Unauthorized")
    void acessoSemToken_retorna401() throws Exception {
        mockMvc.perform(get("/customers"))
                .andExpect(status().isUnauthorized());
    }

    // -------------------------------------------------------------------------
    // Cenário 3: MEMBER não pode fazer POST /customers (403 Forbidden)
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("MEMBER tentando POST /customers retorna 403 Forbidden")
    void memberNaoPodeCriarCliente_retorna403() throws Exception {
        // Passo 1: faz login como member e obtém JWT
        AuthRequest loginRequest = new AuthRequest("member", "pass456");

        MvcResult loginResult = mockMvc.perform(post("/auth/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(loginRequest)))
                .andExpect(status().isOk())
                .andReturn();

        String responseBody = loginResult.getResponse().getContentAsString();
        String token = objectMapper.readTree(responseBody).get("token").asText();

        // Passo 2: tenta criar um cliente com o token de MEMBER
        String newCustomer = """
                {
                    "name": "Teste MEMBER",
                    "phone": "(00) 00000-0000"
                }
                """;

        mockMvc.perform(post("/customers")
                        .header("Authorization", "Bearer " + token)
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(newCustomer))
                .andExpect(status().isForbidden());
    }

    // -------------------------------------------------------------------------
    // Bônus: ADMIN pode listar clientes
    // -------------------------------------------------------------------------

    @Test
    @DisplayName("ADMIN com token válido pode listar clientes")
    void adminComToken_podeListarClientes() throws Exception {
        // Login como admin
        AuthRequest loginRequest = new AuthRequest("admin", "password123");

        MvcResult loginResult = mockMvc.perform(post("/auth/login")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(loginRequest)))
                .andExpect(status().isOk())
                .andReturn();

        String token = objectMapper.readTree(
                loginResult.getResponse().getContentAsString()
        ).get("token").asText();

        // Lista clientes com token válido
        mockMvc.perform(get("/customers")
                        .header("Authorization", "Bearer " + token))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$", hasSize(greaterThanOrEqualTo(5))));
    }
}
