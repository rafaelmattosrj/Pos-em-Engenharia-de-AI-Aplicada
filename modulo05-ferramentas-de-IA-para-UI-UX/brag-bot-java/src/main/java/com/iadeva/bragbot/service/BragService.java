package com.iadeva.bragbot.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.bragbot.gemini.GeminiClient;
import com.iadeva.bragbot.model.BragDocument;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * Equivalente a bragGeneratorFlow de flows.ts: monta o prompt, chama o
 * modelo e monta o BragDocument final com um novo id gerado pelo servidor.
 */
@Service
public class BragService {

    private static final String PROMPT_TEMPLATE = """
            Persona: Você deve atuar como um "Senior Career Consultant" focado em Planos de Desenvolvimento Individual (IDP) para Engenheiros de Software.
            Objetivo: Transformar o rascunho informal do usuário em um "Brag Document" executivo.

            Regra 1: Usar tom profissional, objetivo e focado em impacto, sem adjetivos emocionais.
            Regra 2: Se não existirem métricas exatas, infira a natureza da métrica baseada na ação tomada de forma plausível (ex: "tempo de execução não especificado mas otimizado").
            Regra 3: Responda APENAS com um JSON válido, sem markdown, com os campos exatos:
              title (string): Ação principal + Resultado de alto nível.
              context (string): Situação ou Problema original. O que estava quebrado, lento, o desafio, etc.
              actionTaken (string): Ação técnica ou estratégica passo a passo tomada para resolver o problema.
              businessImpact (string): Qual o impacto de negócio. Tempo ganho, redução de falhas, etc.
              metrics (array de strings): Apenas dados estritamente quantificáveis. Ex: "50% reduction", "10ms latency".
              technologiesUsed (array de strings): Ferramentas, linguagens, bibliotecas e plataformas mencionadas ou inferidas.
            Regra 4: O output deve respeitar o idioma original do input.

            Aqui está o rascunho informal do usuário:
            """;

    private final GeminiClient geminiClient;
    private final ObjectMapper objectMapper;

    public BragService(GeminiClient geminiClient, ObjectMapper objectMapper) {
        this.geminiClient = geminiClient;
        this.objectMapper = objectMapper;
    }

    /**
     * Gera um BragDocument a partir do rascunho informal do usuário.
     *
     * @throws IllegalStateException se o modelo não retornar um JSON válido
     */
    public BragDocument generate(String definition) {
        // Concatenação simples (não String.formatted/String.format) porque o
        // template contém "%" literal (ex.: "50% reduction") que colidiria
        // com a sintaxe de conversão de formato.
        String prompt = PROMPT_TEMPLATE + definition;
        String json = geminiClient.generateJson(prompt, 0.8);

        try {
            JsonNode node = objectMapper.readTree(json);

            return new BragDocument(
                    UUID.randomUUID().toString(),
                    node.path("title").asText(""),
                    node.path("context").asText(""),
                    node.path("actionTaken").asText(""),
                    node.path("businessImpact").asText(""),
                    toStringList(node.path("metrics")),
                    toStringList(node.path("technologiesUsed"))
            );
        } catch (Exception e) {
            throw new IllegalStateException("Falha ao gerar o conteúdo: Nenhum output válido foi retornado pela IA.", e);
        }
    }

    private List<String> toStringList(JsonNode arrayNode) {
        List<String> result = new ArrayList<>();
        if (arrayNode.isArray()) {
            arrayNode.forEach(n -> result.add(n.asText()));
        }
        return result;
    }
}
