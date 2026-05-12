package com.iadeva.customer.model;

/**
 * Payload de resposta da autenticação.
 * Equivalente ao objeto { token, role } retornado pelo POST /auth/login no módulo 07 do curso JS/TS.
 */
public record AuthResponse(String token, String role) {}
