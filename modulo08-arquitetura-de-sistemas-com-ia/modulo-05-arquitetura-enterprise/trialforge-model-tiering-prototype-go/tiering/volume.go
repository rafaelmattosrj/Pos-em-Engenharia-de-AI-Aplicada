package tiering

import "sync"

// ResultadoEstudo é o resultado de orçamento de um estudo depois da rodada
// de volume concorrente.
type ResultadoEstudo struct {
	EstudoID string
	Gasto    float64
	Limite   float64
	Estourou bool
}

// SimularVolumeConcorrente dispara todas as requisicoes em goroutines
// simultâneas (equivalente a Promise.all no original em JS, mas com
// paralelismo real de threads do SO, não cooperativo de um único event
// loop) e devolve o resultado de orçamento de cada estudo ao final.
//
// Porte 1:1 de simularVolumeConcorrente em trialforge-model-tiering-prototype.js / .py,
// adaptado para goroutines de verdade (ver comentário em OrcamentoManager).
func SimularVolumeConcorrente(orcamento *OrcamentoManager, estudos []string, requisicoes []func()) []ResultadoEstudo {
	var wg sync.WaitGroup
	wg.Add(len(requisicoes))
	for _, requisicao := range requisicoes {
		requisicao := requisicao
		go func() {
			defer wg.Done()
			requisicao()
		}()
	}
	wg.Wait()

	resultados := make([]ResultadoEstudo, 0, len(estudos))
	for _, estudoID := range estudos {
		gasto := orcamento.GastoAtual(estudoID)
		limite := orcamento.Limite(estudoID)
		resultados = append(resultados, ResultadoEstudo{
			EstudoID: estudoID,
			Gasto:    gasto,
			Limite:   limite,
			Estourou: gasto > limite+1e-9,
		})
	}
	return resultados
}
