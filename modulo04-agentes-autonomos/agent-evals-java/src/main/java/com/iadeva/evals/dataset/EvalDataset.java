package com.iadeva.evals.dataset;

// Equivalente ao eval_dataset.py — carrega e gerencia o dataset de avaliação

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.evals.model.EvalScenario;
import jakarta.annotation.PostConstruct;
import org.springframework.core.io.ClassPathResource;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.List;

/**
 * Componente responsável por carregar os cenários de avaliação do arquivo JSON.
 * O dataset fica em src/main/resources/eval-dataset.json e é lido uma única vez na inicialização.
 */
@Component
public class EvalDataset {

    private final ObjectMapper objectMapper;
    private List<EvalScenario> scenarios;

    public EvalDataset(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    /**
     * Carrega o dataset na inicialização do contexto Spring.
     * Falha imediatamente se o arquivo não existir ou estiver malformado.
     */
    @PostConstruct
    public void load() {
        try {
            var resource = new ClassPathResource("eval-dataset.json");
            scenarios = objectMapper.readValue(
                    resource.getInputStream(),
                    new TypeReference<List<EvalScenario>>() {}
            );
        } catch (IOException e) {
            throw new IllegalStateException("Falha ao carregar eval-dataset.json: " + e.getMessage(), e);
        }
    }

    /**
     * Retorna todos os cenários do dataset.
     *
     * @return lista imutável de cenários
     */
    public List<EvalScenario> getScenarios() {
        return List.copyOf(scenarios);
    }
}
