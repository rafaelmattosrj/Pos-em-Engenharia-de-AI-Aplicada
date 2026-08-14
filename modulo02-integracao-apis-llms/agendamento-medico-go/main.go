// Agendamento médico via chat — identifica a intenção do paciente (agendar,
// cancelar ou desconhecida) usando um LLM com saída estruturada e roteia
// para o serviço correspondente. Porte Go do projeto Spring Boot
// agendamento-medico-java.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"agendamento-medico/graph"
	"agendamento-medico/llm"
	"agendamento-medico/openrouter"
	"agendamento-medico/service"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "3000")
	apiKey := getEnv("OPENROUTER_API_KEY", "")
	temperature := getEnvFloat("APP_TEMPERATURE", 0.2)
	maxTokens := getEnvInt("APP_MAX_TOKENS", 500)
	fallbackModels := strings.Split(getEnv("APP_FALLBACK_MODELS",
		"qwen/qwen3.6-plus:free,qwen/qwen3-coder:free,openai/gpt-oss-120b:free,stepfun/step-3.5-flash:free,nvidia/nemotron-3-super-120b-a12b:free"),
		",")

	resilientClient := &llm.ResilientClient{
		OpenRouter:     openrouter.NewClient(apiKey),
		FallbackModels: fallbackModels,
		Temperature:    temperature,
		MaxTokens:      maxTokens,
	}

	orchestrator := &graph.Orchestrator{
		IntentService:      &service.IntentService{Client: resilientClient},
		AppointmentService: service.NewAppointmentService(),
		MessageGenerator:   &service.MessageGeneratorService{Client: resilientClient},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/chat", chatHandler(orchestrator))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("agendamento-medico ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
