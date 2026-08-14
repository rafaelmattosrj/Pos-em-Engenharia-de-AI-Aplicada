package com.trialforge.reactagent;

import java.util.Map;

/** Uma chamada de ferramenta decidida pelo modelo — equivalente a
 *  resposta.message.tool_calls[0] no original. */
public record ChamadaFerramenta(String id, String nome, Map<String, Object> argumentos) {
}
