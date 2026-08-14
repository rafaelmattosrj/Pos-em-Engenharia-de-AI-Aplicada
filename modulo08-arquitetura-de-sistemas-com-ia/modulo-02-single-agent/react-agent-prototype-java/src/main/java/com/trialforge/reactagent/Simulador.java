package com.trialforge.reactagent;

import java.io.IOException;
import java.util.List;

/**
 * Simula tres interacoes reais entre a Mariana (usuaria) e o agente, cobrindo os tres
 * comportamentos que a Missao Pratica #02 pede pra demonstrar no proprio prototipo:
 * convergencia normal, ferramenta sem resultado, e nao-convergencia com escalonamento.
 *
 * Porte 1:1 de simularInteracao / imprimirTrilha em react-agent-prototype.js / .py.
 */
public final class Simulador {

    private Simulador() {
    }

    public static void imprimirTrilha(List<AgenteICF.PassoTrilha> trilha) {
        System.out.println("  [Trilha de desenvolvimento — não é a trilha de auditoria do Módulo 5.2]");
        for (AgenteICF.PassoTrilha passo : trilha) {
            if ("chamou_ferramenta".equals(passo.acao())) {
                System.out.println("    volta " + passo.volta() + " (" + passo.duracaoMs() + "ms): chamou "
                        + passo.ferramenta() + "(" + passo.argumentos() + ") -> " + passo.observacao());
            } else {
                System.out.println("    volta " + passo.volta() + " (" + passo.duracaoMs()
                        + "ms): decidiu que já tinha o suficiente, respondeu");
            }
        }
    }

    public static void simularInteracao(ChatClient client) throws IOException, InterruptedException {
        // ---------- Cenário 1: convergência normal, a cláusula existe ----------
        System.out.println("===== Cenário 1: convergência normal (cláusula encontrada) =====");
        String protocolo1 = "Estudo fase II, público-alvo entre 12 e 17 anos, terapia oncológica experimental.";
        System.out.println("[Mariana submete o protocolo ao Gateway]");
        System.out.println("   " + protocolo1);
        System.out.println();

        AgenteICF.ResultadoAgente resultado1 = AgenteICF.agenteICF(client, protocolo1);
        System.out.println("[Agente ICF] Rascunho gerado após " + resultado1.iteracoes()
                + " volta(s) de loop, rodando localmente:");
        System.out.println("   " + resultado1.rascunho());
        imprimirTrilha(resultado1.trilha());
        System.out.println("[Sistema] Rascunho aguardando revisão do Approval Gate antes de virar versão oficial.");
        System.out.println();

        // ---------- Cenário 2: ferramenta não encontra cláusula ----------
        System.out.println("===== Cenário 2: ferramenta não encontra cláusula (população adulta) =====");
        String protocolo2 = "Estudo fase III, população adulta (18-65 anos), diabetes tipo 2, sem envolvimento "
                + "de menores de idade.";
        System.out.println("[Mariana submete o protocolo ao Gateway]");
        System.out.println("   " + protocolo2);
        System.out.println();

        AgenteICF.ResultadoAgente resultado2 = AgenteICF.agenteICF(client, protocolo2);
        imprimirTrilha(resultado2.trilha());
        if (resultado2.escalarParaAprovacaoHumana()) {
            System.out.println("[Orquestrador] Loop não convergiu — " + resultado2.motivo());
            System.out.println("[Sistema] Encaminhando ao Approval Gate para intervenção manual.");
        } else {
            System.out.println("[Agente ICF] Respondeu após " + resultado2.iteracoes()
                    + " volta(s), sem achar cláusula específica.");
            System.out.println(
                    "[Atenção] Repare na trilha: mesmo sem achar a cláusula, o modelo escreveu um rascunho "
                            + "genérico em vez de escalar — desviando da instrução do system prompt. É "
                            + "exatamente esse tipo de desvio que a trilha existe para flagrar, e por isso o "
                            + "Approval Gate revisa antes de qualquer coisa virar oficial.");
        }
        System.out.println();

        // ---------- Cenário 3: não convergência, escalonamento ----------
        // maxIteracoes forçado a 1 só aqui: testado com o modelo real, mesmo sem achar a
        // cláusula, tende a responder algo em vez de insistir por várias voltas (ver o
        // desvio do Cenário 2) — comportamento de LLM não é 100% previsível. Forçar o
        // limiar garante que este cenário sempre demonstre o mecanismo de escalonamento
        // de forma confiável. Em produção, o limiar calibrado continua sendo
        // MAX_ITERACOES_PADRAO = 4, não 1.
        System.out.println(
                "===== Cenário 3: não convergência, escalonamento (limiar forçado a 1 volta pra demo confiável) =====");
        String protocolo3 = "Estudo fase III, população adulta, diabetes tipo 2, sem envolvimento de menores de idade.";
        System.out.println("[Mariana submete o protocolo ao Gateway]");
        System.out.println("   " + protocolo3);
        System.out.println();

        AgenteICF.ResultadoAgente resultado3 = AgenteICF.agenteICF(client, protocolo3, 1);
        imprimirTrilha(resultado3.trilha());
        if (resultado3.escalarParaAprovacaoHumana()) {
            System.out.println("[Orquestrador] Loop não convergiu — " + resultado3.motivo());
            System.out.println(
                    "[Sistema] Encaminhando ao Approval Gate para intervenção manual — nunca falha silenciosamente.");
        } else {
            System.out.println("[Atenção] O modelo convergiu na única volta permitida antes mesmo de precisar escalar.");
        }
    }
}
