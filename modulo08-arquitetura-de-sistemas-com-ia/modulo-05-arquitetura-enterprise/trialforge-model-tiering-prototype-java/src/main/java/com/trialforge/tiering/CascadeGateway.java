package com.trialforge.tiering;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Cascata de Model Tiering (Módulo 5.4): tenta o tier mais barato primeiro,
 * escala pro tier caro se a confiança de BUSCA (achou a cláusula certa?) OU a
 * confiança de RESPOSTA (g(pergunta, resposta) — a resposta gerada ficou fiel
 * à cláusula que recebeu?) ficar abaixo do limiar. Síntese de CSR é regra
 * fixa: pula direto pro tier caro e aciona o Approval Gate.
 *
 * Porte 1:1 de processarComCascata em trialforge-model-tiering-prototype.js / .py.
 */
public class CascadeGateway {

    public static final String MODELO_TIER1 = "gemma4:e2b";      // barato — tentado primeiro, sempre
    public static final String MODELO_TIER2 = "gemma4:latest";   // caro — escalação ou regra fixa (CSR)
    public static final String MODELO_EMBEDDING = "nomic-embed-text";

    public static final double LIMIAR_CASCATA_BUSCA = 0.75;
    public static final double LIMIAR_CASCATA_RESPOSTA = 0.75;

    // Custo estimado por chamada, só pra demonstrar o controle de orçamento — valores
    // ilustrativos, não preço real de nenhum provedor.
    public static final double CUSTO_TIER1 = 0.001;
    public static final double CUSTO_TIER2 = 0.01;

    private final OllamaGateway ollama;
    private final RagIndex ragIndex;
    private final OrcamentoManager orcamento;
    private final ApprovalPrompt approvalPrompt;
    private final AuditTrail auditTrail;

    public CascadeGateway(OllamaGateway ollama, RagIndex ragIndex, OrcamentoManager orcamento,
                           ApprovalPrompt approvalPrompt, AuditTrail auditTrail) {
        this.ollama = ollama;
        this.ragIndex = ragIndex;
        this.orcamento = orcamento;
        this.approvalPrompt = approvalPrompt;
        this.auditTrail = auditTrail;
    }

