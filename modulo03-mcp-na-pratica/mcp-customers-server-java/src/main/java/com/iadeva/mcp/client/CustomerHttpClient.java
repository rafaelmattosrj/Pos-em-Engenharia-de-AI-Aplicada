package com.iadeva.mcp.client;

// Cliente HTTP para a legacy customer API — abstrai todas as chamadas REST para a API legada
// Equivalente ao httpClient.ts / fetch calls nos projetos TypeScript do curso

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.mcp.model.Customer;
import com.iadeva.mcp.model.MutationResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Componente responsável por toda comunicação HTTP com a legacy customer API.
 * Autentica via Bearer token lido da variável de ambiente API_SERVICE_TOKEN.
 */
@Component
public class CustomerHttpClient {

    private static final Logger logger = LoggerFactory.getLogger(CustomerHttpClient.class);

    private final RestClient restClient;
    private final ObjectMapper objectMapper;

    public CustomerHttpClient(
            @Value("${app.legacy-api.base-url:http://localhost:3000}") String baseUrl,
            @Value("${app.legacy-api.service-token:}") String serviceToken,
            ObjectMapper objectMapper
    ) {
        this.objectMapper = objectMapper;
        // Configura RestClient com base URL e header de autenticação padrão
        this.restClient = RestClient.builder()
                .baseUrl(baseUrl)
                .defaultHeader("Authorization", "Bearer " + serviceToken)
                .defaultHeader("Content-Type", "application/json")
                .build();

        logger.info("CustomerHttpClient inicializado com base URL: {}", baseUrl);
    }

    /**
     * Lista todos os clientes cadastrados na API legada.
     *
     * @return lista de clientes
     */
    public List<Customer> listAll() {
        try {
            String responseBody = restClient.get()
                    .uri("/customers")
                    .retrieve()
                    .body(String.class);

            return parseCustomerList(responseBody);
        } catch (Exception e) {
            logger.error("Erro ao listar clientes: {}", e.getMessage());
            throw new RuntimeException("Falha ao listar clientes da API legada", e);
        }
    }

    /**
     * Busca um cliente específico pelo ID.
     *
     * @param id identificador do cliente
     * @return cliente encontrado ou null se não existir
     */
    public Customer findById(String id) {
        try {
            String responseBody = restClient.get()
                    .uri("/customers/{id}", id)
                    .retrieve()
                    .body(String.class);

            return parseCustomer(responseBody);
        } catch (Exception e) {
            logger.error("Erro ao buscar cliente por ID {}: {}", id, e.getMessage());
            return null;
        }
    }

    /**
     * Cria um novo cliente na API legada.
     *
     * @param name  nome do cliente
     * @param phone telefone do cliente
     * @return resultado da operação de criação
     */
    public MutationResult create(String name, String phone) {
        try {
            Map<String, String> payload = Map.of("name", name, "phone", phone);
            String body = objectMapper.writeValueAsString(payload);

            String responseBody = restClient.post()
                    .uri("/customers")
                    .body(body)
                    .retrieve()
                    .body(String.class);

            Customer created = parseCustomer(responseBody);
            return new MutationResult(true, "Cliente criado com sucesso", created);
        } catch (Exception e) {
            logger.error("Erro ao criar cliente: {}", e.getMessage());
            return new MutationResult(false, "Falha ao criar cliente: " + e.getMessage(), null);
        }
    }

    /**
     * Atualiza os dados de um cliente existente.
     *
     * @param id    identificador do cliente
     * @param name  novo nome (pode ser null para não alterar)
     * @param phone novo telefone (pode ser null para não alterar)
     * @return resultado da operação de atualização
     */
    public MutationResult update(String id, String name, String phone) {
        try {
            // Monta payload apenas com campos fornecidos
            Map<String, String> payload;
            if (name != null && phone != null) {
                payload = Map.of("name", name, "phone", phone);
            } else if (name != null) {
                payload = Map.of("name", name);
            } else if (phone != null) {
                payload = Map.of("phone", phone);
            } else {
                return new MutationResult(false, "Nenhum campo fornecido para atualização", null);
            }

            String body = objectMapper.writeValueAsString(payload);

            String responseBody = restClient.put()
                    .uri("/customers/{id}", id)
                    .body(body)
                    .retrieve()
                    .body(String.class);

            Customer updated = parseCustomer(responseBody);
            return new MutationResult(true, "Cliente atualizado com sucesso", updated);
        } catch (Exception e) {
            logger.error("Erro ao atualizar cliente {}: {}", id, e.getMessage());
            return new MutationResult(false, "Falha ao atualizar cliente: " + e.getMessage(), null);
        }
    }

    /**
     * Remove um cliente da API legada pelo ID.
     *
     * @param id identificador do cliente a ser removido
     * @return resultado da operação de exclusão
     */
    public MutationResult delete(String id) {
        try {
            restClient.delete()
                    .uri("/customers/{id}", id)
                    .retrieve()
                    .toBodilessEntity();

            return new MutationResult(true, "Cliente removido com sucesso", null);
        } catch (Exception e) {
            logger.error("Erro ao remover cliente {}: {}", id, e.getMessage());
            return new MutationResult(false, "Falha ao remover cliente: " + e.getMessage(), null);
        }
    }

    // --- Métodos auxiliares de parsing ---

    private Customer parseCustomer(String json) {
        try {
            JsonNode node = objectMapper.readTree(json);
            return new Customer(
                    node.path("id").asText(null),
                    node.path("name").asText(null),
                    node.path("phone").asText(null)
            );
        } catch (Exception e) {
            throw new RuntimeException("Erro ao parsear resposta de cliente: " + json, e);
        }
    }

    private List<Customer> parseCustomerList(String json) {
        try {
            JsonNode array = objectMapper.readTree(json);
            List<Customer> customers = new ArrayList<>();
            if (array.isArray()) {
                for (JsonNode node : array) {
                    customers.add(new Customer(
                            node.path("id").asText(null),
                            node.path("name").asText(null),
                            node.path("phone").asText(null)
                    ));
                }
            }
            return customers;
        } catch (Exception e) {
            throw new RuntimeException("Erro ao parsear lista de clientes: " + json, e);
        }
    }
}
