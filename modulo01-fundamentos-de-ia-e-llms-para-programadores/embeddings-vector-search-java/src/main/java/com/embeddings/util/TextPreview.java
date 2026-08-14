package com.embeddings.util;

/**
 * Utilitario de preview de texto: trunca um texto longo para exibicao,
 * preservando os primeiros {@code maxLength} caracteres e sinalizando
 * o corte com "...".
 *
 * Extraido do corpo de Main para permitir teste unitario isolado.
 */
public final class TextPreview {

    private TextPreview() {
    }

    /**
     * Trunca {@code text} para no maximo {@code maxLength} caracteres,
     * adicionando "..." ao final quando o texto original for maior.
     * Textos com tamanho menor ou igual a {@code maxLength} sao
     * retornados inalterados.
     */
    public static String truncate(String text, int maxLength) {
        if (text == null) {
            return null;
        }
        if (text.length() > maxLength) {
            return text.substring(0, maxLength) + "...";
        }
        return text;
    }
}
