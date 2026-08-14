package com.decisionframework;

/**
 * Uma subtarefa ainda nao classificada, seguindo o template "Decompondo uma tarefa
 * hibrida" do checklist: nome livre, tipo livre (ex.: "Extração/Interpretação" ou
 * "Decisão de Negócio") e as tres respostas booleanas do framework de decisao.
 *
 * <p>Record, e nao classe mutavel: equivalente ao dict de entrada da versao Python,
 * que a funcao de decomposicao explicitamente nao modifica (ver
 * {@link DecisionFramework#decomporTarefaHibrida}).</p>
 */
public record Subtarefa(String nome, String tipo, boolean p1, boolean p2, boolean p3) {
}
