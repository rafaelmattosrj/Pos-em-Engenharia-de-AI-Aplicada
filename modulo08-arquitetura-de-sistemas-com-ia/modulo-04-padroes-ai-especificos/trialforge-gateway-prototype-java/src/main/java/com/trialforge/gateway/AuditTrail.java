package com.trialforge.gateway;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.BufferedReader;
import java.io.FileOutputStream;
import java.io.FileReader;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.Writer;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Audit Trail (Modulo 4.4): trilha de auditoria append-only — nunca
 * sobrescreve (21 CFR Part 11) — cada registro e uma linha JSON acrescentada
 * ao arquivo.
 */
public class AuditTrail {

    private final String caminho;
    private final ObjectMapper mapper = new ObjectMapper();

    public AuditTrail(String caminho) {
        this.caminho = caminho;
    }

    /** Acrescenta um registro (com timestamp automatico) ao arquivo de
     * auditoria. As chaves usam snake_case pra bater com o formato dos dois
     * originais (id_requisicao, cache_hit, gate_acionado, etc.). */
    public void registrar(Map<String, Object> registro) throws IOException {
        Map<String, Object> completo = new LinkedHashMap<>();
        completo.put("timestamp", Instant.now().toString());
        completo.putAll(registro);

        String linha = mapper.writeValueAsString(completo);
        try (Writer writer = new OutputStreamWriter(new FileOutputStream(caminho, true), StandardCharsets.UTF_8)) {
            writer.write(linha);
            writer.write("\n");
        }
    }

    private List<Map<String, Object>> lerLinhas() throws IOException {
        List<Map<String, Object>> linhas = new ArrayList<>();
        try (BufferedReader reader = new BufferedReader(new FileReader(caminho, StandardCharsets.UTF_8))) {
            String linha;
            while ((linha = reader.readLine()) != null) {
                if (linha.isBlank()) {
                    continue;
                }
                @SuppressWarnings("unchecked")
                Map<String, Object> registro = mapper.readValue(linha, Map.class);
                linhas.add(registro);
            }
        }
        return linhas;
    }

    private record Checagem(String descricao, boolean ok) {
    }

    /**
     * Rele o audit-trail.jsonl e confere que as ULTIMAS 5 requisicoes
     * CONCLUIDAS (ignorando pendencias "aguardando_aprovacao") tomaram as
     * decisoes deterministicas esperadas pelo roteiro de demo — cache
     * hit/miss, modelo escolhido, indice roteado, gate acionado ou nao.
     * Nunca compara o texto exato gerado pelo modelo, que varia entre
     * execucoes. Nao confia so no humano assistindo notar os comportamentos.
     */
    public void verificarTrilhaAuditoria(String modeloCaro, double limiarConfianca, DemoLogger logger) throws IOException {
        List<Map<String, Object>> todasLinhas = lerLinhas();

        List<Map<String, Object>> linhasConcluidas = new ArrayList<>();
        for (Map<String, Object> l : todasLinhas) {
            if (!"aguardando_aprovacao".equals(l.get("status_final"))) {
                linhasConcluidas.add(l);
            }
        }

        int inicio = Math.max(0, linhasConcluidas.size() - 5);
        List<Map<String, Object>> linhas = linhasConcluidas.subList(inicio, linhasConcluidas.size());

        if (logger != null) {
            logger.log(String.format(
                    "\n== Verificação: últimas 5 requisições concluídas (%d concluídas, %d entradas no total incluindo pendências) ==",
                    linhasConcluidas.size(), todasLinhas.size()));
        }

        boolean algumaPendencia = todasLinhas.stream().anyMatch(l -> "aguardando_aprovacao".equals(l.get("status_final")));

        List<Checagem> checagens = new ArrayList<>();
        checagens.add(new Checagem("pelo menos 5 requisições concluídas na trilha", linhasConcluidas.size() >= 5));
        checagens.add(new Checagem(
                "ao menos 1 registro \"aguardando_aprovacao\" gravado antes de qualquer decisão", algumaPendencia));
        checagens.add(new Checagem("#1 rotina: sem cache hit", Boolean.FALSE.equals(campo(linhas, 0, "cache_hit"))));
        checagens.add(new Checagem("#1 rotina: confiança alta, sem Approval Gate",
                Boolean.FALSE.equals(campo(linhas, 0, "gate_acionado"))));
        checagens.add(new Checagem("#2 paráfrase: cache HIT",
                Boolean.TRUE.equals(campo(linhas, 1, "cache_hit"))
                        && "respondido_via_cache".equals(campo(linhas, 1, "status_final"))));
        checagens.add(new Checagem("#3 síntese de CSR: modelo caro", modeloCaro.equals(campo(linhas, 2, "modelo_usado"))));
        checagens.add(new Checagem("#3 síntese de CSR: Approval Gate sempre acionado",
                Boolean.TRUE.equals(campo(linhas, 2, "gate_acionado"))));
        checagens.add(new Checagem("#3 síntese de CSR: Multi-Index roteou pro índice csr",
                "csr".equals(campo(linhas, 2, "indice_usado"))));
        checagens.add(new Checagem("#4 tema diferente: cache miss", Boolean.FALSE.equals(campo(linhas, 3, "cache_hit"))));
        checagens.add(new Checagem("#4 tema diferente: Agentic RAG esgotou as 3 iterações",
                numeroIgual(campo(linhas, 3, "iteracoes_agentic"), AgenticRag.MAX_ITERACOES_AGENTIC)
                        && Boolean.TRUE.equals(campo(linhas, 3, "esgotou_agentic"))));
        checagens.add(new Checagem("#4 tema diferente: Approval Gate por confiança baixa",
                Boolean.TRUE.equals(campo(linhas, 3, "gate_acionado")) && menorQue(campo(linhas, 3, "confianca_rag"), limiarConfianca)));
        checagens.add(new Checagem("#5 critério de protocolo: Multi-Index roteou pro índice protocolo",
                "protocolo".equals(campo(linhas, 4, "indice_usado"))));
        checagens.add(new Checagem("#5 critério de protocolo: Agentic RAG confiante já na 1ª iteração",
                numeroIgual(campo(linhas, 4, "iteracoes_agentic"), 1)));
        checagens.add(new Checagem("#5 critério de protocolo: sem Approval Gate",
                Boolean.FALSE.equals(campo(linhas, 4, "gate_acionado"))));

        int passou = 0;
        for (Checagem c : checagens) {
            if (logger != null) {
                logger.log(String.format("  [%s] %s", c.ok() ? "OK" : "FALHOU", c.descricao()));
            }
            if (c.ok()) {
                passou++;
            }
        }
        if (logger != null) {
            logger.log(String.format("Total: %d verificação(ões), %d passou(passaram), %d falhou(falharam).",
                    checagens.size(), passou, checagens.size() - passou));
        }

        if (passou != checagens.size()) {
            throw new IllegalStateException(String.format(
                    "A trilha de auditoria não confirma os 4 caminhos esperados (%d/%d) — reveja audit-trail.jsonl.",
                    passou, checagens.size()));
        }
    }

    private static Object campo(List<Map<String, Object>> linhas, int indice, String chave) {
        if (indice < 0 || indice >= linhas.size()) {
            return null;
        }
        return linhas.get(indice).get(chave);
    }

    private static boolean numeroIgual(Object valor, int esperado) {
        return valor instanceof Number numero && numero.intValue() == esperado;
    }

    private static boolean menorQue(Object valor, double limite) {
        return valor instanceof Number numero && numero.doubleValue() < limite;
    }
}
