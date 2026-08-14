package agent

import (
	"context"
	"testing"
)

type stubChatClient struct {
	responses []string
	calls     int
}

func (s *stubChatClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	response := s.responses[s.calls%len(s.responses)]
	s.calls++
	return response, nil
}

// Cenário 1 de CognitiveArchitecturesTest.java: score abaixo do threshold
// resulta em passed=false.
func TestEvaluate_ReturnsPassedFalseWhenScoreBelowThreshold(t *testing.T) {
	client := &stubChatClient{responses: []string{
		"CORRECTNESS: 0.4\nCOMPLETENESS: 0.5\nQUALITY: 0.6\nFEEDBACK: Resposta incompleta e com imprecisões.",
	}}
	evaluator := &CritiqueEvaluator{Client: client, Threshold: 0.7}

	result, err := evaluator.Evaluate(context.Background(), "Explique machine learning", "ML é uma coisa.")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if result.Passed {
		t.Error("esperava passed=false")
	}
	if result.Score >= 0.7 {
		t.Errorf("esperava score < 0.7, obteve %f", result.Score)
	}
	if result.Feedback == "" {
		t.Error("esperava feedback nao vazio")
	}
	if result.Correctness >= 0.7 || result.Completeness >= 0.7 {
		t.Errorf("esperava correctness/completeness < 0.7, obteve %f/%f", result.Correctness, result.Completeness)
	}
	if client.calls != 1 {
		t.Errorf("esperava 1 chamada ao LLM, obteve %d", client.calls)
	}
}

func TestEvaluate_ReturnsPassedTrueWhenScoreAboveThreshold(t *testing.T) {
	client := &stubChatClient{responses: []string{
		"CORRECTNESS: 0.9\nCOMPLETENESS: 0.85\nQUALITY: 0.8\nFEEDBACK: Excelente resposta.",
	}}
	evaluator := &CritiqueEvaluator{Client: client, Threshold: 0.7}

	result, err := evaluator.Evaluate(context.Background(), "tarefa", "resposta completa")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.Passed {
		t.Error("esperava passed=true")
	}
}

func TestExtractScore_DefaultsToNeutralWhenFieldMissing(t *testing.T) {
	score := extractScore("resposta sem os campos esperados", "CORRECTNESS")
	if score != 0.5 {
		t.Errorf("esperava score neutro 0.5, obteve %f", score)
	}
}
