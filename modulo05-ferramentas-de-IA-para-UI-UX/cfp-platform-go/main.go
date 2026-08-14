// CFP Platform API — backend mínimo que sustenta o front-end Angular do CFP
// (Call for Papers), presente de forma idêntica em modulo-03/cfp-platform e
// modulo-04/cfp-plataform_v1 do curso. Porte Go do backend NestJS original
// (apenas a lógica de servidor — a UI Angular permanece fora de escopo).
package main

import (
	"fmt"
	"log"
	"net/http"

	"cfp-platform/handler"
	"cfp-platform/service"
)

func main() {
	loadDotEnv(".env")
	port := getEnv("PORT", "3000")

	eventHandler := &handler.EventHandler{Service: &service.EventService{}}
	speakerHandler := &handler.SpeakerHandler{Service: &service.SpeakerService{}}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api", handler.Health)
	mux.HandleFunc("POST /api/events", eventHandler.Create)
	mux.HandleFunc("GET /api/events", eventHandler.FindAll)
	mux.HandleFunc("POST /api/speakers", speakerHandler.Create)
	mux.HandleFunc("GET /api/speakers", speakerHandler.FindAll)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 cfp-platform ouvindo em http://localhost%s/api\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
