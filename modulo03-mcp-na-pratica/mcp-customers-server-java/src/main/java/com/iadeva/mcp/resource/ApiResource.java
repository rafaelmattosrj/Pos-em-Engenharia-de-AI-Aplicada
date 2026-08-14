package com.iadeva.mcp.resource;

// Equivalente ao resources/ do curso MCP — expõe documentação como recurso legível pelo LLM
// O resource "info://api" fornece ao LLM contexto sobre a legacy API sem precisar chamar uma tool

import io.modelcontextprotocol.server.McpServerFeatures;
import io.modelcontextprotocol.spec.McpSchema;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.List;

/**
 * Configuração dos resources MCP expostos pelo servidor.
 * Resources são conteúdos estáticos ou dinâmicos que o LLM pode ler para obter contexto.
 * Neste caso expõe documentação da API legada incluindo URL base, endpoints e formato de autenticação.
 */
@Configuration
public class ApiResource {

    @Value("${app.legacy-api.base-url:http://localhost:3000}")
    private String baseUrl;

    /**
     * Resource "info://api" — documentação completa da legacy customer API.
     * O LLM pode consultar este resource para entender a estrutura da API antes de chamar tools.
     */
    @Bean
    public McpServerFeatures.SyncResourceRegistration apiInfoResource() {
        // Conteúdo do resource com documentação da API legada
        String resourceContent = """
                # Legacy Customer API — Documentação
                
                ## URL Base
                %s
                
                ## Autenticação
                Todas as requisições requerem header:
                  Authorization: Bearer <service_token>
                
                O token é configurado via variável de ambiente API_SERVICE_TOKEN.
                
                ## Endpoints Disponíveis
                
                ### GET /customers
                - Descrição: Lista todos os clientes cadastrados
                - Resposta: Array de objetos Customer [ { id, name, phone } ]
                
                ### GET /customers/:id
                - Descrição: Retorna um cliente específico pelo ID
                - Parâmetros: id (path) — identificador único do cliente
                - Resposta: Objeto Customer { id, name, phone }
                - Erro 404: cliente não encontrado
                
                ### POST /customers
                - Descrição: Cria um novo cliente
                - Corpo: { "name": string, "phone": string }
                - Resposta: Objeto Customer criado com ID gerado { id, name, phone }
                
                ### PUT /customers/:id
                - Descrição: Atualiza dados de um cliente existente
                - Parâmetros: id (path) — identificador único do cliente
                - Corpo: { "name"?: string, "phone"?: string } (campos opcionais)
                - Resposta: Objeto Customer atualizado { id, name, phone }
                
                ### DELETE /customers/:id
                - Descrição: Remove permanentemente um cliente
                - Parâmetros: id (path) — identificador único do cliente
                - Resposta: 204 No Content em caso de sucesso
                
                ## Formato dos Dados
                
                ```json
                {
                  "id": "uuid-ou-string-gerado-pela-api",
                  "name": "Nome Completo do Cliente",
                  "phone": "+55 11 99999-9999"
                }
                ```
                
                ## Observações
                - A API legada NÃO possui endpoint de busca/filtro nativo
                - Filtros por nome ou telefone são implementados no servidor MCP em memória
                - IDs são gerados automaticamente pela API legada na criação
                """.formatted(baseUrl);

        // Define o resource MCP com URI, nome, descrição e mime type
        var resourceDefinition = new McpSchema.Resource(
                "info://api",
                "Legacy Customer API Documentation",
                "Documentação completa da legacy customer API: endpoints, autenticação e formato dos dados",
                "text/markdown",
                null
        );

        // Handler que retorna o conteúdo do resource quando o LLM o solicitar
        java.util.function.Function<McpSchema.ReadResourceRequest, McpSchema.ReadResourceResult> handler =
                request -> new McpSchema.ReadResourceResult(
                        List.of(new McpSchema.TextResourceContents(
                                "info://api",
                                "text/markdown",
                                resourceContent
                        ))
                );

        return new McpServerFeatures.SyncResourceRegistration(resourceDefinition, handler);
    }
}
