package com.iadeva.memory.memory;

// Memória longa — equivalente ao long_term_memory.py — fatos persistidos entre execuções

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.iadeva.memory.model.MemoryFact;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Memória de longo prazo: fatos persistidos em disco entre execuções do agente.
 * Garante que conhecimento acumulado não seja perdido ao reiniciar a aplicação.
 * Não duplica fatos com mesmo conteúdo — deduplicação por campo content.
 */
@Component
public class LongTermMemory {

    private final String memoryPath;
    private final ObjectMapper mapper;
    private final File storageFile;

    // Lista em memória sincronizada com o arquivo JSON
    private List<MemoryFact> facts = new ArrayList<>();

    public LongTermMemory(@Value("${app.memory.path:./data}") String memoryPath) {
        this.memoryPath = memoryPath;
        this.mapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        this.storageFile = new File(memoryPath + "/long-term.json");
    }

    @PostConstruct
    public void init() {
        // Cria diretório se não existir
        storageFile.getParentFile().mkdirs();
        if (storageFile.exists()) {
            try {
                facts = mapper.readValue(storageFile, new TypeReference<List<MemoryFact>>() {});
            } catch (IOException e) {
                facts = new ArrayList<>();
            }
        }
    }

    /**
     * Adiciona fato à memória. Ignora silenciosamente se conteúdo já existir (deduplicação).
     */
    public synchronized void addFact(MemoryFact fact) {
        boolean duplicado = facts.stream()
                .anyMatch(f -> f.content().equalsIgnoreCase(fact.content()));
        if (duplicado) return;

        facts.add(fact);
        persist();
    }

    /**
     * Retorna todos os fatos não expirados.
     */
    public List<MemoryFact> getFacts() {
        var agora = java.time.Instant.now();
        return facts.stream()
                .filter(f -> f.expiresAt() == null || f.expiresAt().isAfter(agora))
                .collect(Collectors.toList());
    }

    /**
     * Filtra fatos por fonte (ex: "usuario", "api_externa").
     */
    public List<MemoryFact> getFactsBySource(String source) {
        return getFacts().stream()
                .filter(f -> source.equalsIgnoreCase(f.source()))
                .collect(Collectors.toList());
    }

    /**
     * Remove fato pelo id.
     */
    public synchronized void removeFact(String id) {
        facts.removeIf(f -> f.id().equals(id));
        persist();
    }

    // Persiste o estado atual no arquivo JSON
    private void persist() {
        try {
            mapper.writeValue(storageFile, facts);
        } catch (IOException e) {
            throw new RuntimeException("Falha ao persistir memória de longo prazo: " + e.getMessage(), e);
        }
    }
}
