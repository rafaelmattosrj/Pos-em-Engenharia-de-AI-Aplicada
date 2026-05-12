package com.iadeva.mcp.service;

// Equivalente ao customerService.ts — camada de negócio entre MCP e HTTP client
// Adiciona funcionalidade de busca/filtro em memória que a API legada não oferece nativamente

import com.iadeva.mcp.client.CustomerHttpClient;
import com.iadeva.mcp.model.Customer;
import com.iadeva.mcp.model.MutationResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * Serviço de domínio para operações com clientes.
 * Delega chamadas HTTP ao {@link CustomerHttpClient} e adiciona lógica de negócio
 * como filtragem em memória por múltiplos critérios.
 */
@Service
public class CustomerService {

    private static final Logger logger = LoggerFactory.getLogger(CustomerService.class);

    private final CustomerHttpClient httpClient;

    public CustomerService(CustomerHttpClient httpClient) {
        this.httpClient = httpClient;
    }

    /**
     * Retorna todos os clientes cadastrados.
     *
     * @return lista completa de clientes
     */
    public List<Customer> listAll() {
        logger.debug("Listando todos os clientes");
        return httpClient.listAll();
    }

    /**
     * Busca e filtra clientes em memória por id, nome ou telefone.
     * Funcionalidade extra que a API legada não oferece — equivalente ao searchCustomer do TypeScript.
     * Todos os parâmetros são opcionais; se nenhum for fornecido retorna todos os clientes.
     * O filtro usa correspondência parcial case-insensitive para name e phone.
     *
     * @param id    filtra por ID exato (opcional)
     * @param name  filtra por nome contendo o valor (opcional, case-insensitive)
     * @param phone filtra por telefone contendo o valor (opcional)
     * @return lista de clientes que satisfazem todos os critérios fornecidos
     */
    public List<Customer> searchCustomer(String id, String name, String phone) {
        logger.debug("Buscando clientes — id={}, name={}, phone={}", id, name, phone);

        List<Customer> all = httpClient.listAll();

        return all.stream()
                .filter(c -> id == null || id.isBlank() || id.equals(c.id()))
                .filter(c -> name == null || name.isBlank()
                        || (c.name() != null && c.name().toLowerCase().contains(name.toLowerCase())))
                .filter(c -> phone == null || phone.isBlank()
                        || (c.phone() != null && c.phone().contains(phone)))
                .toList();
    }

    /**
     * Cria um novo cliente.
     *
     * @param name  nome do cliente
     * @param phone telefone do cliente
     * @return resultado da operação
     */
    public MutationResult createCustomer(String name, String phone) {
        logger.debug("Criando cliente — name={}", name);
        return httpClient.create(name, phone);
    }

    /**
     * Atualiza um cliente existente.
     *
     * @param id    identificador do cliente
     * @param name  novo nome (opcional)
     * @param phone novo telefone (opcional)
     * @return resultado da operação
     */
    public MutationResult updateCustomer(String id, String name, String phone) {
        logger.debug("Atualizando cliente — id={}", id);
        return httpClient.update(id, name, phone);
    }

    /**
     * Remove um cliente pelo ID.
     *
     * @param id identificador do cliente
     * @return resultado da operação
     */
    public MutationResult deleteCustomer(String id) {
        logger.debug("Removendo cliente — id={}", id);
        return httpClient.delete(id);
    }
}
