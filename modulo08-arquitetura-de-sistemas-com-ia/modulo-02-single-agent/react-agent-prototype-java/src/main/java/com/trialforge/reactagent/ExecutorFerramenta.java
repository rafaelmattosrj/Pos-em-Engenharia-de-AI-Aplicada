package com.trialforge.reactagent;

import java.util.List;
import java.util.Locale;
import java.util.Map;

/**
 * Execucao real da ferramenta — sempre deterministica, nunca e o modelo quem
 * executa. Busca por palavra-chave, nao string exata: o modelo formula o
 * "tema" em linguagem natural (ex.: "Assentimento para menores de idade em
 * estudos clínicos"), entao a busca real por tras dessa ferramenta seria
 * vetorial (RAG) — aqui simulamos isso com um match por palavra-chave em vez
 * de comparacao exata.
 *
 * Porte 1:1 de executarBuscaClausula em react-agent-prototype.js / .py.
 */
public final class ExecutorFerramenta {

    private ExecutorFerramenta() {
    }

    public record ResultadoBusca(String texto, String fonte, String aviso) {
        static ResultadoBusca semClausula(String aviso) {
            return new ResultadoBusca(null, null, aviso);
        }

        static ResultadoBusca comClausula(String texto, String fonte) {
            return new ResultadoBusca(texto, fonte, null);
        }
    }

    // Testado de verdade: o modelo as vezes formula o tema com "adolescente" ou
    // "pediátrico" em vez de "menor" — um match de palavra unica era fragil demais
    // pra depender da sorte da frase exata que o modelo escolhe.
    private static final List<String> PALAVRAS_MENOR_DE_IDADE = List.of("menor", "adolescente", "pediátric");

    /**
     * Recebe um Map bruto (equivalente aos **kwargs do Python / objeto solto do JS) em
     * vez de (tema, jurisdicao) fixos de proposito: o modelo as vezes formula a chamada
     * com um parametro faltando ou grafado errado (testado de verdade: ja vimos
     * "jurisdicicao" em vez de "jurisdicao"). Validar aqui, em vez de confiar cegamente,
     * evita um crash silencioso ou um resultado errado escapar sem ninguem perceber — a
     * ferramenta trata isso como falha propria, nunca deixa o erro estourar pro chamador.
     */
    public static ResultadoBusca executarBuscaClausula(Map<String, Object> argumentosBrutos) {
        Object temaObj = argumentosBrutos == null ? null : argumentosBrutos.get("tema");
        Object jurisdicaoObj = argumentosBrutos == null ? null : argumentosBrutos.get("jurisdicao");

        if (!(temaObj instanceof String tema) || !(jurisdicaoObj instanceof String jurisdicao)) {
            return ResultadoBusca.semClausula(
                    "parâmetro inválido ou ausente na chamada da ferramenta: " + argumentosBrutos);
        }

        String temaNormalizado = tema.toLowerCase(Locale.ROOT);
        boolean mencionaMenorDeIdade = PALAVRAS_MENOR_DE_IDADE.stream().anyMatch(temaNormalizado::contains);

        if ("ANVISA".equals(jurisdicao) && mencionaMenorDeIdade) {
            return ResultadoBusca.comClausula(
                    "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, "
                            + "além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
                    "RDC ANVISA 466/2012, Art. 4º");
        }

        return ResultadoBusca.semClausula(
                "não encontrado: nenhuma cláusula sobre \"" + tema + "\" na jurisdição " + jurisdicao
                        + ". Tente outra jurisdição ou revise o tema.");
    }
}
