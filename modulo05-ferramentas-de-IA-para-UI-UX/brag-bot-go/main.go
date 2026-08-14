// Brag Bot API — transforma um rascunho informal de uma realização
// profissional em um "Brag Document" executivo estruturado, usando o
// Gemini. Porte Go da rota de API do backend Express/Angular SSR de
// modulo-05/brag-bot (apenas a lógica de servidor — a UI Angular permanece
// fora de escopo).
package main

import (
	"fmt"
	"log"
	"net/http"

	"brag-bot/gemini"
	"brag-bot/handler"
	"brag-bot/service"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "4000")
	apiKey := getEnv("GEMINI_API_KEY", "")
	model := getEnv("GEMINI_MODEL", "gemini-2.5-flash")

	bragHandler := &handler.BragHandler{
		Service: &service.BragService{Client: gemini.NewClient(apiKey, model)},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/brag", bragHandler.Generate)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("brag-bot ouvindo em http://localhost%s/api/brag\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
