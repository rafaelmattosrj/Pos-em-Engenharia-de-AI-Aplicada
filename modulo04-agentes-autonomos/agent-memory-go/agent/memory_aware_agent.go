// Package agent implementa o agente com consciência de memória — equivalente
// a MemoryAwareAgent.java (aula 13 do curso).
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"agent-memory/engine"
	"agent-memory/memory"
	"agent-memory/model"

	"github.com/google/uuid"
)

// ChatClient é a única operação de que MemoryAwareAgent depende para gerar
// a resposta enriquecida.
type ChatClient interface {
	Chat(ctx context.Context, userPrompt string) (string, error)
}

// MemoryAwareAgent integra os 4 tipos de memória no ciclo de execução:
// contexto semântico → fatos → episódios → execução enriquecida →
// persistência → reflexão.
type MemoryAwareAgent struct {
	LongTerm   *memory.LongTermMemory
	Episodic   *memory.EpisodicMemory
	Contextual *memory.ContextualMemory
	Reflection *engine.ReflectionEngine
	Client     ChatClient
}

// Run executa o agente com contexto enriquecido por todos os 4 tipos de
// memória — equivalente a MemoryAwareAgent.run.
func (a *MemoryAwareAgent) Run(ctx context.Context, input string) (model.AgentRunResult, error) {
	shortTerm := memory.NewShortTermMemory()
	var steps []string
	defer shortTerm.Clear()

	// PASSO 1: contexto semântico via ContextualMemory
	steps = append(steps, "Buscando contexto semântico via embeddings")
	semanticContext := a.searchContext(ctx, input)
	shortTerm.Put("contexto_semantico", semanticContext)

	// PASSO 2: fatos relevantes de LongTermMemory
	steps = append(steps, "Recuperando fatos da memória de longo prazo")
	facts := a.LongTerm.GetFacts()
	shortTerm.Put("fatos", facts)

	// PASSO 3: episódios recentes de EpisodicMemory
	steps = append(steps, "Consultando episódios recentes")
	recentEpisodes := a.Episodic.GetRecentEpisodes(3)
	shortTerm.Put("episodios_recentes", recentEpisodes)

	// PASSO 4: execução com contexto enriquecido
	steps = append(steps, "Executando com contexto enriquecido")
	enrichedPrompt := buildEnrichedPrompt(input, facts, recentEpisodes, semanticContext)
	output, err := a.Client.Chat(ctx, enrichedPrompt)
	if err != nil {
		return model.AgentRunResult{}, err
	}
	shortTerm.Put("output", output)

	// PASSO 5: persiste episódio + extrai lições via ReflectionEngine
	steps = append(steps, "Persistindo episódio e extraindo lições")
	episode := model.Episode{
		ID:               uuid.NewString(),
		Input:            input,
		Steps:            append([]string{}, steps...),
		Outcome:          output,
		LessonsExtracted: nil,
		Timestamp:        time.Now(),
	}
	a.Episodic.AddEpisode(episode)

	// Armazena o input no índice semântico para futuras execuções.
	a.Contextual.Store(ctx, input, map[string]any{"tipo": "input", "timestamp": time.Now().Format(time.RFC3339)})

	lessons := a.Reflection.Reflect(ctx, episode)
	if len(lessons) > 0 {
		steps = append(steps, fmt.Sprintf("Lições extraídas: %d", len(lessons)))
	}

	return model.AgentRunResult{
		Output:             output,
		MemoryUsed:         true,
		FactsRetrieved:     len(facts),
		EpisodesConsidered: len(recentEpisodes),
	}, nil
}

// searchContext busca contexto semântico — retorna vazio se a memória
// contextual ainda não tem entradas ou a busca falhar (primeira execução).
func (a *MemoryAwareAgent) searchContext(ctx context.Context, input string) []memory.EmbeddingEntry {
	results, err := a.Contextual.Search(ctx, input, 3)
	if err != nil {
		return nil
	}
	return results
}

func buildEnrichedPrompt(input string, facts []model.MemoryFact, episodes []model.Episode, context []memory.EmbeddingEntry) string {
	var sb strings.Builder

	if len(context) > 0 {
		sb.WriteString("=== CONTEXTO SEMÂNTICO RELEVANTE ===\n")
		for _, e := range context {
			sb.WriteString("- " + e.Content + "\n")
		}
		sb.WriteString("\n")
	}

	if len(facts) > 0 {
		sb.WriteString("=== FATOS CONHECIDOS ===\n")
		for _, f := range facts {
			sb.WriteString(fmt.Sprintf("- %s [fonte: %s]\n", f.Content, f.Source))
		}
		sb.WriteString("\n")
	}

	if len(episodes) > 0 {
		sb.WriteString("=== HISTÓRICO RECENTE ===\n")
		for _, e := range episodes {
			sb.WriteString(fmt.Sprintf("Input anterior: %s → Resultado: %s\n", e.Input, summarize(e.Outcome)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== TAREFA ATUAL ===\n")
	sb.WriteString(input)
	return sb.String()
}

func summarize(text string) string {
	if text == "" {
		return "sem resultado"
	}
	if len(text) > 100 {
		return text[:100] + "..."
	}
	return text
}
