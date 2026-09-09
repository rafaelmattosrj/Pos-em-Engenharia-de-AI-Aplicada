// Porte de src/index.ts -- sobe o servidor HTTP na porta 3000.
package main

import (
	"fmt"
	"net/http"

	"notas-api/httpapi"
	"notas-api/service"
	"notas-api/store"
)

func main() {
	taskService := service.New(store.NewInMemoryTaskStore())
	mux := http.NewServeMux()
	mux.Handle("/tasks", httpapi.NewHandler(taskService))
	mux.Handle("/tasks/", httpapi.NewHandler(taskService))

	port := "3000"
	fmt.Printf("HTTP server listening on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}
