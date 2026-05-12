package com.iadeva.memory.engine;

// Equivalente ao reflection_engine.py da aula 14 — extrai lições reutilizáveis de execuções passadas

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.iadeva.memory.model.Episode;
import com.iadeva.memory.model.Lesson;
import jakarta.annotation.PostConstruct;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Motor de reflexão evolutiva: analisa episódios de execução e extrai lições generalizáveis.
 * Só processa episódios que contêm erros ou comportamentos notáveis — evita ruído.
 * Lições persistidas em ./data/lessons.json e reutilizadas em execuções futuras.
 */
@Component
public class ReflectionEngine {

    private final ChatClient chatClient;
    private final ObjectMapper mapper;
    private final File lessonsFile;

    // Lições acumuladas entre execuções
    private List<Lesson> lessons = new ArrayList<>();

    public ReflectionEngine(
            ChatClient.Builder chatClientBuilder,
            @Value("${app.memory.path:./data}") String memoryPath) {
        this.chatClient = chatClientBuilder.build();
        this.mapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        this.lessonsFile = new File(memoryPath + "/lessons.json");
    }

    @PostConstruct
    public void init() {
        lessonsFile.getParentFile().mkdirs();
        if (lessonsFile.exists()) {
            try {
                Lesson[] loaded = mapper.readValue(lessonsFile, Lesson[].class);
                lessons = new ArrayList<>(Arrays.asList(loaded));
            } catch (IOException e) {
                lessons = new ArrayList<>();
            }
        }
    }

    /**
     * Analisa um episódio e extrai lições, se houver erro ou comportamento notável.
     * Retorna lista vazia se o episódio não contém aprendizados relevantes.
     *
     * @param episode episódio de execução a ser refletido
     * @return lições extraídas (pode ser vazia)
     */
    public List<Lesson> reflect(Episode episode) {
        // Só reflete quando há indicação de erro ou resultado não trivial
        boolean merecereflexao = episode.outcome() != null
                && (episode.outcome().toLowerCase().contains("erro")
                || episode.outcome().toLowerCase().contains("error")
                || episode.outcome().toLowerCase().contains("falha")
                || episode.outcome().toLowerCase().contains("notável")
                || episode.outcome().toLowerCase().contains("aprendizado"));

        if (!merecereflexao) {
            return List.of();
        }

        String prompt = construirPromptReflexao(episode);

        String resposta = chatClient.prompt()
                .user(prompt)
                .call()
                .content();

        List<Lesson> licoesExtraidas = parsearLicoes(resposta, episode);

        // Persiste apenas lições generalizáveis e acionáveis
        List<Lesson> licoesFiltradas = licoesExtraidas.stream()
                .filter(l -> l.generalizability() != null && !l.generalizability().isBlank())
                .collect(Collectors.toList());

        if (!licoesFiltradas.isEmpty()) {
            lessons.addAll(licoesFiltradas);
            persistirLicoes();
        }

        return licoesFiltradas;
    }

    /**
     * Retorna todas as lições acumuladas.
     */
    public List<Lesson> getLessons() {
        return List.copyOf(lessons);
    }

    // Monta o prompt estruturado para extração de lições
    private String construirPromptReflexao(Episode episode) {
        String passos = episode.steps() != null
                ? String.join("\n- ", episode.steps())
                : "nenhum passo registrado";

        return """
                Analise esta execução de agente e extraia lições aprendidas:
                
                INPUT: %s
                
                PASSOS EXECUTADOS:
                - %s
                
                RESULTADO: %s
                
                Extraia UMA lição no seguinte formato exato (cada campo em uma linha):
                SITUAÇÃO: [descrição da situação que gerou o aprendizado]
                AÇÃO: [o que foi feito]
                RESULTADO: [o que aconteceu]
                APRENDIZADO: [lição concisa e acionável]
                GENERALIZABILIDADE: [em quais outras situações essa lição se aplica — deixe em branco se não for generalizável]
                
                Se não houver lição relevante, responda apenas: SEM_LICAO
                """.formatted(episode.input(), passos, episode.outcome());
    }

    // Parseia a resposta do LLM no formato estruturado definido no prompt
    private List<Lesson> parsearLicoes(String resposta, Episode episode) {
        if (resposta == null || resposta.isBlank() || resposta.contains("SEM_LICAO")) {
            return List.of();
        }

        try {
            String situacao = extrairCampo(resposta, "SITUAÇÃO:");
            String acao = extrairCampo(resposta, "AÇÃO:");
            String resultado = extrairCampo(resposta, "RESULTADO:");
            String aprendizado = extrairCampo(resposta, "APRENDIZADO:");
            String generalizabilidade = extrairCampo(resposta, "GENERALIZABILIDADE:");

            if (aprendizado == null || aprendizado.isBlank()) {
                return List.of();
            }

            Lesson licao = new Lesson(
                    UUID.randomUUID().toString(),
                    situacao,
                    acao,
                    resultado,
                    aprendizado,
                    generalizabilidade,
                    Instant.now()
            );

            return List.of(licao);
        } catch (Exception e) {
            return List.of();
        }
    }

    private String extrairCampo(String texto, String prefixo) {
        return Arrays.stream(texto.split("\n"))
                .filter(linha -> linha.startsWith(prefixo))
                .map(linha -> linha.substring(prefixo.length()).trim())
                .findFirst()
                .orElse("");
    }

    private void persistirLicoes() {
        try {
            mapper.writeValue(lessonsFile, lessons);
        } catch (IOException e) {
            throw new RuntimeException("Falha ao persistir lições: " + e.getMessage(), e);
        }
    }
}
