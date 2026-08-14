package service

import (
	"errors"
	"os/exec"
)

// runCommand executa um comando externo e retorna sua saída combinada
// (stdout+stderr), igual ao capture_output=True, text=True do
// subprocess.run em Python.
func runCommand(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).CombinedOutput()
	return string(output), err
}

// isCommandNotFoundErr detecta se o erro indica que o binário não foi
// encontrado no PATH — equivalente ao except FileNotFoundError em Python.
func isCommandNotFoundErr(err error) bool {
	var execErr *exec.Error
	return err != nil && errors.As(err, &execErr)
}
