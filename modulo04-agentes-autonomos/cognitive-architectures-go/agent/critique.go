package agent

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cognitive-architectures/model"
)

const critiquePromptTemplate = `Você é um avaliador crítico de qualidade. Avalie o OUTPUT abaixo considerando a TAREFA original.

TAREFA: %s

OUTPUT GERADO:
%s

Avalie em três dimensões com pontuação de 0.0 a 1.0 e retorne EXATAMENTE neste formato:
CORRECTNESS: <0.0-1.0>
COMPLETENESS: <0.0-1.0>
QUALITY: <0.0-1.0>
FEEDBACK: <texto com sugestões de melhoria>

Critérios:
- CORRECTNESS: o output é factualmente correto e logicamente consistente?
- COMPLETENESS: o output cobre todos os aspectos da tarefa solicitada?
- QUALITY: o output é claro, bem estruturado e fácil de entender?
`

// CritiqueEvaluator avalia a qualidade de um output usando o LLM como juiz
// em três dimensões independentes — equivalente a CritiqueEvaluator.java.
type CritiqueEvaluator struct {
	Client    ChatClient
	Threshold float64
}

// Evaluate avalia output em relação a originalTask.
func (e *CritiqueEvaluator) Evaluate(ctx context.Context, originalTask, output string) (model.CritiqueResult, error) {
	prompt := fmt.Sprintf(critiquePromptTemplate, originalTask, output)

	response, err := e.Client.Chat(ctx, prompt)
	if err != nil {
		return model.CritiqueResult{}, err
	}

	return parseCritiqueResponse(response, e.threshold()), nil
}

func (e *CritiqueEvaluator) threshold() float64 {
	if e.Threshold == 0 {
		return 0.7
	}
	return e.Threshold
}

var scorePatterns = map[string]*regexp.Regexp{
	"CORRECTNESS":  regexp.MustCompile(`(?i)CORRECTNESS:\s*([0-9]*\.?[0-9]+)`),
	"COMPLETENESS": regexp.MustCompile(`(?i)COMPLETENESS:\s*([0-9]*\.?[0-9]+)`),
	"QUALITY":      regexp.MustCompile(`(?i)QUALITY:\s*([0-9]*\.?[0-9]+)`),
}

func parseCritiqueResponse(response string, threshold float64) model.CritiqueResult {
	correctness := extractScore(response, "CORRECTNESS")
	completeness := extractScore(response, "COMPLETENESS")
	quality := extractScore(response, "QUALITY")
	feedback := extractFeedback(response)

	score := (correctness + completeness + quality) / 3.0
	passed := score >= threshold

	return model.CritiqueResult{
		Score: score, Passed: passed, Feedback: feedback,
		Correctness: correctness, Completeness: completeness, Quality: quality,
	}
}

// extractScore extrai o valor numérico de uma linha "FIELD: 0.X" via regex.
// Retorna 0.5 (neutro) quando o campo não é encontrado ou não parseia.
func extractScore(text, field string) float64 {
	pattern, ok := scorePatterns[field]
	if !ok {
		return 0.5
	}
	matches := pattern.FindStringSubmatch(text)
	if matches == nil {
		return 0.5
	}
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.5
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// extractFeedback extrai o texto de feedback da resposta estruturada.
func extractFeedback(text string) string {
	idx := strings.Index(text, "FEEDBACK:")
	if idx == -1 {
		return "Nenhum feedback fornecido."
	}
	feedback := strings.TrimSpace(text[idx+len("FEEDBACK:"):])
	if feedback == "" {
		return "Nenhum feedback fornecido."
	}
	return feedback
}
