package com.decisionframework;

/**
 * As quatro classificacoes possiveis do checklist, exatamente como o texto original
 * (decision-framework-checklist.md). Centralizadas aqui para evitar divergencia de
 * string entre a logica e os testes/demo — equivalente ao objeto CLASSIFICACAO da
 * versao JavaScript e as strings literais da versao Python.
 *
 * <p>Modelado como enum (em vez de constantes String soltas, como no original JS) por
 * ser mais idiomatico em Java: type-safety em tempo de compilacao, sem custo de
 * paridade de comportamento, ja que {@link #toString()} devolve exatamente o mesmo
 * texto que as versoes JS/Python retornam.</p>
 */
public enum Classificacao {

    REGRA_DETERMINISTICA("Regra determinística"),
    APPROVAL_GATE_OBRIGATORIO("Agente com Approval Gate obrigatório"),
    AGENTE_AUTONOMO("Agente autônomo, com observabilidade completa"),
    REGRA_ENUMERAVEL("Regra determinística (mesmo parecendo complexa, se é enumerável, é regra)");

    private final String descricao;

    Classificacao(String descricao) {
        this.descricao = descricao;
    }

    @Override
    public String toString() {
        return descricao;
    }
}
