package com.iadeva.evals;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Ponto de entrada do framework de avaliação de agentes.
 * Equivalente ao conjunto de scripts eval_*.py do módulo 04 Python,
 * porém exposto como API REST para integração com CI/CD.
 */
@SpringBootApplication
public class EvalsApplication {

    public static void main(String[] args) {
        SpringApplication.run(EvalsApplication.class, args);
    }
}
