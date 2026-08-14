package com.trialforge.tiering;

import com.fasterxml.jackson.databind.JsonNode;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/**
 * Verificação pós-execução: relê a trilha de auditoria e confere as DECISÕES
 * determinísticas (tier usado, se escalou, se bloqueou por orçamento), nunca
 * o texto exato gerado pelo modelo. Olha só as últimas 4 entradas — a trilha
 * é append-only, nunca apagada entre execuções.
 *
 * Porte 1:1 de verificarTrilhaAuditoria em trialforge-model-tiering-prototype.js / .py.
 */
public final class AuditTrailVerifier {

    private AuditTrailVerifier() {
    }

    public record Checagem(String descricao, boolean ok) {
    }

    public static List<Checagem> verificar(List<JsonNode> todasLinhas) {
        List<JsonNode> linhas = todasLinhas.size() <= 4
                ? todasLinhas
                : todasLinhas.subList(todasLinhas.size() - 4, todasLinhas.size());

        List<Checagem> checagens = new ArrayList<>();
        checagens.add(new Checagem("pelo menos 4 entradas na trilha", todasLinhas.size() >= 4));
        checagens.add(new Checagem("#1 estudo-A rotina: Tier 1 resolve, sem escalar",
                campoTexto(linhas, 0, "tier_usado").equals("Tier 1") && !campoBool(linhas, 0, "escalou_cascata")));
        checagens.add(new Checagem("#2 estudo-A tema diferente: escala pro Tier 2",
                campoBool(linhas, 1, "escalou_cascata") && campoTexto(linhas, 1, "tier_usado").equals("Tier 2 (escalado)")));
        checagens.add(new Checagem("#3 estudo-A síntese de CSR: regra fixa pro Tier 2, sem cascata",
                campoTexto(linhas, 2, "tier_usado").equals("Tier 2") && !campoBool(linhas, 2, "escalou_cascata")));
        checagens.add(new Checagem("#3 síntese de CSR: aprovado no Approval Gate", campoBool(linhas, 2, "aprovado")));
        checagens.add(new Checagem("#4 estudo-B: bloqueado por orçamento antes de chamar o modelo",
                campoTexto(linhas, 3, "status_final").equals("bloqueado_por_orcamento")));
        checagens.add(new Checagem(
                "#1 e #2: escalação decidida pela confiança da RESPOSTA (g(pergunta,resposta)), não só da busca",
                campo(linhas, 0, "confianca_resposta").isNumber() && campo(linhas, 1, "confianca_resposta").isNumber()));
        return checagens;
    }

    /** Executa a verificação, imprime cada checagem e lança se alguma falhar. */
    public static void verificarEImprimir(AuditTrail auditTrail) throws IOException {
        List<JsonNode> todasLinhas = auditTrail.lerTodas();
        System.out.printf("%n== Verificação: últimas 4 entradas da trilha de auditoria (%d no total) ==%n", todasLinhas.size());

        List<Checagem> checagens = verificar(todasLinhas);
        long passou = checagens.stream().filter(Checagem::ok).count();
        for (Checagem c : checagens) {
            System.out.printf("  [%s] %s%n", c.ok() ? "OK" : "FALHOU", c.descricao());
        }
        System.out.printf("Total: %d verificação(ões), %d passou(passaram), %d falhou(falharam).%n",
                checagens.size(), passou, checagens.size() - passou);

        if (passou != checagens.size()) {
            throw new IllegalStateException(String.format(
                    "A trilha de auditoria não confirma os 3 comportamentos exigidos (%d/%d) — reveja audit-trail-tiering.jsonl.",
                    passou, checagens.size()));
        }
    }

    private static JsonNode campo(List<JsonNode> linhas, int indice, String chave) {
        if (indice >= linhas.size()) {
            return com.fasterxml.jackson.databind.node.MissingNode.getInstance();
        }
        return linhas.get(indice).path(chave);
    }

    private static String campoTexto(List<JsonNode> linhas, int indice, String chave) {
        return campo(linhas, indice, chave).asText("");
    }

    private static boolean campoBool(List<JsonNode> linhas, int indice, String chave) {
        return campo(linhas, indice, chave).asBoolean(false);
    }
}
