// Package dataset carrega o dataset de avaliação — equivalente a
// EvalDataset.java. O JSON é embarcado no binário via go:embed (equivalente
// a ClassPathResource, que empacota o arquivo dentro do jar), garantindo que
// esteja sempre disponível independente do diretório de trabalho.
package dataset

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"agent-evals/model"
)

//go:embed eval-dataset.json
var evalDatasetJSON []byte

// EvalDataset é o componente responsável por expor os cenários de avaliação
// carregados na inicialização.
type EvalDataset struct {
	scenarios []model.EvalScenario
}

// Load carrega o dataset embarcado. Falha imediatamente (panic) se o JSON
// estiver malformado — equivalente ao IllegalStateException lançado por
// EvalDataset.load() em caso de falha.
func Load() *EvalDataset {
	var scenarios []model.EvalScenario
	if err := json.Unmarshal(evalDatasetJSON, &scenarios); err != nil {
		panic(fmt.Sprintf("falha ao carregar eval-dataset.json: %v", err))
	}
	return &EvalDataset{scenarios: scenarios}
}

// Scenarios retorna todos os cenários do dataset.
func (d *EvalDataset) Scenarios() []model.EvalScenario {
	result := make([]model.EvalScenario, len(d.scenarios))
	copy(result, d.scenarios)
	return result
}
