package com.trialforge.gateway;

/**
 * Resultado de uma busca hibrida dentro de um indice. {@code iteracoesUsadas}
 * e {@code esgotouLimite} só fazem sentido depois que o Agentic RAG decide
 * parar — {@link #bruto} cria o resultado "cru" de uma busca hibrida isolada,
 * {@link #comIteracao} anexa a informacao de quantas iteracoes foram usadas.
 */
public record ResultadoBusca(
        String indice,
        Clausula clausula,
        double similaridadeCosseno,
        double scoreBM25,
        double scoreRRF,
        int iteracoesUsadas,
        boolean esgotouLimite) {

    public static ResultadoBusca bruto(String indice, Clausula clausula, double similaridadeCosseno,
            double scoreBM25, double scoreRRF) {
        return new ResultadoBusca(indice, clausula, similaridadeCosseno, scoreBM25, scoreRRF, 0, false);
    }

    public ResultadoBusca comIteracao(int iteracoesUsadas, boolean esgotouLimite) {
        return new ResultadoBusca(indice, clausula, similaridadeCosseno, scoreBM25, scoreRRF, iteracoesUsadas,
                esgotouLimite);
    }
}
