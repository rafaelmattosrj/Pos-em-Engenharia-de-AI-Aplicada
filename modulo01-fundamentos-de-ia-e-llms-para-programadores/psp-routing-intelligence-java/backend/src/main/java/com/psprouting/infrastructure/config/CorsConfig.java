package com.psprouting.infrastructure.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * Libera chamadas do frontend Vite (porta padrao 5173) descrito em IDEIA.md.
 *
 * O frontend React em si nao foi implementado nesta porta — apenas o
 * backend/API foi escopado (ver README) — mas o CORS fica configurado para
 * que um cliente local (frontend real, Postman, curl a partir de outra
 * origem) consiga chamar a API sem bloqueio do navegador.
 */
@Configuration
public class CorsConfig implements WebMvcConfigurer {

    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/api/**")
                .allowedOriginPatterns("*")
                .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
                .allowedHeaders("*");
    }
}
