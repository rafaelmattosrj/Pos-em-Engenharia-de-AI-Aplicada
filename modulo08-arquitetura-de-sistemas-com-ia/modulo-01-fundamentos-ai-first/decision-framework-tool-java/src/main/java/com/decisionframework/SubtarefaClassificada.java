package com.decisionframework;

/**
 * O mesmo conteudo de {@link Subtarefa}, com o campo a mais "classificacao" —
 * resultado de aplicar {@link DecisionFramework#classificarTarefa} naquela
 * subtarefa especifica. Equivalente ao dict novo (`{**subtarefa, "classificacao": ...}`)
 * devolvido pela versao Python, ou ao objeto spread (`{...subtarefa, classificacao}`)
 * da versao JavaScript.
 */
public record SubtarefaClassificada(
        String nome,
        String tipo,
        boolean p1,
        boolean p2,
        boolean p3,
        Classificacao classificacao) {

    /**
     * Cria uma {@code SubtarefaClassificada} a partir de uma {@link Subtarefa} e da
     * classificacao ja calculada para ela, preservando nome/tipo/p1/p2/p3 originais.
     */
    public static SubtarefaClassificada de(Subtarefa subtarefa, Classificacao classificacao) {
        return new SubtarefaClassificada(
                subtarefa.nome(), subtarefa.tipo(), subtarefa.p1(), subtarefa.p2(), subtarefa.p3(), classificacao);
    }
}
