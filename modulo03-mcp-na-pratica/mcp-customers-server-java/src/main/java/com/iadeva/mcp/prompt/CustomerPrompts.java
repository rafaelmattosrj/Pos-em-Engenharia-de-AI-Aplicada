package com.iadeva.mcp.prompt;

// Equivalente ao prompts/ do curso MCP — templates reutilizáveis para o LLM
// Prompts MCP são templates pré-definidos que guiam o LLM em tarefas recorrentes

import io.modelcontextprotocol.server.McpServerFeatures;
import io.modelcontextprotocol.spec.McpSchema;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.List;
import java.util.Map;

/**
 * Configuração dos prompts MCP expostos pelo servidor.
 * Prompts são templates que o cliente MCP pode invocar para obter mensagens
 * pré-formatadas, facilitando interações padronizadas com o LLM.
 */
@Configuration
public class CustomerPrompts {

    /**
     * Prompt: search-customer-prompt
     * Gera uma mensagem estruturada para que o LLM busque clientes de forma eficiente.
     * Recebe a query do usuário e produz um prompt orientando o uso da tool searchCustomer.
     */
    @Bean
    public McpServerFeatures.SyncPromptRegistration searchCustomerPrompt() {
        // Definição do prompt com nome, descrição e parâmetros aceitos
        var promptDefinition = new McpSchema.Prompt(
                "search-customer-prompt",
                "Template para busca de clientes. Use quando o usuário quiser encontrar um ou mais clientes por nome, telefone ou ID.",
                List.of(
                        new McpSchema.PromptArgument("query", "Termo de busca — pode ser nome, telefone ou ID do cliente", true)
                )
        );

        // Handler que recebe os argumentos e retorna as mensagens do prompt
        java.util.function.Function<McpSchema.GetPromptRequest, McpSchema.GetPromptResult> handler = request -> {
            String query = request.arguments() != null
                    ? request.arguments().getOrDefault("query", "").toString()
                    : "";

            String promptText = """
                    Você é um assistente especializado em gerenciamento de clientes.
                    
                    O usuário quer buscar clientes com o seguinte termo: "%s"
                    
                    Siga estas instruções:
                    1. Analise o termo de busca para identificar se é um ID, nome ou telefone
                    2. Use a tool `searchCustomer` com o parâmetro apropriado
                    3. Se o termo parece ser um ID (ex: UUID ou código alfanumérico), use o parâmetro `id`
                    4. Se parece ser um nome, use o parâmetro `name`
                    5. Se parece ser um telefone (contém números e/ou caracteres como +, -, (), espaços), use `phone`
                    6. Apresente os resultados de forma clara, listando id, nome e telefone de cada cliente encontrado
                    7. Se nenhum cliente for encontrado, informe ao usuário e sugira refinamentos na busca
                    """.formatted(query);

            return new McpSchema.GetPromptResult(
                    "Template para busca de clientes por query",
                    List.of(new McpSchema.PromptMessage(
                            McpSchema.Role.USER,
                            new McpSchema.TextContent(promptText)
                    ))
            );
        };

        return new McpServerFeatures.SyncPromptRegistration(promptDefinition, handler);
    }

    /**
     * Prompt: create-customer-prompt
     * Gera uma mensagem estruturada para que o LLM crie um cliente com validação básica.
     * Recebe nome e telefone e orienta o LLM a usar a tool createCustomer corretamente.
     */
    @Bean
    public McpServerFeatures.SyncPromptRegistration createCustomerPrompt() {
        // Definição do prompt com parâmetros obrigatórios de criação
        var promptDefinition = new McpSchema.Prompt(
                "create-customer-prompt",
                "Template para criação de clientes. Use quando o usuário quiser cadastrar um novo cliente fornecendo nome e telefone.",
                List.of(
                        new McpSchema.PromptArgument("name", "Nome completo do cliente a ser cadastrado", true),
                        new McpSchema.PromptArgument("phone", "Telefone de contato do cliente", true)
                )
        );

        // Handler que constrói o prompt de criação com os dados fornecidos
        java.util.function.Function<McpSchema.GetPromptRequest, McpSchema.GetPromptResult> handler = request -> {
            Map<String, Object> args = request.arguments() != null ? request.arguments() : Map.of();
            String name = args.getOrDefault("name", "").toString();
            String phone = args.getOrDefault("phone", "").toString();

            String promptText = """
                    Você é um assistente especializado em gerenciamento de clientes.
                    
                    O usuário quer cadastrar um novo cliente com os seguintes dados:
                    - Nome: %s
                    - Telefone: %s
                    
                    Siga estas instruções:
                    1. Verifique se o nome está preenchido e não é vazio
                    2. Verifique se o telefone está preenchido e não é vazio
                    3. Se algum dado estiver ausente, informe o usuário e solicite o dado faltante
                    4. Se os dados estiverem completos, use a tool `createCustomer` com name="%s" e phone="%s"
                    5. Após criar, confirme o sucesso ao usuário informando o ID gerado para o novo cliente
                    6. Em caso de erro, informe o usuário com a mensagem de erro retornada
                    """.formatted(name, phone, name, phone);

            return new McpSchema.GetPromptResult(
                    "Template para criação de novo cliente",
                    List.of(new McpSchema.PromptMessage(
                            McpSchema.Role.USER,
                            new McpSchema.TextContent(promptText)
                    ))
            );
        };

        return new McpServerFeatures.SyncPromptRegistration(promptDefinition, handler);
    }
}
