package com.iadeva.analyzer;

import com.iadeva.analyzer.tool.CsvToJsonTool;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

// Testes unitários para CsvToJsonTool — cobertura dos casos principais de conversão
class CsvToJsonToolTest {

    private CsvToJsonTool tool;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        tool = new CsvToJsonTool();
        objectMapper = new ObjectMapper();
    }

    @Test
    @DisplayName("CSV com header e 2 linhas deve converter corretamente para array JSON")
    void csvComHeaderEDuasLinhasConverteCorretamente() throws Exception {
        // Dado: CSV com header e duas linhas de dados
        String csv = """
                produto,quantidade,preco
                Notebook,10,2500.00
                Mouse,50,45.00
                """;

        // Quando: converter para JSON
        String json = tool.convert(csv);

        // Então: resultado deve ser array com 2 objetos e campos corretos
        JsonNode array = objectMapper.readTree(json);
        assertThat(array.isArray()).isTrue();
        assertThat(array.size()).isEqualTo(2);

        JsonNode primeiroItem = array.get(0);
        assertThat(primeiroItem.get("produto").asText()).isEqualTo("Notebook");
        assertThat(primeiroItem.get("quantidade").asText()).isEqualTo("10");
        assertThat(primeiroItem.get("preco").asText()).isEqualTo("2500.00");

        JsonNode segundoItem = array.get(1);
        assertThat(segundoItem.get("produto").asText()).isEqualTo("Mouse");
        assertThat(segundoItem.get("quantidade").asText()).isEqualTo("50");
    }

    @Test
    @DisplayName("CSV vazio ou em branco deve retornar JSON array vazio")
    void csvVazioRetornaJsonArrayVazio() {
        // Cenário 1: string vazia
        assertThat(tool.convert("")).isEqualTo("[]");

        // Cenário 2: string nula
        assertThat(tool.convert(null)).isEqualTo("[]");

        // Cenário 3: string com apenas espaços
        assertThat(tool.convert("   ")).isEqualTo("[]");
    }

    @Test
    @DisplayName("CSV com valores contendo vírgulas entre aspas deve processar corretamente")
    void csvComValoresComVirgulasEntreAspasProcessaCorretamente() throws Exception {
        // Dado: CSV onde um campo contém vírgula dentro de aspas duplas
        String csv = """
                produto,descricao,preco
                Notebook,"Notebook Dell, 16GB RAM",2500.00
                Teclado,"Teclado mecânico, RGB",250.00
                """;

        // Quando: converter para JSON
        String json = tool.convert(csv);

        // Então: os valores com vírgulas devem ser preservados integralmente
        JsonNode array = objectMapper.readTree(json);
        assertThat(array.isArray()).isTrue();
        assertThat(array.size()).isEqualTo(2);

        JsonNode notebook = array.get(0);
        assertThat(notebook.get("produto").asText()).isEqualTo("Notebook");
        assertThat(notebook.get("descricao").asText()).isEqualTo("Notebook Dell, 16GB RAM");
        assertThat(notebook.get("preco").asText()).isEqualTo("2500.00");

        JsonNode teclado = array.get(1);
        assertThat(teclado.get("descricao").asText()).isEqualTo("Teclado mecânico, RGB");
    }
}
