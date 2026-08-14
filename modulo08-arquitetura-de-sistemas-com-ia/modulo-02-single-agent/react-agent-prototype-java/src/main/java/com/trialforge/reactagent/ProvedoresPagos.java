package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.List;
import java.util.Map;

/**
 * Três alternativas PAGAS ao Ollama local usado em {@link Main} / {@link OllamaReactClient}:
 * Claude (Anthropic), Gemini (Google) e GPT (OpenAI). A lição do diagrama de referência do
 * Módulo 1.2 vale aqui: o modelo é a peça que se troca — o loop ReAct ({@link AgenteICF}),
 * a ferramenta ({@link ExecutorFerramenta}) e o critério de parada não mudam em nenhuma das três.
 *
 * <p>Repare também que cada provedor declara o schema da MESMA ferramenta num formato
 * ligeiramente diferente: Claude e Gemini usam a mesma forma achatada (name/description
 * soltos na raiz do schema), o GPT usa uma forma aninhada, dentro de {@code "function": {...}}
 * (igual ao formato nativo do Ollama, usado em {@link ToolSchema}). É exatamente esse tipo
 * de fragmentação que um protocolo padronizado como o MCP existe para resolver.
 *
 * <p><b>IMPORTANTE — arquivo de referência, NÃO executado pelo fluxo principal ({@link Main}).</b>
 * Os três SDKs pagos (Anthropic, Google GenAI, OpenAI) não fazem parte das dependências
 * deste projeto — o {@code pom.xml} só declara {@code jackson-databind} (produção) e
 * JUnit/AssertJ (teste). Adicionar um SDK pago exigiria rede + credenciais pagas na hora do
 * build, o que quebraria {@code mvn compile} offline para quem só quer rodar a demo local
 * com Ollama. Por isso {@code chamarClaude}/{@code chamarGemini}/{@code chamarGpt} abaixo têm
 * corpo "de referência" — lançam {@link UnsupportedOperationException} se chamados direto, e
 * documentam, em Javadoc, a chamada real que o SDK faria. Os métodos {@code schemaClaude} /
 * {@code schemaGemini} / {@code schemaGpt} SÃO executáveis: constroem o schema real de cada
 * provedor com Jackson (sem depender de SDK nenhum), pra você comparar as três formas lado a
 * lado.
 *
 * <p>Para usar de verdade: adicione a dependência do SDK do provedor escolhido ao
 * {@code pom.xml}, configure a variável de ambiente da chave (ver {@code .env.example} e o
 * README.md deste projeto) e substitua o corpo do método pela chamada real documentada no
 * Javadoc de cada um.
 *
 * <p>Porte de referência de {@code provedores-pagos.js} / {@code provedores_pagos.py}.
 */
public final class ProvedoresPagos {

    private ProvedoresPagos() {
    }

    private static final String DESCRICAO_FERRAMENTA =
            "Busca cláusulas regulatórias de estudos clínicos por tema e jurisdição. Use quando "
                    + "precisar de texto normativo (ANVISA ou FDA) para compor uma seção do documento.";

    /** Resultado unificado das três chamadas: ou uma chamada de ferramenta, ou uma resposta final. */
    public record ChamadaResultado(String tipo, String nome, Map<String, Object> args, String texto) {
        static ChamadaResultado chamadaFerramenta(String nome, Map<String, Object> args) {
            return new ChamadaResultado("chamada_ferramenta", nome, args, null);
        }

        static ChamadaResultado respostaFinal(String texto) {
            return new ChamadaResultado("resposta_final", null, null, texto);
        }
    }

    // ---------- 1. Claude (Anthropic) — adicionar com.anthropic:anthropic-java ao pom.xml ----------

    /** Schema da ferramenta no formato do Claude: name/description soltos na raiz, {@code input_schema}. */
    public static ObjectNode schemaClaude(ObjectMapper mapper) {
        ObjectNode tool = mapper.createObjectNode();
        tool.put("name", ToolSchema.NOME_FERRAMENTA);
        tool.put("description", DESCRICAO_FERRAMENTA);
        ObjectNode inputSchema = tool.putObject("input_schema");
        inputSchema.put("type", "object");
        ObjectNode properties = inputSchema.putObject("properties");
        properties.putObject("tema").put("type", "string");
        ObjectNode jurisdicao = properties.putObject("jurisdicao");
        jurisdicao.put("type", "string");
        ArrayNode enumNode = jurisdicao.putArray("enum");
        enumNode.add("ANVISA");
        enumNode.add("FDA");
        ArrayNode required = inputSchema.putArray("required");
        required.add("tema");
        required.add("jurisdicao");
        return tool;
    }

