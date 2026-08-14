package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

/**
 * Ferramenta tipada (schema explicito, nunca texto livre) usada pelo loop ReAct.
 *
 * Porte 1:1 de BUSCAR_CLAUSULA_REGULATORIA em react-agent-prototype.js / .py.
 */
public final class ToolSchema {

    private ToolSchema() {
    }

    public static final String NOME_FERRAMENTA = "buscar_clausula_regulatoria";

    private static final String DESCRICAO =
            "Busca cláusulas regulatórias de estudos clínicos por tema e jurisdição. Use quando "
                    + "precisar de texto normativo (ANVISA ou FDA) para compor uma seção do documento.";

    /** Monta o schema no formato aceito pela API nativa de tools do Ollama (mesmo formato do OpenAI). */
    public static ObjectNode buscarClausulaRegulatoria(ObjectMapper mapper) {
        ObjectNode root = mapper.createObjectNode();
        root.put("type", "function");

        ObjectNode function = root.putObject("function");
        function.put("name", NOME_FERRAMENTA);
        function.put("description", DESCRICAO);

        ObjectNode parameters = function.putObject("parameters");
        parameters.put("type", "object");

        ObjectNode properties = parameters.putObject("properties");
        properties.putObject("tema").put("type", "string");

        ObjectNode jurisdicao = properties.putObject("jurisdicao");
        jurisdicao.put("type", "string");
        ArrayNode enumNode = jurisdicao.putArray("enum");
        enumNode.add("ANVISA");
        enumNode.add("FDA");

        ArrayNode required = parameters.putArray("required");
        required.add("tema");
        required.add("jurisdicao");

        return root;
    }
}
