package com.iadeva.customer.service;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.util.Date;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Serviço de autenticação e geração de tokens.
 * Equivalente ao authService.ts / generateToken() do módulo 07 do curso JS/TS.
 *
 * Responsabilidades:
 * - Validar credenciais de usuários em memória
 * - Gerar JWT assinado com HS256 contendo claim "role"
 * - Gerar service tokens (UUID) via header X-Super-Secret
 * - Validar tokens (JWT ou service token)
 */
@Service
public class AuthService {

    @Value("${app.jwt.secret}")
    private String jwtSecret;

    @Value("${app.jwt.expiration}")
    private long jwtExpirationMs;

    /**
     * Usuários em memória: username → { password, role }
     * Equivalente ao objeto USERS no módulo 07 do curso JS/TS.
     */
    private static final Map<String, String[]> USERS = Map.of(
            "admin",  new String[]{"password123", "ADMIN"},
            "member", new String[]{"pass456",     "MEMBER"}
    );

    /**
     * Armazena service tokens válidos (UUID → role).
     * Equivalente ao Set/Map de serviceTokens no módulo 07 do curso JS/TS.
     */
    private final Map<String, String> serviceTokens = new ConcurrentHashMap<>();

    /**
     * Valida credenciais e retorna o JWT se válidas.
     * Retorna Optional.empty() se inválidas.
     */
    public Optional<String> login(String username, String password) {
        String[] userInfo = USERS.get(username);
        if (userInfo == null || !userInfo[0].equals(password)) {
            return Optional.empty();
        }
        String role = userInfo[1];
        String token = generateJwt(username, role);
        return Optional.of(token);
    }

    /**
     * Retorna a role do usuário sem gerar token (usado internamente).
     */
    public Optional<String> getRoleForUser(String username) {
        String[] info = USERS.get(username);
        return info != null ? Optional.of(info[1]) : Optional.empty();
    }

    /**
     * Gera um service token UUID e armazena internamente com role ADMIN.
     * Requer o segredo correto (X-Super-Secret).
     * Nota: validação do segredo é feita pelo AuthController.
     */
    public String createServiceToken() {
        String token = UUID.randomUUID().toString();
        serviceTokens.put(token, "ADMIN");
        return token;
    }

    /**
     * Valida um token (JWT ou service token) e retorna a role.
     * Retorna Optional.empty() se inválido.
     */
    public Optional<String> validateToken(String token) {
        // Primeiro tenta como service token
        if (serviceTokens.containsKey(token)) {
            return Optional.of(serviceTokens.get(token));
        }
        // Depois tenta como JWT
        try {
            SecretKey key = Keys.hmacShaKeyFor(jwtSecret.getBytes(StandardCharsets.UTF_8));
            var claims = Jwts.parser()
                    .verifyWith(key)
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();
            String role = claims.get("role", String.class);
            return Optional.ofNullable(role);
        } catch (Exception e) {
            return Optional.empty();
        }
    }

    /**
     * Extrai o subject (username) de um JWT sem validar role.
     */
    public Optional<String> extractUsername(String token) {
        try {
            SecretKey key = Keys.hmacShaKeyFor(jwtSecret.getBytes(StandardCharsets.UTF_8));
            String subject = Jwts.parser()
                    .verifyWith(key)
                    .build()
                    .parseSignedClaims(token)
                    .getPayload()
                    .getSubject();
            return Optional.ofNullable(subject);
        } catch (Exception e) {
            return Optional.empty();
        }
    }

    /**
     * Retorna true se o token for um service token registrado.
     */
    public boolean isServiceToken(String token) {
        return serviceTokens.containsKey(token);
    }

    // -------------------------------------------------------------------------
    // Privado
    // -------------------------------------------------------------------------

    private String generateJwt(String username, String role) {
        SecretKey key = Keys.hmacShaKeyFor(jwtSecret.getBytes(StandardCharsets.UTF_8));
        return Jwts.builder()
                .subject(username)
                .claim("role", role)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + jwtExpirationMs))
                .signWith(key)
                .compact();
    }
}
