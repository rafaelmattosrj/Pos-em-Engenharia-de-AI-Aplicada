// Package decisionframework e' o subconjunto do gate de 4 perguntas + AHP
// (Modulo 1.2/1.3), necessario so pra reavaliar o caso Amplitude Saude
// Empresarial (Modulo 3.2).
//
// ADAPTACAO: porte AUTOCONTIDO de
// modulo-01-decision-framework/decision-framework-tool.js (funcoes
// derivarPesosAHP, avaliarFramework, RECOMENDACAO, PERGUNTAS,
// CHAVES_PERGUNTAS, carregarConfiguracao). O original em JS importa essas
// funcoes diretamente do arquivo do Modulo 1 via require(); como este repo
// nao tem um mecanismo de modulo compartilhado entre projetos Maven/Go
// independentes, a logica foi duplicada aqui (mesmo dado de entrada,
// amplitude-seguros-casos.json, embutido via go:embed).
package decisionframework

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
)

//go:embed amplitude-seguros-casos.json
var configJSON []byte

const (
	RecomendacaoFineTuning          = "Fine-tuning vale a pena"
	RecomendacaoContinuarPromptRag  = "Continue com prompt + RAG. Fine-tuning ainda não."
	RecomendacaoEsperar             = "Espere acumular dado, depois treine."
	RecomendacaoBloqueadoGovernanca = "Bloqueado: resolva a governança do dado antes de reavaliar."
)

var ChavesPerguntas = []string{"p1", "p2", "p3", "p4"}

var Perguntas = map[string]string{
	"p1": "A tarefa é estreita e repetida, ou aberta e variável?",
	"p2": "Já esgotou prompt engineering + RAG + roteamento, sem chegar na qualidade/custo/latência necessários?",
	"p3": "Tem dado de exemplo suficiente, diverso e de qualidade pra treinar?",
	"p4": "A tarefa é estável o bastante pra não virar esteira de retreino constante?",
}

func DerivarPesosAHP(matriz [][]float64) []float64 {
	n := len(matriz)
	mediasGeometricas := make([]float64, n)
	for i, linha := range matriz {
		produto := 1.0
		for _, v := range linha {
			produto *= v
		}
		mediasGeometricas[i] = math.Pow(produto, 1.0/float64(n))
	}
	soma := 0.0
	for _, v := range mediasGeometricas {
		soma += v
	}
	pesos := make([]float64, n)
	for i, v := range mediasGeometricas {
		pesos[i] = v / soma
	}
	return pesos
}

type Sinal struct {
	Score float64
	Sinal string
}

type ResultadoAvaliacao struct {
	Aprovado          bool
	FalhaSoDado       bool
	Recomendacao      string
	PerguntasFalhas   []int
	ScoreComposto     float64
	SinaisPorPergunta map[string]Sinal
}

func AvaliarFramework(scores map[string]float64, pesos []float64, limiarVerde float64) ResultadoAvaliacao {
	sinais := make(map[string]Sinal, len(ChavesPerguntas))
	var falhas []int
	for i, chave := range ChavesPerguntas {
		score := scores[chave]
		verde := score >= limiarVerde
		sinalTexto := "VERMELHO"
		if verde {
			sinalTexto = "VERDE"
		}
		sinais[chave] = Sinal{Score: score, Sinal: sinalTexto}
		if !verde {
			falhas = append(falhas, i+1)
		}
	}

	scoreComposto := 0.0
	for i, chave := range ChavesPerguntas {
		scoreComposto += scores[chave] * pesos[i]
	}

	aprovado := len(falhas) == 0
	falhaSoDado := len(falhas) == 1 && falhas[0] == 3
	recomendacao := RecomendacaoContinuarPromptRag
	if aprovado {
		recomendacao = RecomendacaoFineTuning
	}

	return ResultadoAvaliacao{
		Aprovado: aprovado, FalhaSoDado: falhaSoDado, Recomendacao: recomendacao,
		PerguntasFalhas: falhas, ScoreComposto: math.Round(scoreComposto*10000) / 10000, SinaisPorPergunta: sinais,
	}
}

type OpcaoReal struct {
	CustoDeErroEsperadoPorChamada float64
	TaxaCrescimentoScorePorMes    float64
	ScoreAlvo                     float64
}

type Financeiro struct {
	VolumeInicialMensal  int
	CrescimentoMensalModa float64
	OpcaoReal            *OpcaoReal
}

type Caso struct {
	ID         string
	Scores     map[string]float64
	Financeiro *Financeiro
}

type Configuracao struct {
	LimiarVerde float64
	MatrizAHP   [][]float64
	Casos       []Caso
}

func (c Configuracao) Caso(id string) (Caso, error) {
	for _, caso := range c.Casos {
		if caso.ID == id {
			return caso, nil
		}
	}
	return Caso{}, fmt.Errorf("caso nao encontrado: %s", id)
}

type configBruta struct {
	LimiarVerde float64 `json:"limiarVerde"`
	Ahp         struct {
		Matriz [][]float64 `json:"matriz"`
	} `json:"ahp"`
	Casos []struct {
		ID     string             `json:"id"`
		Scores map[string]float64 `json:"scores"`
		Financeiro *struct {
			VolumeInicialMensal int `json:"volumeInicialMensal"`
			CrescimentoMensal   struct {
				Moda float64 `json:"moda"`
			} `json:"crescimentoMensal"`
			OpcaoReal *struct {
				CustoDeErroEsperadoPorChamada float64 `json:"custoDeErroEsperadoPorChamada"`
				TaxaCrescimentoScorePorMes    float64 `json:"taxaCrescimentoScorePorMes"`
				ScoreAlvo                     float64 `json:"scoreAlvo"`
			} `json:"opcaoReal"`
		} `json:"financeiro"`
	} `json:"casos"`
}

func CarregarConfiguracao() (Configuracao, error) {
	var bruta configBruta
	if err := json.Unmarshal(configJSON, &bruta); err != nil {
		return Configuracao{}, err
	}

	casos := make([]Caso, 0, len(bruta.Casos))
	for _, c := range bruta.Casos {
		var financeiro *Financeiro
		if c.Financeiro != nil {
			var opcaoReal *OpcaoReal
			if c.Financeiro.OpcaoReal != nil {
				opcaoReal = &OpcaoReal{
					CustoDeErroEsperadoPorChamada: c.Financeiro.OpcaoReal.CustoDeErroEsperadoPorChamada,
					TaxaCrescimentoScorePorMes:    c.Financeiro.OpcaoReal.TaxaCrescimentoScorePorMes,
					ScoreAlvo:                     c.Financeiro.OpcaoReal.ScoreAlvo,
				}
			}
			financeiro = &Financeiro{
				VolumeInicialMensal:   c.Financeiro.VolumeInicialMensal,
				CrescimentoMensalModa: c.Financeiro.CrescimentoMensal.Moda,
				OpcaoReal:             opcaoReal,
			}
		}
		casos = append(casos, Caso{ID: c.ID, Scores: c.Scores, Financeiro: financeiro})
	}

	return Configuracao{LimiarVerde: bruta.LimiarVerde, MatrizAHP: bruta.Ahp.Matriz, Casos: casos}, nil
}
