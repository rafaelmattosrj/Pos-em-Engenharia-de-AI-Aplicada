package com.iadeva.bragbot;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Ponto de entrada — equivalente a server.ts (a parte de API do backend
 * Express/Angular SSR de brag-bot). A UI Angular permanece fora de escopo;
 * apenas a rota POST /api/brag e o flow de geração via Gemini são portados.
 */
@SpringBootApplication
public class BragBotApplication {
    public static void main(String[] args) {
        SpringApplication.run(BragBotApplication.class, args);
    }
}
