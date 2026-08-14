package service

import (
	"fmt"
	"os"
	"path/filepath"
)

// RunbookService consulta a base de runbooks — equivalente a
// consult_runbook, definido em labs/modulo10_remediation.py (RAG sobre
// runbooks).
type RunbookService struct {
	DataPath string
}

// ConsultRunbook lê o runbook oficial de um serviço específico e retorna os
// passos de remediação — equivalente a consult_runbook.
func (s RunbookService) ConsultRunbook(serviceName string) string {
	dataPath := s.DataPath
	if dataPath == "" {
		dataPath = "./data"
	}

	runbookPath := filepath.Join(dataPath, "runbook_"+serviceName+".md")
	content, err := os.ReadFile(runbookPath)
	if err != nil {
		return fmt.Sprintf("Error: Runbook for service '%s' not found.", serviceName)
	}
	return string(content)
}
