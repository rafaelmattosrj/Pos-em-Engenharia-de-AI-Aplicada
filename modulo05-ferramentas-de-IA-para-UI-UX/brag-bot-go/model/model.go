// Package model define os tipos de domínio — equivalente ao BragInputSchema
// e BragSchema (Zod) de flows.ts.
package model

// BragRequest é o payload recebido em POST /api/brag.
type BragRequest struct {
	Definition string `json:"definition"`
}

// BragDocument é o "Brag Document" executivo gerado a partir do rascunho
// informal do usuário.
type BragDocument struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Context          string   `json:"context"`
	ActionTaken      string   `json:"actionTaken"`
	BusinessImpact   string   `json:"businessImpact"`
	Metrics          []string `json:"metrics"`
	TechnologiesUsed []string `json:"technologiesUsed"`
}
