package com.iadeva.customer.security;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.AuthenticationEntryPoint;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

/**
 * Configuração central do Spring Security.
 * Equivalente ao middleware de autenticação e autorização do módulo 07 do curso JS/TS.
 *
 * Regras:
 * - /auth/**  → público (sem autenticação)
 * - /health   → público
 * - GET /customers e GET /customers/{id} → MEMBER ou ADMIN
 * - POST/PUT/DELETE /customers/** → somente ADMIN
 * - Qualquer outra rota → autenticado
 */
@Configuration
@EnableWebSecurity
public class SecurityConfig {

    private final JwtFilter jwtFilter;
    private final RateLimitFilter rateLimitFilter;

    public SecurityConfig(JwtFilter jwtFilter, RateLimitFilter rateLimitFilter) {
        this.jwtFilter = jwtFilter;
        this.rateLimitFilter = rateLimitFilter;
    }

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        // EntryPoint para retornar 401 (e não 403) quando não há autenticação
        AuthenticationEntryPoint unauthorizedHandler = (request, response, ex) -> {
            response.setStatus(HttpStatus.UNAUTHORIZED.value());
            response.setContentType("application/json");
            response.getWriter().write("{\"error\": \"Não autenticado\"}");
        };

        http
            // Desabilita CSRF (API stateless não usa cookies de sessão)
            .csrf(AbstractHttpConfigurer::disable)
            // Stateless: sem HttpSession
            .sessionManagement(sm ->
                sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            // Regras de autorização
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/auth/**").permitAll()
                .requestMatchers("/health").permitAll()
                // Leitura: MEMBER e ADMIN
                .requestMatchers(HttpMethod.GET, "/customers", "/customers/**").hasAnyRole("MEMBER", "ADMIN")
                // Escrita: somente ADMIN
                .requestMatchers(HttpMethod.POST, "/customers").hasRole("ADMIN")
                .requestMatchers(HttpMethod.PUT, "/customers/**").hasRole("ADMIN")
                .requestMatchers(HttpMethod.DELETE, "/customers/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            // Tratamento de exceções: 401 para não autenticado, 403 para não autorizado
            .exceptionHandling(ex -> ex.authenticationEntryPoint(unauthorizedHandler))
            // Registra os filtros customizados antes do filtro padrão de autenticação
            .addFilterBefore(rateLimitFilter, UsernamePasswordAuthenticationFilter.class)
            .addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }
}
