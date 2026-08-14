// Nexus Tools — ferramentas de AIOps/DevSecOps/FinOps expostas como
// endpoints HTTP. Porte Go das tools/*.py (decoradas com @tool do CrewAI)
// do projeto Nexus AIOps do módulo 06 do curso. A orquestração multi-agente
// hierárquica do CrewAI (core/agents.py, labs/*.py) NÃO é portada — apenas
// a lógica de negócio das ferramentas.
package main

import (
	"fmt"
	"log"
	"net/http"

	"nexus-tools/handler"
	"nexus-tools/service"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	dataPath := getEnv("DATA_PATH", "./data")

	handlers := &handler.Handlers{
		Runbook: service.RunbookService{DataPath: dataPath},
	}

	mux := http.NewServeMux()
	handlers.Register(mux)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("nexus-tools ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
