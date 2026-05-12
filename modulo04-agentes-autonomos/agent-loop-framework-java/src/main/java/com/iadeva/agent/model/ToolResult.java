package com.iadeva.agent.model;

/**
 * Resultado da execução de uma tool pelo Executor.
 * Equivalente ao ToolResult do executor.py no curso Python.
 *
 * @param toolName Nome da tool que foi executada
 * @param success  Indica se a execução foi bem-sucedida
 * @param output   Saída da tool em caso de sucesso
 * @param error    Mensagem de erro em caso de falha
 */
public record ToolResult(
        String toolName,
        boolean success,
        String output,
        String error
) {

    /** Fábrica de resultado de sucesso */
    public static ToolResult ok(String toolName, String output) {
        return new ToolResult(toolName, true, output, null);
    }

    /** Fábrica de resultado de erro */
    public static ToolResult fail(String toolName, String error) {
        return new ToolResult(toolName, false, null, error);
    }
}
