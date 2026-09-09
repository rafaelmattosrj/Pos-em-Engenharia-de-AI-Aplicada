package routing

import (
	"encoding/json"
	"fmt"
	"strings"

	"psp-routing-intelligence/domain"
)

// ParsedRecommendation e a recomendacao decodificada da resposta JSON do
// LLM — ainda sem SimilarCases, que e anexado pelo Service apos a busca de
// similaridade (equivalente a com.psprouting.application.ParsedRecommendation).
type ParsedRecommendation struct {
	Primary    domain.PSP
	Confidence float64
	Reasoning  string
	Fallback   []domain.PSP
}

// InvalidLLMResponseError e retornado quando a resposta do LLM nao pode ser
// interpretada como o JSON de recomendacao esperado — equivalente a
// InvalidLlmResponseException da versao Java.
type InvalidLLMResponseError struct {
	Message string
}

func (e *InvalidLLMResponseError) Error() string {
	return e.Message
}

type rawRecommendation struct {
	Primary    *string  `json:"primary"`
	Confidence *float64 `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
	Fallback   []string `json:"fallback"`
}

// ParseRecommendation decodifica a resposta em texto do LLM no
// ParsedRecommendation esperado pelo prompt RAG. O prompt pede JSON puro
// ("sem markdown, sem texto adicional"), mas LLMs gratuitos as vezes
// envolvem a resposta em blocos ```json ... ``` — por isso a extracao
// remove esses marcadores antes de desserializar, em vez de assumir que o
// modelo sempre obedece a instrucao a risca (mesmo tratamento defensivo de
// RecommendationParser na versao Java).
func ParseRecommendation(rawResponse string) (ParsedRecommendation, error) {
	cleaned := stripMarkdownFences(rawResponse)

	var raw rawRecommendation
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		return ParsedRecommendation{}, &InvalidLLMResponseError{
			Message: fmt.Sprintf("resposta do LLM nao e um JSON valido: %s", truncate(rawResponse)),
		}
	}

	if raw.Primary == nil || raw.Confidence == nil {
		return ParsedRecommendation{}, &InvalidLLMResponseError{
			Message: fmt.Sprintf("resposta do LLM nao contem os campos obrigatorios (primary/confidence): %s", truncate(rawResponse)),
		}
	}

	primary, err := domain.ParsePSP(*raw.Primary)
	if err != nil {
		return ParsedRecommendation{}, &InvalidLLMResponseError{
			Message: fmt.Sprintf("PSP desconhecido na resposta do LLM: %q — %s", *raw.Primary, truncate(rawResponse)),
		}
	}

	fallback := make([]domain.PSP, 0, len(raw.Fallback))
	for _, f := range raw.Fallback {
		psp, err := domain.ParsePSP(f)
		if err != nil {
			return ParsedRecommendation{}, &InvalidLLMResponseError{
				Message: fmt.Sprintf("PSP desconhecido na resposta do LLM: %q — %s", f, truncate(rawResponse)),
			}
		}
		fallback = append(fallback, psp)
	}

	return ParsedRecommendation{
		Primary:    primary,
		Confidence: *raw.Confidence,
		Reasoning:  raw.Reasoning,
		Fallback:   fallback,
	}, nil
}

func stripMarkdownFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	if idx := strings.IndexByte(trimmed, '\n'); idx != -1 {
		trimmed = trimmed[idx+1:]
	}
	if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
		trimmed = trimmed[:idx]
	}
	return strings.TrimSpace(trimmed)
}

func truncate(message string) string {
	const maxLength = 200
	if len(message) > maxLength {
		return message[:maxLength]
	}
	return message
}
