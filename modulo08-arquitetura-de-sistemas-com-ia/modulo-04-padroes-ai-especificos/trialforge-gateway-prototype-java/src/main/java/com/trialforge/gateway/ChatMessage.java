package com.trialforge.gateway;

/**
 * Mensagem de chat (role/content) no formato aceito pelo Ollama.
 */
public record ChatMessage(String role, String content) {
}
