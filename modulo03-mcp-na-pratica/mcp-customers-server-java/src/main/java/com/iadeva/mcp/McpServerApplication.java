package com.iadeva.mcp;

// Equivalente ao index.ts dos projetos MCP do curso — ponto de entrada do servidor MCP
// Inicializa o contexto Spring e registra tools, resources e prompts via auto-configuração do Spring AI MCP

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class McpServerApplication {

    public static void main(String[] args) {
        SpringApplication.run(McpServerApplication.class, args);
    }
}
