package com.trialforge.evalgate;

/**
 * Item do golden set: pergunta com cláusula esperada CONHECIDA de antemão
 * (não é RAG buscando a cláusula — aqui a cláusula certa já está fixada, só
 * se mede se o modelo, dado o contexto certo, produz uma resposta fiel a ele).
 */
public record GoldenItem(String pergunta, String clausula, String fonte) {
}
