package com.iadeva.customer.security;

import com.iadeva.customer.service.AuthService;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.List;

/**
 * Filtro que intercepta cada requisição e valida o token Bearer no header Authorization.
 * Equivalente ao middleware authenticateToken() do módulo 07 do curso JS/TS.
 *
 * Fluxo:
 * 1. Extrai "Bearer <token>" do header Authorization
 * 2. Valida via AuthService (suporta JWT e service tokens)
 * 3. Se válido, popula o SecurityContext com a role do usuário
 * 4. Se inválido, deixa a cadeia continuar sem autenticação (o SecurityConfig rejeita depois)
 */
@Component
public class JwtFilter extends OncePerRequestFilter {

    private final AuthService authService;

    public JwtFilter(AuthService authService) {
        this.authService = authService;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        String authHeader = request.getHeader("Authorization");

        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            String token = authHeader.substring(7);

            authService.validateToken(token).ifPresent(role -> {
                // Monta a authority com prefixo ROLE_ exigido pelo Spring Security
                var authority = new SimpleGrantedAuthority("ROLE_" + role);
                var authentication = new UsernamePasswordAuthenticationToken(
                        token, null, List.of(authority)
                );
                SecurityContextHolder.getContext().setAuthentication(authentication);
            });
        }

        filterChain.doFilter(request, response);
    }
}
