package com.trialforge.messagequeue;

import java.util.Map;

/**
 * O resultado do ICF antes de saber quantas tentativas o documento levou —
 * e exatamente o que fica gravado em {@link DocumentoRegistro#resultado()}.
 * {@link #comTentativas(int)} monta o {@link ResultadoICF} final devolvido
 * ao chamador, igual ao spread {@code {...resultado, tentativas}} do JS.
 */
public record ResultadoICFBase(String agente, String secao, Map<String, Integer> criterioUsado,
                                int versaoUsada, int versaoAoConcluir) {

    public ResultadoICF comTentativas(int tentativas) {
        return new ResultadoICF(agente, secao, criterioUsado, versaoUsada, versaoAoConcluir, tentativas);
    }
}
