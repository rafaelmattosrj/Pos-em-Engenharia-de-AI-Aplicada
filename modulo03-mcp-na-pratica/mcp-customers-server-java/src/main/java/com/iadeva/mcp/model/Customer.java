package com.iadeva.mcp.model;

// Representa um cliente da legacy API — mapeado a partir do JSON retornado pela API legada
// Equivalente à interface/type Customer dos projetos TypeScript do curso

/**
 * Modelo imutável de cliente usando Java Record.
 *
 * @param id    identificador único do cliente
 * @param name  nome completo do cliente
 * @param phone telefone de contato do cliente
 */
public record Customer(
        String id,
        String name,
        String phone
) {
}
