package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.List;

/**
 * Resposta do modelo pra uma volta do loop ReAct.
 *
 * {@code mensagemBruta} guarda o node JSON exato devolvido pelo Ollama (equivalente a
 * {@code resposta.message} no original) — precisa ser reempurrado pro historico tal
 * como veio, tool_calls inclusos, pra proxima volta do loop fazer sentido pro modelo.
 */
public record ModeloResposta(ObjectNode mensagemBruta, String content, List<ChamadaFerramenta> chamadas) {

    public boolean temChamadaFerramenta() {
        return chamadas != null && !chamadas.isEmpty();
    }
}
