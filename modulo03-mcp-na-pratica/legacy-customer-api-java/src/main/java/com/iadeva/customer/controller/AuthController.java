package com.iadeva.customer.controller;

import com.iadeva.customer.model.AuthRequest;
import com.iadeva.customer.model.AuthResponse;
import com.iadeva.customer.service.AuthService;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;
import java.util.Optional;

/**
 * Controller de autenticação.
 * Equivalente ao authRouter.ts / auth.ts do módulo 07 do curso JS/TS.
 *
 * Endpoints:
 * - POST /auth/login           → valida usuário/senha, retorna JWT + role
 * - POST /auth/service-token   → requer header X-Super-Secret, retorna UUID service token
 */
@RestController
@RequestMapping("/auth")
public class AuthController {

    private final AuthService authService;

    @Value("${app.super-secret}")
    private String expectedSuperSecret;

    public AuthController(AuthService authService) {
        this.authService = authService;
    }

    /**
     * POST /auth/login
     * Body: { "username": "...", "password": "..." }
     * Retorna: { "token": "...", "role": "ADMIN|MEMBER" }
     */
    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody AuthRequest request) {
        Optional<String> tokenOpt = authService.login(request.username(), request.password());

        if (tokenOpt.isEmpty()) {
            return ResponseEntity
                    .status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Credenciais inválidas"));
        }

        String token = tokenOpt.get();
        // Determina a role para incluir na resposta
        String role = authService.validateToken(token).orElse("UNKNOWN");

        return ResponseEntity.ok(new AuthResponse(token, role));
    }

    /**
     * POST /auth/service-token
     * Header: X-Super-Secret: superSecret123
     * Retorna: { "token": "<UUID>" }
     */
    @PostMapping("/service-token")
    public ResponseEntity<?> serviceToken(
            @RequestHeader(value = "X-Super-Secret", required = false) String superSecret) {

        if (superSecret == null || !expectedSuperSecret.equals(superSecret)) {
            return ResponseEntity
                    .status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Header X-Super-Secret ausente ou inválido"));
        }

        String token = authService.createServiceToken();
        return ResponseEntity.ok(Map.of("token", token));
    }
}
