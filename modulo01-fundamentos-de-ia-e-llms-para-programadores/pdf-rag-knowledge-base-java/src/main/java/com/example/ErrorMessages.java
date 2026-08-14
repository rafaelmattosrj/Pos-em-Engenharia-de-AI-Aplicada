package com.example;

/**
 * Utilitario para truncar mensagens de erro exibidas ao usuario — equivalente
 * ao truncamento em 200 caracteres feito em Main.java (e em main.go, no porte
 * Go irmao) ao reportar falhas do LLM.
 *
 * Extraido para corrigir e testar isoladamente o caso de mensagem nula
 * (a versao original de Main.java chamava {@code e.getMessage().substring(...)}
 * diretamente, o que lançaria NullPointerException caso a excecao nao
 * tivesse mensagem).
 */
public final class ErrorMessages {

    private ErrorMessages() {
    }

    public static String truncate(String message, int maxLength) {
        if (message == null) {
            return "";
        }
        return message.length() > maxLength ? message.substring(0, maxLength) : message;
    }
}
