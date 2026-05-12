package com.iadeva.analyzer.tool;

import com.fasterxml.jackson.databind.MappingIterator;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.csv.CsvMapper;
import com.fasterxml.jackson.dataformat.csv.CsvSchema;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.List;
import java.util.Map;

// Equivalente à tool customizada csvToJson.ts — processamento local sem precisar de MCP externo
@Component
public class CsvToJsonTool {

    private final CsvMapper csvMapper = new CsvMapper();
    private final ObjectMapper objectMapper = new ObjectMapper();

    /**
     * Converte dados CSV para JSON.
     * O CSV deve conter header na primeira linha.
     * Utiliza Jackson CSV para parsing seguro, incluindo valores com vírgulas entre aspas.
     *
     * @param csv String com conteúdo CSV (com header)
     * @return String JSON representando array de objetos
     */
    @Tool(description = "Converte dados em formato CSV para JSON. Use esta tool quando os dados de entrada estiverem em formato CSV com cabeçalho.")
    public String convert(String csv) {
        if (csv == null || csv.isBlank()) {
            return "[]";
        }

        try {
            // Schema com detecção automática de colunas a partir do header
            CsvSchema schema = CsvSchema.emptySchema().withHeader();

            MappingIterator<Map<String, String>> iterator = csvMapper
                    .readerFor(Map.class)
                    .with(schema)
                    .readValues(csv.trim());

            List<Map<String, String>> rows = iterator.readAll();
            return objectMapper.writeValueAsString(rows);

        } catch (IOException e) {
            throw new RuntimeException("Erro ao converter CSV para JSON: " + e.getMessage(), e);
        }
    }
}
