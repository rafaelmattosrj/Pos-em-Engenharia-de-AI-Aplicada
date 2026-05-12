package com.iadeva.mcp.config;

// Equivalente ao server setup no index.ts do curso MCP
// Configura o servidor MCP com informações de identificação e registra todos os componentes

import com.iadeva.mcp.prompt.CustomerPrompts;
import com.iadeva.mcp.resource.ApiResource;
import com.iadeva.mcp.tool.CustomerTools;
import io.modelcontextprotocol.server.McpServerFeatures;
import org.springframework.ai.tool.ToolCallbackProvider;
import org.springframework.ai.tool.method.MethodToolCallbackProvider;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Configuração central do servidor MCP.
 * Responsável por montar o ToolCallbackProvider que registra todos os métodos
 * anotados com @Tool na classe CustomerTools como ferramentas MCP invocáveis.
 *
 * Resources e Prompts são registrados automaticamente via seus próprios beans
 * do tipo McpServerFeatures.SyncResourceRegistration e
 * McpServerFeatures.SyncPromptRegistration, respectivamente,
 * detectados pelo auto-configurador do Spring AI MCP Server.
 */
@Configuration
public class McpConfig {

    /**
     * Registra todas as tools de CustomerTools no servidor MCP.
     * O MethodToolCallbackProvider inspeciona os métodos anotados com @Tool
     * e os disponibiliza como ferramentas invocáveis pelo LLM via protocolo MCP.
     *
     * @param customerTools instância do componente com as tools definidas
     * @return provider que expõe as tools ao servidor MCP
     */
    @Bean
    public ToolCallbackProvider customerToolCallbackProvider(CustomerTools customerTools) {
        return MethodToolCallbackProvider.builder()
                .toolObjects(customerTools)
                .build();
    }
}
