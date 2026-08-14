package com.trialforge.reactagent;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/**
 * Loop ReAct: Pensamento -> Acao -> Observacao -> Resposta Final. Criterio de
 * parada explicito no orquestrador, nunca deixado para o modelo.
 *
 * {@code maxIteracoes} tem default calibrado (4). O unico lugar que o sobrescreve e
 * o cenario de demonstracao do escalonamento em {@link Simulador#simularInteracao},
 * igual ao original.
 *
 * Retorna tambem uma "trilha": um trace de DESENVOLVIMENTO (uma entrada por volta
 * do loop, com o que o modelo decidiu e quanto tempo levou), pra inspecionar o
 * raciocinio enquanto se constroi — nao e observabilidade de producao, so pra debugar.
 *
 * Porte 1:1 de agenteICF em react-agent-prototype.js / react_agent_prototype.py.
 */
public final class AgenteICF {

    private AgenteICF() {
    }

    public static final int MAX_ITERACOES_PADRAO = 4;

    public record PassoTrilha(
            int volta,
            String acao,
            String ferramenta,
            java.util.Map<String, Object> argumentos,
            ExecutorFerramenta.ResultadoBusca observacao,
            long duracaoMs) {

        static PassoTrilha respostaFinal(int volta, long duracaoMs) {
            return new PassoTrilha(volta, "resposta_final", null, null, null, duracaoMs);
        }

        static PassoTrilha chamouFerramenta(int volta, String ferramenta, java.util.Map<String, Object> argumentos,
                ExecutorFerramenta.ResultadoBusca observacao, long duracaoMs) {
            return new PassoTrilha(volta, "chamou_ferramenta", ferramenta, argumentos, observacao, duracaoMs);
        }
    }

    public record ResultadoAgente(
            String rascunho,
            Integer iteracoes,
            boolean escalarParaAprovacaoHumana,
            String motivo,
            List<PassoTrilha> trilha) {

        static ResultadoAgente respostaFinal(String rascunho, int iteracoes, List<PassoTrilha> trilha) {
            return new ResultadoAgente(rascunho, iteracoes, false, null, trilha);
        }

        static ResultadoAgente escalado(String motivo, List<PassoTrilha> trilha) {
            return new ResultadoAgente(null, null, true, motivo, trilha);
        }
    }

    public static ResultadoAgente agenteICF(ChatClient client, String protocolo) throws IOException, InterruptedException {
        return agenteICF(client, protocolo, MAX_ITERACOES_PADRAO);
    }

    public static ResultadoAgente agenteICF(ChatClient client, String protocolo, int maxIteracoes)
            throws IOException, InterruptedException {
        ObjectMapper mapper = new ObjectMapper();

        List<ObjectNode> historico = new ArrayList<>();
        historico.add(mensagem(mapper, "system",
                "Você redige seções de ICF para estudos clínicos. Sempre que precisar de uma cláusula "
                        + "regulatória, chame a ferramenta em vez de perguntar ao usuário ou de escrever de "
                        + "memória. Nunca peça esclarecimento ao usuário: decida os parâmetros da ferramenta "
                        + "a partir do protocolo fornecido."));
        historico.add(mensagem(mapper, "user",
                "Protocolo do estudo: " + protocolo + "\n\n"
                        + "Jurisdição regulatória deste estudo: ANVISA.\n\n"
                        + "Gere a seção de assentimento do ICF (Termo de Consentimento) para este estudo, se "
                        + "aplicável. Use a ferramenta disponível para buscar a cláusula regulatória correta "
                        + "antes de escrever o texto."));

        List<ObjectNode> tools = List.of(ToolSchema.buscarClausulaRegulatoria(mapper));

        List<PassoTrilha> trilha = new ArrayList<>();

        for (int volta = 1; volta <= maxIteracoes; volta++) {
            long inicioVolta = System.currentTimeMillis();

            // Pensamento: o modelo decide, com base no historico acumulado, se ja sabe o suficiente
            ModeloResposta resposta = client.chat(historico, tools);
            long duracaoMs = System.currentTimeMillis() - inicioVolta;

            if (!resposta.temChamadaFerramenta()) {
                // Resposta Final: o modelo decidiu que ja tem o que precisa
                trilha.add(PassoTrilha.respostaFinal(volta, duracaoMs));
                return ResultadoAgente.respostaFinal(resposta.content(), volta, trilha);
            }

            // Acao: executa a ferramenta fora do modelo (codigo deterministico)
            ChamadaFerramenta chamada = resposta.chamadas().get(0);
            ExecutorFerramenta.ResultadoBusca observacao = ExecutorFerramenta.executarBuscaClausula(chamada.argumentos());
            trilha.add(PassoTrilha.chamouFerramenta(volta, chamada.nome(), chamada.argumentos(), observacao, duracaoMs));

            // Observacao: volta como novo contexto para o proximo Pensamento
            historico.add(resposta.mensagemBruta());
            ObjectNode toolMsg = mapper.createObjectNode();
            toolMsg.put("role", "tool");
            toolMsg.put("tool_name", chamada.nome());
            toolMsg.put("content", toJson(mapper, observacao));
            historico.add(toolMsg);
        }

        // Limite de iteracoes atingido sem convergencia: nunca falha silenciosamente
        return ResultadoAgente.escalado("não convergiu em " + maxIteracoes + " volta(s)", trilha);
    }

    private static ObjectNode mensagem(ObjectMapper mapper, String role, String content) {
        ObjectNode node = mapper.createObjectNode();
        node.put("role", role);
        node.put("content", content);
        return node;
    }

    private static String toJson(ObjectMapper mapper, ExecutorFerramenta.ResultadoBusca resultado) {
        try {
            return mapper.writeValueAsString(resultado);
        } catch (com.fasterxml.jackson.core.JsonProcessingException e) {
            throw new IllegalStateException("falha ao serializar observação da ferramenta", e);
        }
    }
}
