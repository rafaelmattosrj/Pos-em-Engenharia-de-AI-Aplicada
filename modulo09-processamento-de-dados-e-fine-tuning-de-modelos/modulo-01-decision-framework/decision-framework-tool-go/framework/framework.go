// Package framework implementa o gate de 4 perguntas com score de confiança
// ponderado por AHP, orquestrado atrás do gate de governança -- equivalente a
// avaliarFramework/avaliarCasoCompleto em decision-framework-tool.js.
package framework

import (
	"math"

	"decision-framework-tool/config"
	"decision-framework-tool/governance"
)

// Recomendacao é a mensagem final apresentada ao usuário.
type Recomendacao string

const (
	FineTuning           Recomendacao = "Fine-tuning vale a pena"
	ContinuarPromptRAG   Recomendacao = "Continue com prompt + RAG. Fine-tuning ainda não."
	Esperar              Recomendacao = "Espere acumular dado, depois treine."
	BloqueadoPorGovernanca Recomendacao = "Bloqueado: resolva a governança do dado antes de reavaliar."
)

// ChavesPerguntas é a ordem fixa das 4 perguntas do framework.
var ChavesPerguntas = []string{"p1", "p2", "p3", "p4"}

// SinalPergunta é o veredito individual de uma pergunta.
type SinalPergunta struct {
	Score float64
	Sinal string // "VERDE" ou "VERMELHO"
}

// Resultado é o veredito completo do pipeline (governança + gate de 4 perguntas).
type Resultado struct {
	BloqueadoPorGovernanca bool
	MotivosGovernanca      []string
	Aprovado               bool
	FalhaSoDado            bool
	DecisaoTecnicaEmAberto bool
	Recomendacao           Recomendacao
	PerguntasFalhas        []int
	ScoreComposto          *float64
	SinaisPorPergunta      map[string]SinalPergunta
}

// AvaliarFramework é o gate de 4 perguntas isolado, sem o gate de governança --
// usado pelas demos de AHP de comitê.
func AvaliarFramework(scores map[string]float64, pesos []float64, limiarVerde float64) Resultado {
	sinais := make(map[string]SinalPergunta, len(ChavesPerguntas))
	var perguntasFalhas []int

	for i, chave := range ChavesPerguntas {
		score := scores[chave]
		verde := score >= limiarVerde
		sinal := "VERMELHO"
		if verde {
			sinal = "VERDE"
		}
		sinais[chave] = SinalPergunta{Score: score, Sinal: sinal}
		if !verde {
			perguntasFalhas = append(perguntasFalhas, i+1)
		}
	}

	scoreComposto := 0.0
	for i, chave := range ChavesPerguntas {
		scoreComposto += scores[chave] * pesos[i]
	}

	aprovado := len(perguntasFalhas) == 0
	// "só falha por dado" -- a única reprovação que Real Options resolve
	falhaSoDado := len(perguntasFalhas) == 1 && perguntasFalhas[0] == 3

	recomendacao := ContinuarPromptRAG
	if aprovado {
		recomendacao = FineTuning
	}

	arredondado := roundTo(scoreComposto, 4)
	return Resultado{
		Aprovado:               aprovado,
		FalhaSoDado:            falhaSoDado,
		DecisaoTecnicaEmAberto: aprovado, // aprovado no gate abre a escolha de técnica (LoRA/full/API), Módulos 3 e 4
		Recomendacao:           recomendacao,
		PerguntasFalhas:        perguntasFalhas,
		ScoreComposto:          &arredondado,
		SinaisPorPergunta:      sinais,
	}
}

// AvaliarCasoCompleto orquestra o pipeline completo: governança primeiro
// (bloqueador, grátis), só entra no gate de 4 perguntas se a governança aprovar.
func AvaliarCasoCompleto(caso config.Caso, pesos []float64, limiarVerde float64) Resultado {
	g := governance.Validar(caso.Governanca)
	if !g.Aprovado {
		return Resultado{
			BloqueadoPorGovernanca: true,
			MotivosGovernanca:      g.Motivos,
			Recomendacao:           BloqueadoPorGovernanca,
		}
	}
	r := AvaliarFramework(caso.Scores, pesos, limiarVerde)
	r.BloqueadoPorGovernanca = false
	return r
}

func roundTo(v float64, casas int) float64 {
	fator := math.Pow(10, float64(casas))
	return math.Round(v*fator) / fator
}
