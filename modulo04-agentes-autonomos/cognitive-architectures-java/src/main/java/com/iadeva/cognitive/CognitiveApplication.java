package com.iadeva.cognitive;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Ponto de entrada da aplicação de arquiteturas cognitivas.
 * Implementa ReAct, Plan-Execute e Reflection usando Spring AI.
 */
@SpringBootApplication
public class CognitiveApplication {

    public static void main(String[] args) {
        SpringApplication.run(CognitiveApplication.class, args);
    }
}
