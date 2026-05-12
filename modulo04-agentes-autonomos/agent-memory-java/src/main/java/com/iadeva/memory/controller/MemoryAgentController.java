package com.iadeva.memory.controller;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.iadeva.memory.agent.MemoryAwareAgent;
import com.iadeva.memory.engine.ReflectionEngine;
import com.iadeva.memory.memory.EpisodicMemory;
import com.iadeva.memory.model.AgentRunRequest;
import com.iadeva.memory.model.AgentRunResult;
import com.iadeva.memory.model.Episode;
import com.iadeva.memory.model.Lesson;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.File;
import java.util.List;
import java.util.Map;

/**
 * Controller REST do sistema de memória.
 * Expõe endpoints para execução com/sem memória e consulta de episódios e lições.
 */
@RestController
@RequestMapping("/agent-memory")
public class MemoryAgentController {

    private final MemoryAwareAgent memoryAwareAgent;
    private final EpisodicMemory episodicMemory;
    private final ReflectionEngine reflectionEngine;
    private final ChatClient chatClient;
    private final String memoryPath;
    private final ObjectMapper mapper;

    public MemoryAgentController(
            MemoryAwareAgent memoryAwareAgent,
            EpisodicMemory episodicMemory,
            ReflectionEngine reflectionEngine,
            ChatClient.Builder chatClientBuilder,
            @Value("${app.memory.path:./data}") String memoryPath) {
        this.memoryAwareAgent = memoryAwareAgent;
        this.episodicMemory = episodicMemory;
        this.reflectionEngine = reflectionEngine;
        this.chatClient = chatClientBuilder.build();
        this.memoryPath = memoryPath;
        this.mapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    /**
     * POST /agent-memory/run
     * Executa o agente com todos os 4 tipos de memória ativos.
     */
    @PostMapping("/run")
    public ResponseEntity<AgentRunResult> runWithMemory(@RequestBody AgentRunRequest request) {
        AgentRunResult result = memoryAwareAgent.run(request.input());
        return ResponseEntity.ok(result);
    }

    /**
     * POST /agent-memory/run-without
     * Executa sem memória — apenas ChatClient direto — para comparação de qualidade.
     */
    @PostMapping("/run-without")
    public ResponseEntity<AgentRunResult> runWithoutMemory(@RequestBody AgentRunRequest request) {
        String output = chatClient.prompt()
                .user(request.input())
                .call()
                .content();

        AgentRunResult result = new AgentRunResult(output, false, 0, 0);
        return ResponseEntity.ok(result);
    }

    /**
     * GET /agent-memory/episodes
     * Lista todos os episódios registrados, do mais recente ao mais antigo.
     */
    @GetMapping("/episodes")
    public ResponseEntity<List<Episode>> getEpisodes() {
        return ResponseEntity.ok(episodicMemory.getEpisodes());
    }

    /**
     * GET /agent-memory/lessons
     * Lista todas as lições extraídas pela ReflectionEngine.
     */
    @GetMapping("/lessons")
    public ResponseEntity<List<Lesson>> getLessons() {
        File lessonsFile = new File(memoryPath + "/lessons.json");
        if (!lessonsFile.exists()) {
            return ResponseEntity.ok(List.of());
        }
        try {
            List<Lesson> lessons = mapper.readValue(lessonsFile, new TypeReference<List<Lesson>>() {});
            return ResponseEntity.ok(lessons);
        } catch (Exception e) {
            return ResponseEntity.ok(List.of());
        }
    }

    /**
     * DELETE /agent-memory/clear
     * Remove todos os arquivos de memória persistida (long-term, episodes, lessons).
     */
    @DeleteMapping("/clear")
    public ResponseEntity<Map<String, String>> clearMemory() {
        List<String> arquivos = List.of("long-term.json", "episodes.json", "lessons.json");
        List<String> removidos = new java.util.ArrayList<>();

        for (String nome : arquivos) {
            File arquivo = new File(memoryPath + "/" + nome);
            if (arquivo.exists() && arquivo.delete()) {
                removidos.add(nome);
            }
        }

        return ResponseEntity.ok(Map.of(
                "status", "sucesso",
                "arquivosRemovidos", String.join(", ", removidos)
        ));
    }
}
