package com.iadeva.mcp.model;

// Resultado de operações de mutação (create, update, delete) — encapsula sucesso/falha e dados do cliente afetado
// Equivalente ao tipo de retorno das funções de mutação nos projetos TypeScript do curso

/**
 * Resultado de operações que modificam dados de clientes.
 *
 * @param success  indica se a operação foi realizada com sucesso
 * @param message  mensagem descritiva do resultado ou do erro
 * @param customer dados do cliente afetado (pode ser null em caso de delete ou erro)
 */
public record MutationResult(
        boolean success,
        String message,
        Customer customer
) {
}
