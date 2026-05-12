package com.iadeva.memory;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Ponto de entrada da aplicação de memória para agentes.
 * Sistema com 4 tipos de memória + embeddings + reflexão evolutiva.
 * Equivalente às aulas 13-15 do módulo 04 Python.
 */
@SpringBootApplication
public class MemoryAgentApplication {

    public static void main(String[] args) {
        SpringApplication.run(MemoryAgentApplication.class, args);
    }
}
