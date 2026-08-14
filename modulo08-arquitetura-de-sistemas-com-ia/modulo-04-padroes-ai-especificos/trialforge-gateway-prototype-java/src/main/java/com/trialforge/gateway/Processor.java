package com.trialforge.gateway;

import java.io.PrintStream;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * O Gateway propriamente dito: os 4 grupos de padrao do Modulo 4 encadeados
 * numa requisicao so (RAG, Intent-Based Routing + Model Router, Semantic
 * Cache + Response Streaming, Confidence Threshold + Approval Gate + Audit
 * Trail).
 */
public class Processor {

    private final Embedder embedder;
    private final ChatStreamer chat;
    private final Map<String, IndicePreparado> preparados;
    private final SemanticCache cache;
    private final AuditTrail audit;
    private final Approver approval;
    private final PrintStream out;

    private final String modeloBarato;
    private final String modeloCaro;
    private final double limiarCache;
    private final double limiarConfianca;

    private int proximoId = 0;

    public Processor(Embedder embedder, ChatStreamer chat, Map<String, IndicePreparado> preparados,
            SemanticCache cache, AuditTrail audit, Approver approval, PrintStream out,
            String modeloBarato, String modeloCaro, double limiarCache, double limiarConfianca) {
        this.embedder = embedder;
        this.chat = chat;
        this.preparados = preparados;
        this.cache = cache;
        this.audit = audit;
        this.approval = approval;
        this.out = out;
        this.modeloBarato = modeloBarato;
        this.modeloCaro = modeloCaro;
        this.limiarCache = limiarCache;
        this.limiarConfianca = limiarConfianca;
    }

    private void log(String formato, Object... args) {
        if (out == null) {
            return;
        }
        out.println(String.format(formato, args));
    }

    // DemoLogger espera uma mensagem JA formatada (uma unica String) — usado
    // ao repassar o logger do Processor pro AgenticRag, que formata suas
    // proprias mensagens internamente antes de logar. Nao reusa log(formato,
    // args) aqui pra nao rodar String.format() DUAS vezes sobre um texto que
    // pode conter "%" (ex.: dentro do tema/texto de uma clausula).
    private void logLinhaPronta(String mensagemFormatada) {
        if (out != null) {
            out.println(mensagemFormatada);
        }
    }

