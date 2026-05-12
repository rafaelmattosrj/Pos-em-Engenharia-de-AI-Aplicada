package com.iadeva.customer.security;

import com.iadeva.customer.service.AuthService;
import io.github.bucket4j.Bandwidth;
import io.github.bucket4j.Bucket;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Filtro de rate limiting por token: 90 requisições por minuto.
 * Equivalente ao middleware de rate limiting com express-rate-limit do módulo 07 do curso JS/TS.
 *
 * Usa Bucket4j (token bucket algorithm):
 * - Um bucket por token de autenticação
 * - Capacidade: 90 tokens
 * - Recarga: 90 tokens a cada 60 segundos
 * - Retorna 429 Too Many Requests quando o bucket estiver vazio
 */
@Component
public class RateLimitFilter extends OncePerRequestFilter {

    /** Mapa de buckets: token → Bucket. ConcurrentHashMap para thread-safety. */
    private final Map<String, Bucket> buckets = new ConcurrentHashMap<>();

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        String authHeader = request.getHeader("Authorization");

        // Aplica rate limiting apenas para requisições autenticadas
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            String token = authHeader.substring(7);
            Bucket bucket = buckets.computeIfAbsent(token, this::createBucket);

            if (!bucket.tryConsume(1)) {
                response.setStatus(429);
                response.setContentType("application/json");
                response.getWriter().write(
                        "{\"error\": \"Too many requests. Limit: 90 req/min per token.\"}"
                );
                return;
            }
        }

        filterChain.doFilter(request, response);
    }

    /**
     * Cria um novo bucket com limite de 90 requisições por minuto.
     */
    private Bucket createBucket(String token) {
        Bandwidth limit = Bandwidth.builder()
                .capacity(90)
                .refillGreedy(90, Duration.ofMinutes(1))
                .build();
        return Bucket.builder()
                .addLimit(limit)
                .build();
    }
}
