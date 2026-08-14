package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Cobre os mesmos três comportamentos demonstrados em simularInteracao() de
 * react-agent-prototype.js / react_agent_prototype.py: convergência normal (com
 * chamada de ferramenta), resposta final direta, e não-convergência com
 * escalonamento — usando um {@link FakeChatClient} no lugar do Ollama real.
 */
class AgenteICFTest {

    private final ObjectMapper mapper = new ObjectMapper();

    private ObjectNode mensagemComToolCall(String nomeFerramenta, Map<String, Object> argumentos) {
        ObjectNode mensagem = mapper.createObjectNode();
        mensagem.put("role", "assistant");
        mensagem.putNull("content");
        ArrayNode toolCalls = mensagem.putArray("tool_calls");
        ObjectNode toolCall = toolCalls.addObject();
        toolCall.put("id", "call-1");
        ObjectNode function = toolCall.putObject("function");
        function.put("name", nomeFerramenta);
        ObjectNode args = function.putObject("arguments");
        argumentos.forEach((k, v) -> args.put(k, String.valueOf(v)));
        return mensagem;
    }

    private ModeloResposta respostaComToolCall(String nomeFerramenta, Map<String, Object> argumentos) {
        ObjectNode mensagemBruta = mensagemComToolCall(nomeFerramenta, argumentos);
        return new ModeloResposta(mensagemBruta, null,
                List.of(new ChamadaFerramenta("call-1", nomeFerramenta, argumentos)));
    }

    private ModeloResposta respostaFinal(String texto) {
        ObjectNode mensagem = mapper.createObjectNode();
        mensagem.put("role", "assistant");
        mensagem.put("content", texto);
        return new ModeloResposta(mensagem, texto, List.of());
    }

    @Test
    void agenteICF_chamaFerramentaEDepoisRespondeFinal_convergeComTrilhaDeDuasVoltas() throws Exception {
        FakeChatClient fake = new FakeChatClient()
                .comResposta(respostaComToolCall(ToolSchema.NOME_FERRAMENTA,
                        Map.of("tema", "Assentimento para menores de idade em estudos clínicos", "jurisdicao", "ANVISA")))
                .comResposta(respostaFinal("Rascunho da seção de assentimento, com a cláusula ANVISA citada."));

        AgenteICF.ResultadoAgente resultado = AgenteICF.agenteICF(fake,
                "Estudo fase II, público-alvo entre 12 e 17 anos.");

        assertThat(resultado.escalarParaAprovacaoHumana()).isFalse();
        assertThat(resultado.iteracoes()).isEqualTo(2);
        assertThat(resultado.rascunho()).contains("assentimento");
        assertThat(resultado.trilha()).hasSize(2);
        assertThat(resultado.trilha().get(0).acao()).isEqualTo("chamou_ferramenta");
        assertThat(resultado.trilha().get(1).acao()).isEqualTo("resposta_final");
        assertThat(fake.chamadas).isEqualTo(2);
    }

    @Test
    void agenteICF_respondeDireto_convergeComUmaVoltaSoNaTrilha() throws Exception {
        FakeChatClient fake = new FakeChatClient().comResposta(respostaFinal("Não se aplica assentimento."));

        AgenteICF.ResultadoAgente resultado = AgenteICF.agenteICF(fake,
                "Estudo fase III, população adulta, sem envolvimento de menores de idade.");

        assertThat(resultado.escalarParaAprovacaoHumana()).isFalse();
        assertThat(resultado.iteracoes()).isEqualTo(1);
        assertThat(resultado.trilha()).hasSize(1);
    }

    @Test
    void agenteICF_naoConvergeNoLimiteDeIteracoes_escalaParaAprovacaoHumana() throws Exception {
        FakeChatClient fake = new FakeChatClient()
                .comResposta(respostaComToolCall(ToolSchema.NOME_FERRAMENTA,
                        Map.of("tema", "Assentimento para menores de idade", "jurisdicao", "ANVISA")));

        AgenteICF.ResultadoAgente resultado = AgenteICF.agenteICF(fake,
                "Estudo fase III, população adulta.", 1);

        assertThat(resultado.escalarParaAprovacaoHumana()).isTrue();
        assertThat(resultado.rascunho()).isNull();
        assertThat(resultado.motivo()).contains("não convergiu em 1 volta(s)");
        assertThat(resultado.trilha()).hasSize(1);
        assertThat(fake.chamadas).isEqualTo(1);
    }

    @Test
    void agenteICF_ferramentaNaoEncontraClausula_observacaoRefleteAviso() throws Exception {
        FakeChatClient fake = new FakeChatClient()
                .comResposta(respostaComToolCall(ToolSchema.NOME_FERRAMENTA,
                        Map.of("tema", "Consentimento informado de população adulta", "jurisdicao", "ANVISA")))
                .comResposta(respostaFinal("Cláusula de assentimento não se aplica a este estudo."));

        AgenteICF.ResultadoAgente resultado = AgenteICF.agenteICF(fake,
                "Estudo fase III, população adulta, diabetes tipo 2.");

        assertThat(resultado.trilha().get(0).observacao().texto()).isNull();
        assertThat(resultado.trilha().get(0).observacao().aviso()).contains("não encontrado");
    }
}
