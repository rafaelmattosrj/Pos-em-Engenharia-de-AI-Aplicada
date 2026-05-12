package com.iadeva.agent.observability;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.iadeva.agent.model.AgentTrace;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * Serializa o trace do agente para JSON para análise post-mortem.
 * Equivalente ao trace_exporter.py — serializa trace para JSON para análise post-mortem.
 *
 * Permite exportar o AgentTrace como JSON formatado ou salvar em arquivo.
 */
@Component
public class TraceExporter {

    private static final Logger log = LoggerFactory.getLogger(TraceExporter.class);

    private final ObjectMapper objectMapper;

    public TraceExporter(ObjectMapper objectMapper) {
        // Configura pretty-print para legibilidade
        this.objectMapper = objectMapper.copy()
                .enable(SerializationFeature.INDENT_OUTPUT);
    }

    /**
     * Serializa o AgentTrace para uma String JSON formatada.
     *
     * @param trace Trace a ser serializado
     * @return JSON formatado (pretty-print) ou mensagem de erro
     */
    public String exportToJson(AgentTrace trace) {
        try {
            return objectMapper.writeValueAsString(trace);
        } catch (Exception e) {
            log.error("[TraceExporter] Falha ao serializar trace: {}", e.getMessage());
            return "{\"error\": \"Falha ao serializar trace: " + e.getMessage() + "\"}";
        }
    }

    /**
     * Salva o AgentTrace em um arquivo JSON no caminho especificado.
     *
     * @param trace    Trace a ser salvo
     * @param filename Nome do arquivo (ex: "trace_20240101_120000.json")
     */
    public void saveToFile(AgentTrace trace, String filename) {
        try {
            String json = exportToJson(trace);
            Path path = Paths.get(filename);

            // Cria diretórios pai se necessário
            if (path.getParent() != null) {
                Files.createDirectories(path.getParent());
            }

            Files.writeString(path, json);
            log.info("[TraceExporter] Trace salvo em: {}", path.toAbsolutePath());
        } catch (IOException e) {
            log.error("[TraceExporter] Falha ao salvar trace em '{}': {}", filename, e.getMessage());
        }
    }
}