    /**
     * Referência (não executada) — com a dependência real instalada, o corpo seria algo como:
     * <pre>{@code
     * AnthropicClient claude = AnthropicOkHttpClient.fromEnv(); // lê ANTHROPIC_API_KEY
     * Message resposta = claude.messages().create(MessageCreateParams.builder()
     *         .model("claude-sonnet-5")
     *         .maxTokens(1024)
     *         .addTool(schemaClaude(mapper))
     *         .messages(historico)
     *         .build());
     * Optional<ToolUseBlock> chamada = resposta.content().stream()
     *         .filter(ContentBlock::isToolUse).map(ContentBlock::asToolUse).findFirst();
     * return chamada.isPresent()
     *         ? ChamadaResultado.chamadaFerramenta(chamada.get().name(), chamada.get().input())
     *         : ChamadaResultado.respostaFinal(resposta.content().get(0).asText().text());
     * }</pre>
     */
    public static ChamadaResultado chamarClaude(List<ObjectNode> historico) {
        throw new UnsupportedOperationException(
                "Referência não executável: adicione com.anthropic:anthropic-java ao pom.xml e configure "
                        + "ANTHROPIC_API_KEY (ver .env.example e README.md).");
    }

    // ---------- 2. Gemini (Google) — adicionar com.google.genai:google-genai ao pom.xml ----------

    /** Schema da ferramenta no formato do Gemini: achatado como o Claude, mas com {@code type: "function"}. */
    public static ObjectNode schemaGemini(ObjectMapper mapper) {
        ObjectNode tool = mapper.createObjectNode();
        tool.put("type", "function");
        tool.put("name", ToolSchema.NOME_FERRAMENTA);
        tool.put("description", DESCRICAO_FERRAMENTA);
        ObjectNode parameters = tool.putObject("parameters");
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
        return tool;
    }

    /**
     * Nota de assinatura: {@code chamarClaude}/{@code chamarGpt} recebem {@code historico}
     * (o mesmo histórico de mensagens usado no loop ReAct); este método recebe só
     * {@code protocolo} (String) porque a API de Interactions do Gemini gerencia o histórico
     * de conversa do lado do servidor, não como uma lista que o chamador monta e reenvia a
     * cada turno — é uma diferença real de formato entre provedores, não um descuido. Ao
     * adaptar este sketch pro seu próprio protótipo, ajuste o chamador de acordo (não assuma
     * as três funções intercambiáveis por assinatura, só por papel na arquitetura).
     *
     * <p>Referência (não executada) — com a dependência real instalada, o corpo seria algo como:
     * <pre>{@code
     * Client gemini = Client.builder().build(); // lê GEMINI_API_KEY
     * Interaction interacao = gemini.interactions().create(InteractionCreateParams.builder()
     *         .model("gemini-3.5-flash")
     *         .input(protocolo)
     *         .addTool(schemaGemini(mapper))
     *         .build());
     * Optional<FunctionCallStep> chamada = interacao.steps().stream()
     *         .filter(Step::isFunctionCall).map(Step::asFunctionCall).findFirst();
     * return chamada.isPresent()
     *         ? ChamadaResultado.chamadaFerramenta(chamada.get().name(), chamada.get().arguments())
     *         : ChamadaResultado.respostaFinal(interacao.outputText());
     * }</pre>
     */
    public static ChamadaResultado chamarGemini(String protocolo) {
        throw new UnsupportedOperationException(
                "Referência não executável: adicione com.google.genai:google-genai ao pom.xml e configure "
                        + "GEMINI_API_KEY (ver .env.example e README.md).");
    }

    // ---------- 3. GPT (OpenAI) — adicionar com.openai:openai-java ao pom.xml ----------

    /** Schema da ferramenta no formato do GPT: forma aninhada, {@code type: "function"} + bloco {@code function}. */
    public static ObjectNode schemaGpt(ObjectMapper mapper) {
        // Mesmo formato aninhado já usado nativamente pelo Ollama — ver ToolSchema.
        return ToolSchema.buscarClausulaRegulatoria(mapper);
    }

    /**
     * Referência (não executada) — com a dependência real instalada, o corpo seria algo como:
     * <pre>{@code
     * OpenAIClient gpt = OpenAIOkHttpClient.fromEnv(); // lê OPENAI_API_KEY
     * ChatCompletion resposta = gpt.chat().completions().create(ChatCompletionCreateParams.builder()
     *         .model("gpt-5.6")
     *         .messages(historico)
     *         .addTool(schemaGpt(mapper))
     *         .build());
     * Optional<ChatCompletionMessageToolCall> chamada =
     *         resposta.choices().get(0).message().toolCalls().stream().findFirst();
     * return chamada.isPresent()
     *         ? ChamadaResultado.chamadaFerramenta(chamada.get().function().name(),
     *               parseArguments(chamada.get().function().arguments()))
     *         : ChamadaResultado.respostaFinal(resposta.choices().get(0).message().content().orElse(null));
     * }</pre>
     */
    public static ChamadaResultado chamarGpt(List<ObjectNode> historico) {
        throw new UnsupportedOperationException(
                "Referência não executável: adicione com.openai:openai-java ao pom.xml e configure "
                        + "OPENAI_API_KEY (ver .env.example e README.md).");
    }
}
