package com.iadeva.customer.model;

/**
 * Payload de requisição de login.
 * Equivalente ao body { username, password } do POST /auth/login no módulo 07 do curso JS/TS.
 */
public record AuthRequest(String username, String password) {}
