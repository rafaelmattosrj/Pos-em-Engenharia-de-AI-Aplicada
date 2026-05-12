package com.iadeva.mcp.tool;

// Equivalente ao tools/customerTools.ts — define as ferramentas expostas pelo servidor MCP
// Cada método anotado com @Tool vira uma tool disponível para o LLM via protocolo MCP

import com.iadeva.mcp.model.Customer;
import com.iadeva.mcp.model.MutationResult;
import com.iadeva.mcp.service.CustomerService;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.ai.tool.annotation.ToolParam;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * Conjunto de tools MCP para gerenciamento de clientes.
 * Cada método público anotado com {@code @Tool} é registrado como uma ferramenta
 * invocável pelo LLM através do protocolo MCP.
 */
@Component
public class CustomerTools {

    private final CustomerService customerService;

    public CustomerTools(CustomerService customerService) {
        this.customerService = customerService;
    }

    /**
     * Tool: lista todos os clientes cadastrados na API legada.
     *
     * @return lista de todos os clientes
     */
    @Tool(description = "Lista todos os clientes cadastrados no sistema. Retorna id, nome e telefone de cada cliente.")
    public List<Customer> listCustomers() {
        return customerService.listAll();
    }

    /**
     * Tool: busca e filtra clientes por critérios parciais.
     * Todos os parâmetros são opcionais; combina os filtros fornecidos.
     *
     * @param id    filtra por ID exato do cliente
     * @param name  filtra por nome (busca parcial, case-insensitive)
     * @param phone filtra por telefone (busca parcial)
     * @return lista de clientes que satisfazem os critérios
     */
    @Tool(description = """
            Busca clientes por critérios flexíveis. Todos os parâmetros são opcionais.
            Use 'id' para busca exata por identificador.
            Use 'name' para busca parcial case-insensitive por nome.
            Use 'phone' para busca parcial por telefone.
            Se nenhum parâmetro for fornecido, retorna todos os clientes.
            """)
    public List<Customer> searchCustomer(
            @ToolParam(description = "ID exato do cliente (opcional)", required = false) String id,
            @ToolParam(description = "Nome ou parte do nome do cliente (opcional)", required = false) String name,
            @ToolParam(description = "Telefone ou parte do telefone do cliente (opcional)", required = false) String phone
    ) {
        return customerService.searchCustomer(id, name, phone);
    }

    /**
     * Tool: cria um novo cliente na API legada.
     *
     * @param name  nome completo do cliente
     * @param phone telefone de contato do cliente
     * @return resultado da operação indicando sucesso/falha e dados do cliente criado
     */
    @Tool(description = "Cria um novo cliente no sistema com nome e telefone. Retorna os dados do cliente criado incluindo o ID gerado.")
    public MutationResult createCustomer(
            @ToolParam(description = "Nome completo do cliente") String name,
            @ToolParam(description = "Telefone de contato do cliente") String phone
    ) {
        return customerService.createCustomer(name, phone);
    }

    /**
     * Tool: atualiza os dados de um cliente existente.
     *
     * @param id    identificador do cliente a ser atualizado
     * @param name  novo nome (opcional)
     * @param phone novo telefone (opcional)
     * @return resultado da operação com os dados atualizados
     */
    @Tool(description = """
            Atualiza os dados de um cliente existente identificado pelo ID.
            Forneça apenas os campos que deseja alterar (name e/ou phone).
            Pelo menos um dos campos name ou phone deve ser informado.
            """)
    public MutationResult updateCustomer(
            @ToolParam(description = "ID do cliente a ser atualizado") String id,
            @ToolParam(description = "Novo nome do cliente (opcional)", required = false) String name,
            @ToolParam(description = "Novo telefone do cliente (opcional)", required = false) String phone
    ) {
        return customerService.updateCustomer(id, name, phone);
    }

    /**
     * Tool: remove um cliente pelo ID.
     *
     * @param id identificador do cliente a ser removido
     * @return resultado da operação indicando sucesso/falha
     */
    @Tool(description = "Remove permanentemente um cliente do sistema pelo ID. Esta operação não pode ser desfeita.")
    public MutationResult deleteCustomer(
            @ToolParam(description = "ID do cliente a ser removido") String id
    ) {
        return customerService.deleteCustomer(id);
    }
}