    /**
     * Executa o pipeline completo pra uma pergunta: classifica a intencao,
     * consulta o Semantic Cache, roteia pro modelo e indice certos, busca a
     * clausula via RAG agentico, gera a resposta em streaming e, se
     * necessario, escala pro Approval Gate — registrando cada decisao na
     * trilha de auditoria. Devolve {@code null} quando o rascunho e
     * rejeitado no gate, igual ao {@code return null}/{@code return None}
     * dos originais.
     */
    public String processarRequisicao(String pergunta) throws Exception {
        proximoId++;
        String idRequisicao = "req-" + proximoId;
        log("\n[Gateway] Requisição recebida: \"%s\"", pergunta);

        String intencao = IntentClassifier.classificarIntencao(pergunta);
        log("[Intent-Based Routing] Intenção classificada: %s", intencao);

        List<Double> perguntaEmbedding = embedder.embedar(pergunta);

        // Semantic Cache: so entra em perguntas de rotina — sintese de CSR nunca usa cache (Modulo 4.3)
        if (!"sintese_csr".equals(intencao)) {
            SemanticCache.ResultadoConsulta resultadoCache = cache.consultar(perguntaEmbedding);
            if (resultadoCache.entrada() != null && resultadoCache.similaridade() >= limiarCache) {
                log("[Semantic Cache] HIT (similaridade %.3f) — resposta reaproveitada, sem chamar o modelo.",
                        resultadoCache.similaridade());
                audit.registrar(Map.of(
                        "id_requisicao", idRequisicao,
                        "pergunta", pergunta,
                        "intencao", intencao,
                        "cache_hit", true,
                        "similaridade_cache", resultadoCache.similaridade(),
                        "status_final", "respondido_via_cache"));
                return resultadoCache.entrada().resposta();
            }
            log("[Semantic Cache] MISS (melhor similaridade %.3f) — segue pro modelo.", resultadoCache.similaridade());
        }

        // Model Router (Modulo 4.2): a intencao decide qual modelo processa
        String modelo = "sintese_csr".equals(intencao) ? modeloCaro : modeloBarato;
        String tipoModelo = "sintese_csr".equals(intencao) ? "caro/capaz" : "barato/rápido";
        log("[Model Router] Modelo escolhido: %s (%s)", modelo, tipoModelo);

        // RAG (Modulo 4.1): Multi-Index roteia pro dominio certo, Hybrid Search busca dentro
        // dele (BM25 + embedding, fundidos por RRF), Agentic RAG insiste com estrategia mais
        // ampla se a confianca vier baixa.
        String indiceInicial = Indices.INDICE_POR_INTENCAO.get(intencao);
        log("[Multi-Index] Roteando pro índice \"%s\" (Módulo 4.1)", indiceInicial);
        ResultadoBusca resultadoRag = AgenticRag.buscarClausulaAgentica(
                preparados, pergunta, perguntaEmbedding, indiceInicial, limiarConfianca, this::logLinhaPronta);
        double confianca = resultadoRag.similaridadeCosseno();
        Clausula clausula = resultadoRag.clausula();
        log("[RAG] Cláusula final: \"%s\" — índice \"%s\", cosseno %.3f, BM25 %.3f, RRF %.4f, %d iteração(ões).",
                clausula.tema(), resultadoRag.indice(), confianca, resultadoRag.scoreBM25(), resultadoRag.scoreRRF(),
                resultadoRag.iteracoesUsadas());

        // Geracao com streaming (Modulo 4.3) — token a token, nao espera tudo pronto
        log("[Response Streaming] Gerando resposta:");
        out.print("    ");
        StringBuilder rascunho = new StringBuilder();
        List<ChatMessage> mensagens = List.of(
                new ChatMessage("system",
                        "Você redige respostas curtas e precisas sobre regras de estudos clínicos, "
                                + "citando a fonte regulatória fornecida."),
                new ChatMessage("user", String.format(
                        "Pergunta: %s\n\nCláusula regulatória relevante: %s\nFonte: %s\n\nResponda usando essa cláusula.",
                        pergunta, clausula.texto(), clausula.fonte())));
        chat.chatStream(modelo, mensagens, chunk -> {
            out.print(chunk);
            rascunho.append(chunk);
        });
        out.println();

        // Confidence Threshold + Approval Gate (Modulo 4.4)
        boolean precisaAprovacao = "sintese_csr".equals(intencao) || confianca < limiarConfianca;
        boolean aprovado = true;
        if (precisaAprovacao) {
            String motivo = "sintese_csr".equals(intencao)
                    ? "síntese de CSR — erro caro e irreversível, gate sempre obrigatório (Módulo 1.3)"
                    : String.format("confiança abaixo do limiar (%.3f < %s)", confianca, limiarConfianca);
            log("[Confidence Threshold] Escalando pro Approval Gate — motivo: %s", motivo);
            // Registro gravado ANTES do prompt, nao depois: o pedido pendente
            // precisa sobreviver independente de quando (ou se) alguem responde.
            Map<String, Object> pendencia = new LinkedHashMap<>();
            pendencia.put("id_requisicao", idRequisicao);
            pendencia.put("pergunta", pergunta);
            pendencia.put("intencao", intencao);
            pendencia.put("status_final", "aguardando_aprovacao");
            pendencia.put("motivo_gate", motivo);
            audit.registrar(pendencia);
            aprovado = approval.pedirAprovacaoHumana(rascunho.toString());
        } else {
            log("[Confidence Threshold] Confiança %.3f acima do limiar — segue sem Approval Gate.", confianca);
        }

        Map<String, Object> registroFinal = new LinkedHashMap<>();
        registroFinal.put("id_requisicao", idRequisicao);
        registroFinal.put("pergunta", pergunta);
        registroFinal.put("intencao", intencao);
        registroFinal.put("cache_hit", false);
        registroFinal.put("modelo_usado", modelo);
        registroFinal.put("indice_usado", resultadoRag.indice());
        registroFinal.put("iteracoes_agentic", resultadoRag.iteracoesUsadas());
        registroFinal.put("esgotou_agentic", resultadoRag.esgotouLimite());
        registroFinal.put("confianca_rag", confianca);
        registroFinal.put("gate_acionado", precisaAprovacao);
        registroFinal.put("aprovado", aprovado);
        registroFinal.put("status_final", aprovado ? "aprovado" : "rejeitado");
        audit.registrar(registroFinal);

        if (!aprovado) {
            log("[Approval Gate] Rascunho rejeitado — não vira oficial.");
            return null;
        }

        // Alimenta o Semantic Cache com essa pergunta+resposta pra proxima vez (so rotina)
        if (!"sintese_csr".equals(intencao)) {
            cache.adicionar(pergunta, perguntaEmbedding, rascunho.toString());
        }

        log("[Gateway] Requisição concluída — trilha de auditoria registrada.");
        return rascunho.toString();
    }
}
