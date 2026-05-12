package com.iadeva.mcp;

// Testes unitários do CustomerService — valida a lógica de filtragem em memória
// que é a principal funcionalidade adicionada pelo servidor MCP além da API legada

import com.iadeva.mcp.client.CustomerHttpClient;
import com.iadeva.mcp.model.Customer;
import com.iadeva.mcp.service.CustomerService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

/**
 * Testes unitários para {@link CustomerService}.
 * Foca na lógica de filtragem em memória do método searchCustomer,
 * que é a funcionalidade extra adicionada pelo servidor MCP.
 */
@ExtendWith(MockitoExtension.class)
class CustomerServiceTest {

    @Mock
    private CustomerHttpClient httpClient;

    @InjectMocks
    private CustomerService customerService;

    // Dados de teste reutilizados em todos os cenários
    private List<Customer> clientesFixture;

    @BeforeEach
    void setUp() {
        // Fixture com 3 clientes para cobrir cenários de filtragem
        clientesFixture = List.of(
                new Customer("1", "João Silva", "+55 11 91111-1111"),
                new Customer("2", "Maria Souza", "+55 21 92222-2222"),
                new Customer("3", "João Oliveira", "+55 11 93333-3333")
        );
    }

    @Test
    @DisplayName("searchCustomer filtra corretamente por nome parcial case-insensitive")
    void searchCustomer_filtrandoPorNome_deveRetornarApenasClientesComNomeCorrespondente() {
        // dado que a API retorna todos os clientes
        when(httpClient.listAll()).thenReturn(clientesFixture);

        // quando busco por "joão" (minúsculo, deve ser case-insensitive)
        List<Customer> resultado = customerService.searchCustomer(null, "joão", null);

        // então deve retornar apenas os dois Joões
        assertThat(resultado).hasSize(2);
        assertThat(resultado)
                .extracting(Customer::name)
                .containsExactlyInAnyOrder("João Silva", "João Oliveira");
    }

    @Test
    @DisplayName("searchCustomer filtra corretamente por telefone parcial")
    void searchCustomer_filtrandoPorTelefone_deveRetornarApenasClientesComTelefoneCorrespondente() {
        // dado que a API retorna todos os clientes
        when(httpClient.listAll()).thenReturn(clientesFixture);

        // quando busco pelo DDD 11 (dois clientes têm esse DDD)
        List<Customer> resultado = customerService.searchCustomer(null, null, "11");

        // então deve retornar apenas os clientes com DDD 11
        assertThat(resultado).hasSize(2);
        assertThat(resultado)
                .extracting(Customer::id)
                .containsExactlyInAnyOrder("1", "3");
    }

    @Test
    @DisplayName("searchCustomer sem filtros retorna todos os clientes")
    void searchCustomer_semFiltros_deveRetornarTodosOsClientes() {
        // dado que a API retorna todos os clientes
        when(httpClient.listAll()).thenReturn(clientesFixture);

        // quando busco sem nenhum critério (todos os parâmetros nulos)
        List<Customer> resultado = customerService.searchCustomer(null, null, null);

        // então deve retornar todos os 3 clientes sem filtrar
        assertThat(resultado).hasSize(3);
        assertThat(resultado).containsExactlyInAnyOrderElementsOf(clientesFixture);
    }
}
