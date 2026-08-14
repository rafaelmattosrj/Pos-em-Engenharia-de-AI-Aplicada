package com.iadeva.cfp;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Ponto de entrada da aplicação — equivalente a main.ts do backend NestJS
 * de cfp-platform/cfp-platform_v1 (módulo 05 do curso, UI/UX).
 *
 * Ainda não é um servidor de produção — é apenas o backend mínimo para
 * sustentar o front-end Angular do CFP (Call for Papers).
 */
@SpringBootApplication
public class CfpPlatformApplication {
    public static void main(String[] args) {
        SpringApplication.run(CfpPlatformApplication.class, args);
    }
}
