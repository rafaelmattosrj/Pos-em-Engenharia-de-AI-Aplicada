package com.iadeva.memory.agent;

// Equivalente ao memory_aware_agent.py da aula 13 — agente que integra os 4 tipos de memória no ciclo

import com.iadeva.memory.engine.ReflectionEngine;
import com.iadeva.memory.memory.ContextualMemory;
import com.iadeva.memory.memory.EpisodicMemory;
import com.iadeva.memory.memory.LongTermMemory;
import com.iadeva.memory.memory.ShortTermMemory;
import com.iadeva.memory.model.AgentRunResult;
import com.iadeva.memory.model.Episode;
import com.iadeva.memory.model.Lesson;
import com.iadeva.memory.model.MemoryFact;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

/**
 * Agente com consciência de memória: integra os 4 tipos de memória no ciclo de execução.
 * Fluxo: contexto semântico → fatos → episódios → execução enriquecida → persistência → reflexão.
 */
@Component
public class MemoryAwareAgent {

    private final ApplicationContext applicationContext;
    private final LongTermMemory longTermMemory;
    private final EpisodicMemory episodicMemory;
    private final ContextualMemory contextualMemory;
    private final ReflectionEngine reflectionEngine;
    private final ChatClient chatClient;

    public MemoryAwareAgent(
            ApplicationContext applicationContext,
            LongTermMemory longTermMemory,
            EpisodicMemory episodicMemory,
            ContextualMemory contextualMemory,
            ReflectionEngine reflectionEngine,
            ChatClient.Builder chatClientBuilder) {
        this.applicationContext = applicationContext;
        this.longTermMemory = longTermMemory;
        this.episodicMemory = episodicMemory;
        this.contextualMemory = contextualMemory;
        this.reflectionEngine = reflectionEngine;
        this.chatClient = chatClientBuilder.build();
    }

    /**
     * Executa o agente com contexto enriquecido por todos os 4 tipos de memória.
     *
     * @param input pergunta ou instrução do usuário
     * @return resultado com output e métricas de uso de memória
     */
    public AgentRunResult run(String input) {
        // Obtém instância prototype isolada para esta execução
        ShortTermMemory shortTerm = applicationContext.getBean(ShortTermMemory.class);
        List<String> steps = new ArrayList<>();

        try {
            // PASSO 1: Busca contexto semântico via ContextualMemory
            steps.add("Buscando contexto semântico via embeddings");
            List<ContextualMemory.EmbeddingEntry> contextoSemantico = buscarContexto(input);
            shortTerm.put("contexto_semantico", contextoSemantico);

            // PASSO 2: Recupera fatos relevantes de LongTermMemory
            steps.add("Recuperando fatos da memória de longo prazo");
            List<MemoryFact> fatos = longTermMemory.getFacts();
            shortTerm.put("fatos", fatos);

            // PASSO 3: Consulta episódios recentes de EpisodicMemory
            steps.add("Consultando episódios recentes");
            List<Episode> episodiosRecentes = episodicMemory.getRecentEpisodes(3);
            shortTerm.put("episodios_recentes", episodiosRecentes);

            // PASSO 4: Executa com contexto enriquecido via ChatClient
            steps.add("Executando com contexto enriquecido");
            String promptEnriquecido = construirPromptEnriquecido(input, fatos, episodiosRecentes, contextoSemantico);
            String output = chatClient.prompt()
                    .user(promptEnriquecido)
                    .call()
                    .content();
            shortTerm.put("output", output);

            // PASSO 5: Persiste episódio + extrai lições via ReflectionEngine
            steps.add("Persistindo episódio e extraindo lições");
            Episode episodio = new Episode(
                    UUID.randomUUID().toString(),
                    input,
                    List.copyOf(steps),
                    output,
                    List.of(),
                    Instant.now()
            );
            episodicMemory.addEpisode(episodio);

            // Armazena o input no índice semântico para futuras execuções
            contextualMemory.store(input, Map.of("tipo", "input", "timestamp", Instant.now().toString()));

            // Reflexão: extrai lições se houver aprendizado relevante
            List<Lesson> licoes = reflectionEngine.reflect(episodio);
            if (!licoes.isEmpty()) {
                steps.add("Lições extraídas: " + licoes.size());
            }

            return new AgentRunResult(
                    output,
                    true,
                    fatos.size(),
                    episodiosRecentes.size()
            );

        } finally {
            // Garante limpeza da memória de curto prazo ao fim da execução
            shortTerm.clear();
        }
    }

    // Busca contexto semântico — retorna lista vazia se memória contextual ainda não tem entradas
    private List<ContextualMemory.EmbeddingEntry> buscarContexto(String input) {
        try {
            return contextualMemory.search(input, 3);
        } catch (Exception e) {
            // Memória contextual pode estar vazia na primeira execução
            return List.of();
        }
    }

    // Monta prompt enriquecido com todos os contextos disponíveis
    private String construirPromptEnriquecido(
            String input,
            List<MemoryFact> fatos,
            List<Episode> episodios,
            List<ContextualMemory.EmbeddingEntry> contexto) {

        StringBuilder sb = new StringBuilder();

        // Contexto semântico relevante
        if (!contexto.isEmpty()) {
            sb.append("=== CONTEXTO SEMÂNTICO RELEVANTE ===\n");
            contexto.forEach(e -> sb.append("- ").append(e.content()).append("\n"));
            sb.append("\n");
        }

        // Fatos da memória de longo prazo
        if (!fatos.isEmpty()) {
            sb.append("=== FATOS CONHECIDOS ===\n");
            fatos.forEach(f -> sb.append("- ").append(f.content())
                    .append(" [fonte: ").append(f.source()).append("]\n"));
            sb.append("\n");
        }

        // Histórico de episódios recentes
        if (!episodios.isEmpty()) {
            sb.append("=== HISTÓRICO RECENTE ===\n");
            episodios.forEach(e -> sb.append("Input anterior: ").append(e.input())
                    .append(" → Resultado: ").append(resumir(e.outcome())).append("\n"));
            sb.append("\n");
        }

        sb.append("=== TAREFA ATUAL ===\n").append(input);
        return sb.toString();
    }

    private String resumir(String texto) {
        if (texto == null) return "sem resultado";
        return texto.length() > 100 ? texto.substring(0, 100) + "..." : texto;
    }
}
