// Porte de src/cli.ts -- CLI sobre um JSONFileTaskStore.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"notas-api/cli"
	"notas-api/service"
	"notas-api/store"
)

func main() {
	storePath := os.Getenv("TASK_CLI_STORE_PATH")
	if storePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		storePath = filepath.Join(cwd, ".tasks-cli-store.json")
	}

	taskStore, err := store.NewJSONFileTaskStore(storePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	taskService := service.New(taskStore)
	os.Exit(runGuarded(taskService))
}

// runGuarded recupera de um *store.PersistenceError vindo de dentro de
// persist() -- o original (cli.ts) captura TaskStorePersistenceError tanto na
// construcao do store quanto durante os comandos, e imprime so a mensagem.
func runGuarded(taskService *service.TaskService) (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			if persistErr, ok := r.(*store.PersistenceError); ok {
				fmt.Fprintln(os.Stderr, persistErr.Error())
				exitCode = 1
				return
			}
			panic(r)
		}
	}()
	return cli.Run(os.Args[1:], taskService, os.Stdout, os.Stderr)
}
