package com.iadeva.nexustools;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Ponto de entrada — expõe como endpoints HTTP as ferramentas
 * (tools/*.py, decoradas com @tool do CrewAI) do projeto Nexus AIOps do
 * módulo 06 do curso. A orquestração multi-agente hierárquica do CrewAI
 * (core/agents.py, labs/*.py) não é portada — apenas a lógica de negócio
 * de cada ferramenta, que na versão Python seria invocada pelo LLM via
 * function calling.
 */
@SpringBootApplication
public class NexusToolsApplication {
    public static void main(String[] args) {
        SpringApplication.run(NexusToolsApplication.class, args);
    }
}
