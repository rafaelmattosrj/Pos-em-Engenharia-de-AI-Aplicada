package com.decisionframework;

import java.util.ArrayList;
import java.util.List;

/**
 * Framework de Decisao de Tres Perguntas (checklist do Modulo 1.3), portado de
 * decision-framework-tool.js / decision_framework_tool.py (mesma pasta do modulo
 * original, {@code modulo-01-fundamentos-ai-first}).
 *
 * <p>Logica pura, 100% deterministica: nao chama nenhum modelo, nao faz nenhuma
 * chamada de rede. E a arvore de decisao do checklist virando codigo, pergunta por
 * pergunta.</p>
 */
public final class DecisionFramework {

    private DecisionFramework() {
        // Classe utilitaria: apenas metodos estaticos, sem estado.
    }

    /**
     * Aplica as tres perguntas do checklist, em ordem, exatamente como a arvore:
     * Pergunta 1 decide sozinha quando e {@code true} (nem chega a olhar p2 ou p3);
     * senao passa para a Pergunta 2; senao passa para a Pergunta 3.
     *
     * @param p1 Existe uma regra finita que cobre mais de 90% dos casos REAIS (ja
     *           observados, nao hipoteticos)?
     * @param p2 O erro e caro E a acao e irreversivel (nao da pra desfazer depois)?
     * @param p3 O comportamento da tarefa muda de acordo com o contexto de entrada?
     * @return uma das quatro classificacoes do checklist
     */
    public static Classificacao classificarTarefa(boolean p1, boolean p2, boolean p3) {
        if (p1) {
            return Classificacao.REGRA_DETERMINISTICA;
        }
        if (p2) {
            return Classificacao.APPROVAL_GATE_OBRIGATORIO;
        }
        if (p3) {
            return Classificacao.AGENTE_AUTONOMO;
        }
        return Classificacao.REGRA_ENUMERAVEL;
    }

    /**
     * Decompoe uma tarefa hibrida em subtarefas ja classificadas, seguindo o
     * "Template: decompondo uma tarefa hibrida" do checklist. Nao e uma logica nova:
     * aplica {@link #classificarTarefa} em cada subtarefa, uma de cada vez, e devolve
     * uma lista nova com o campo classificacao preenchido — as subtarefas de entrada
     * nao sao modificadas (records sao imutaveis por natureza, entao esse
     * comportamento e garantido pelo compilador, nao so por convencao).
     *
     * @param subtarefas lista de subtarefas a classificar
     * @return nova lista, na mesma ordem, com cada subtarefa classificada
     */
    public static List<SubtarefaClassificada> decomporTarefaHibrida(List<Subtarefa> subtarefas) {
        List<SubtarefaClassificada> resultado = new ArrayList<>(subtarefas.size());
        for (Subtarefa subtarefa : subtarefas) {
            Classificacao classificacao = classificarTarefa(subtarefa.p1(), subtarefa.p2(), subtarefa.p3());
            resultado.add(SubtarefaClassificada.de(subtarefa, classificacao));
        }
        return resultado;
    }
}
