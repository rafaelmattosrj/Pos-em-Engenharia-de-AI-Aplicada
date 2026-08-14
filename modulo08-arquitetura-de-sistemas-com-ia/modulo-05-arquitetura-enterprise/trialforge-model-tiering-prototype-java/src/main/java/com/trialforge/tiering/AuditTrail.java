package com.trialforge.tiering;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.time.Instant;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Trilha de auditoria: registro append-only em JSON Lines, um registro por
 * requisição processada — mesma disciplina do Módulo 4.5, relida depois para
 * confirmar as DECISÕES determinísticas (tier usado, se escalou, se bloqueou
 * por orçamento), nunca o texto exato gerado pelo modelo.
 *
 * Adaptação: cada porte (JS/Python original, Java, Go) grava em seu próprio
 * arquivo local {@code audit-trail-tiering.jsonl}, dentro da respectiva pasta
 * do projeto — não no arquivo compartilhado da pasta pai — para as execuções
 * de diferentes linguagens não colidirem/disputarem o mesmo arquivo. Ver README.
 */
public class AuditTrail {

    private final Path arquivo;
    private final ObjectMapper mapper = new ObjectMapper();
    private final Object lock = new Object();

    public AuditTrail(Path arquivo) {
        this.arquivo = arquivo;
    }

    public void registrar(Map<String, Object> registro) {
        Map<String, Object> linha = new LinkedHashMap<>();
        linha.put("timestamp", Instant.now().toString());
        linha.putAll(registro);
        synchronized (lock) {
            try {
                String json = mapper.writeValueAsString(linha) + System.lineSeparator();
                Files.writeString(arquivo, json, StandardCharsets.UTF_8,
                        StandardOpenOption.CREATE, StandardOpenOption.APPEND);
            } catch (IOException e) {
                throw new UncheckedIoAuditException("Falha ao gravar trilha de auditoria", e);
            }
        }
    }

    public List<JsonNode> lerTodas() throws IOException {
        if (!Files.exists(arquivo)) {
            return List.of();
        }
        List<JsonNode> resultado = new ArrayList<>();
        for (String linha : Files.readAllLines(arquivo, StandardCharsets.UTF_8)) {
            if (linha.isBlank()) continue;
            resultado.add(mapper.readTree(linha));
        }
        return resultado;
    }

    /** Erro não verificado (append em disco não deveria falhar num demo local). */
    public static class UncheckedIoAuditException extends RuntimeException {
        public UncheckedIoAuditException(String message, Throwable cause) {
            super(message, cause);
        }
    }
}
