package com.iadeva.memory.memory;

// Memória episódica — equivalente ao episodic_memory.py — histórico de execuções passadas

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.iadeva.memory.model.Episode;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Memória episódica: registra cada execução completa do agente como um episódio.
 * Permite que o agente aprenda com o histórico de interações passadas.
 */
@Component
public class EpisodicMemory {

    private final ObjectMapper mapper;
    private final File storageFile;

    // Lista em memória sincronizada com o arquivo JSON
    private List<Episode> episodes = new ArrayList<>();

    public EpisodicMemory(@Value("${app.memory.path:./data}") String memoryPath) {
        this.mapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        this.storageFile = new File(memoryPath + "/episodes.json");
    }

    @PostConstruct
    public void init() {
        storageFile.getParentFile().mkdirs();
        if (storageFile.exists()) {
            try {
                episodes = mapper.readValue(storageFile, new TypeReference<List<Episode>>() {});
            } catch (IOException e) {
                episodes = new ArrayList<>();
            }
        }
    }

    /**
     * Registra um novo episódio de execução.
     */
    public synchronized void addEpisode(Episode episode) {
        episodes.add(episode);
        persist();
    }

    /**
     * Retorna todos os episódios ordenados do mais recente para o mais antigo.
     */
    public List<Episode> getEpisodes() {
        return episodes.stream()
                .sorted(Comparator.comparing(Episode::timestamp).reversed())
                .collect(Collectors.toList());
    }

    /**
     * Retorna os N episódios mais recentes — útil para contextualizar a execução atual.
     */
    public List<Episode> getRecentEpisodes(int limit) {
        return getEpisodes().stream()
                .limit(limit)
                .collect(Collectors.toList());
    }

    // Persiste o estado atual no arquivo JSON
    private void persist() {
        try {
            mapper.writeValue(storageFile, episodes);
        } catch (IOException e) {
            throw new RuntimeException("Falha ao persistir memória episódica: " + e.getMessage(), e);
        }
    }
}
