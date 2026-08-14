package com.trialforge.agentcomponents;

import java.util.List;

/**
 * 3. FERRAMENTAS — funcao deterministica que o agente aciona.
 *
 * Versao minima, so pra tornar tangivel "ferramenta = funcao deterministica que o
 * agente delega". O schema formal (JSON, enum, validacao de parametro) fica a cargo
 * de react-agent-prototype — aqui e so a peca em isolamento.
 *
 * Porte 1:1 de agent-components-demo.js / agent_components_demo.py (secao 3).
 */
public final class Ferramentas {

    private Ferramentas() {
    }

    public record ResultadoClausula(String texto, String fonte, String aviso) {
        static ResultadoClausula semClausula(String aviso) {
            return new ResultadoClausula(null, null, aviso);
        }

        static ResultadoClausula comClausula(String texto, String fonte) {
            return new ResultadoClausula(texto, fonte, null);
        }
    }

    public static ResultadoClausula buscarClausulaAssentimento(List<Integer> faixaEtaria) {
        boolean cobreMenor = faixaEtaria.stream().anyMatch(idade -> idade < 18);
        if (!cobreMenor) {
            return ResultadoClausula.semClausula("população adulta: cláusula de assentimento não se aplica");
        }
        return ResultadoClausula.comClausula(
                "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, além do "
                        + "consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
                "RDC ANVISA 466/2012, Art. 4º"
        );
    }
}