    public String processarComCascata(String pergunta, String estudoId) throws IOException, InterruptedException {
        System.out.printf("%n[Gateway] Estudo \"%s\" — requisição: \"%s\"%n", estudoId, pergunta);

        String intencao = IntentClassifier.classificarIntencao(pergunta);
        double custoMaximoPossivel = intencao.equals(IntentClassifier.SINTESE_CSR) ? CUSTO_TIER2 : CUSTO_TIER1 + CUSTO_TIER2;

        // Orçamento reservado (não só checado) ANTES de qualquer chamada de modelo.
        if (!orcamento.reservarOrcamento(estudoId, custoMaximoPossivel)) {
            System.out.printf("[Orçamento] BLOQUEADO — estudo \"%s\" ultrapassaria o limite antes mesmo de chamar o modelo.%n", estudoId);
            Map<String, Object> registro = new LinkedHashMap<>();
            registro.put("estudoId", estudoId);
            registro.put("pergunta", pergunta);
            registro.put("status_final", "bloqueado_por_orcamento");
            auditTrail.registrar(registro);
            return null;
        }
        double custoRealUsado = 0;

        double[] perguntaEmbedding = ollama.embed(MODELO_EMBEDDING, pergunta);
        RagIndex.ResultadoBusca resultadoBusca = ragIndex.buscarClausula(perguntaEmbedding);
        double confiancaBusca = resultadoBusca.similaridade();
        Clausula clausula = resultadoBusca.clausula();
        int indiceClausula = resultadoBusca.indice();

        String tierUsado;
        String rascunho;
        boolean escalou = false;
        Double confiancaResposta = null;

        if (intencao.equals(IntentClassifier.SINTESE_CSR)) {
            // Regra fixa (Módulo 1.3): erro caro e irreversível, pula direto pro tier caro.
            System.out.println("[Model Tiering] Síntese de CSR — regra fixa, direto pro Tier 2 (caro).");
            tierUsado = "Tier 2";
            rascunho = gerarComTier(MODELO_TIER2, pergunta, clausula);
            custoRealUsado += CUSTO_TIER2;
        } else {
            System.out.printf("[Model Tiering] Tentando Tier 1 (barato) primeiro — cláusula encontrada com confiança de busca %.3f%n", confiancaBusca);
            tierUsado = "Tier 1";
            rascunho = gerarComTier(MODELO_TIER1, pergunta, clausula);
            custoRealUsado += CUSTO_TIER1;

            confiancaResposta = ragIndex.calcularConfiancaResposta(rascunho, indiceClausula);
            System.out.printf("[Model Tiering] Confiança da resposta do Tier 1 — g(pergunta, resposta): %.3f%n", confiancaResposta);

            boolean buscaFalhou = confiancaBusca < LIMIAR_CASCATA_BUSCA;
            boolean respostaFalhou = confiancaResposta < LIMIAR_CASCATA_RESPOSTA;

            if (buscaFalhou || respostaFalhou) {
                String motivo = buscaFalhou && respostaFalhou ? "busca e resposta"
                        : buscaFalhou ? "confiança de busca" : "confiança de resposta";
                System.out.printf("[Model Tiering] Escalando pro Tier 2 — motivo: %s abaixo do limiar (busca %.3f, resposta %.3f).%n",
                        motivo, confiancaBusca, confiancaResposta);
                escalou = true;
                tierUsado = "Tier 2 (escalado)";
                rascunho = gerarComTier(MODELO_TIER2, pergunta, clausula);
                custoRealUsado += CUSTO_TIER2;
                confiancaResposta = ragIndex.calcularConfiancaResposta(rascunho, indiceClausula);
            } else {
                System.out.println("[Model Tiering] Busca e resposta acima do limiar — Tier 1 resolve, sem escalar.");
            }
        }

        // A reserva cobriu o pior caso; devolve o que sobrou se o custo real ficou menor.
        orcamento.liberarSobra(estudoId, custoMaximoPossivel - custoRealUsado);

        // Estreitamento de escopo deliberado em relação ao Módulo 4.5: aqui só CSR aciona
        // o Approval Gate — uma escalação de cascata por confiança baixa só ajusta o tier.
        boolean precisaAprovacao = intencao.equals(IntentClassifier.SINTESE_CSR);
        boolean aprovado = true;
        if (precisaAprovacao) {
            aprovado = approvalPrompt.approve(rascunho);
        }

        Map<String, Object> registro = new LinkedHashMap<>();
        registro.put("estudoId", estudoId);
        registro.put("pergunta", pergunta);
        registro.put("intencao", intencao);
        registro.put("tier_usado", tierUsado);
        registro.put("escalou_cascata", escalou);
        registro.put("confianca_busca", confiancaBusca);
        registro.put("confianca_resposta", confiancaResposta);
        registro.put("gasto_acumulado", orcamento.gastoAtual(estudoId));
        registro.put("orcamento_limite", orcamento.limite(estudoId));
        registro.put("aprovado", aprovado);
        registro.put("status_final", aprovado ? "aprovado" : "rejeitado");
        auditTrail.registrar(registro);

        System.out.printf("[Orçamento] Estudo \"%s\": gasto acumulado %.4f / limite %s%n",
                estudoId, orcamento.gastoAtual(estudoId), orcamento.limite(estudoId));
        return rascunho;
    }

    private String gerarComTier(String modelo, String pergunta, Clausula clausula) throws IOException, InterruptedException {
        String system = "Você redige respostas curtas e precisas sobre regras de estudos clínicos, "
                + "citando a fonte regulatória fornecida.";
        String user = "Pergunta: " + pergunta + "\n\nCláusula regulatória relevante: " + clausula.texto()
                + "\nFonte: " + clausula.fonte() + "\n\nResponda usando essa cláusula.";
        String rascunho = ollama.chatStream(modelo, system, user, System.out::print);
        System.out.println();
        return rascunho;
    }
}
