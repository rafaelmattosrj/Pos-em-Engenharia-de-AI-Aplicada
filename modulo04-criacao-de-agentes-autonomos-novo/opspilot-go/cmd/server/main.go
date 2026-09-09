// Porte de src/index.ts (bootstrapOpsPilot + main). Sem OPENROUTER_API_KEY
// configurada, roda com um FakeChatModel (respostas fixas) -- ver README:
// uma implementação real de llm.ChatModel plugaria um provedor
// OpenAI-compatible aqui no lugar do fake.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"opspilot/domain"
	"opspilot/httpapi"
	"opspilot/llm"
	"opspilot/store"
	"opspilot/strategies"
	"opspilot/team"
	"opspilot/tools"
)

func main() {
	opsStore := store.NewInMemoryOpsStore()
	seed(opsStore)

	model := llm.NewFakeChatModelFunc(canned)

	analistaTools := []tools.Tool{
		tools.ListAlerts(opsStore),
		tools.ListIncidents(opsStore),
		tools.ConsultarRunbook(opsStore),
		tools.CheckProviderStatus(),
	}
	executorTools := []tools.Tool{
		tools.OpenIncident(opsStore),
		tools.ResolveIncident(opsStore),
		tools.ListIncidents(opsStore),
	}

	react := strategies.NewReactStrategy(model, analistaTools)
	teamStrategy := team.NewTeamStrategy(
		team.CreateDecideNext(model),
		map[team.Role]team.RoleRunner{
			team.RoleAnalista:   team.NewAnalistaRunner(model, analistaTools),
			team.RolePlanejador: team.NewPlanejadorRunner(model),
			team.RoleExecutor:   team.NewExecutorRunner(model, executorTools),
		},
		1,
	)

	strategyMap := map[string]domain.ReasoningStrategy{
		react.Name():        react,
		teamStrategy.Name(): teamStrategy,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.Handle("/chat", httpapi.NewChatHandler(strategyMap, "team"))

	addr := ":" + port
	fmt.Printf("OpsPilot (nucleo portado) ouvindo em http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func seed(opsStore store.OpsStore) {
	opsStore.SeedAlert(domain.Alert{
		ID: "alert-1", Service: "checkout-api", Description: "Latencia acima de 2s",
		Severity: domain.SeverityHigh, Status: domain.AlertFiring,
	})
	opsStore.SeedRunbook(domain.Runbook{
		Service: "checkout-api",
		Content: "1. Verificar pool de conexoes.\n2. Escalar replicas.\n3. Acionar time de pagamentos se persistir.",
	})
}

// canned é uma resposta fixa e determinística -- adequada a demo/smoke test,
// não a uso real.
func canned(messages []llm.ChatMessage) (llm.ModelResponse, error) {
	return llm.TextResponse(`{"next":"done","brief":"Diagnostico concluido (modelo fake, sem chamada real)."}`), nil
}
