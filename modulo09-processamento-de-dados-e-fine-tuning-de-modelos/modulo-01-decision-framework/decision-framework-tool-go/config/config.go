// Package config carrega amplitude-seguros-casos.json (embutido no binário via
// go:embed) -- equivalente a carregarConfiguracao() em decision-framework-tool.js.
package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed amplitude-seguros-casos.json
var casosJSON []byte

// Triangular é uma distribuição triangular {min, moda, max}.
type Triangular struct {
	Min  float64 `json:"min"`
	Moda float64 `json:"moda"`
	Max  float64 `json:"max"`
}

// OpcaoReal só está presente quando o caso é elegível a Real Options (reprovação
// exclusivamente por dado insuficiente, pergunta p3).
type OpcaoReal struct {
	CustoDeErroEsperadoPorChamada float64 `json:"custoDeErroEsperadoPorChamada"`
	TaxaCrescimentoScorePorMes    float64 `json:"taxaCrescimentoScorePorMes"`
	ScoreAlvo                     float64 `json:"scoreAlvo"`
}

// Financeiro é o bloco de parâmetros de NPV/DCF, Monte Carlo e (opcional) Real Options.
type Financeiro struct {
	VolumeInicialMensal       float64     `json:"volumeInicialMensal"`
	CrescimentoMensal         Triangular  `json:"crescimentoMensal"`
	CustoPorChamadaStatusQuo  Triangular  `json:"custoPorChamadaStatusQuo"`
	CustoPorChamadaFineTuned  Triangular  `json:"custoPorChamadaFineTuned"`
	CustoTreinamento          float64     `json:"custoTreinamento"`
	HorizonteMeses            int         `json:"horizonteMeses"`
	TaxaDescontoMensal        float64     `json:"taxaDescontoMensal"`
	OpcaoReal                 *OpcaoReal  `json:"opcaoReal,omitempty"`
}

// Governanca é o gate binário de governança de dado (LGPD).
type Governanca struct {
	DadoSensivelLGPD   bool   `json:"dadoSensivelLGPD"`
	BaseLegalDefinida  bool   `json:"baseLegalDefinida"`
	BaseLegalDescricao string `json:"baseLegalDescricao"`
	DpaAssinado        bool   `json:"dpaAssinado"`
}

// Caso é um caso de negócio da Amplitude Seguros.
type Caso struct {
	ID         string             `json:"id"`
	Nome       string             `json:"nome"`
	Tarefa     string             `json:"tarefa"`
	Scores     map[string]float64 `json:"scores"`
	Governanca Governanca         `json:"governanca"`
	Financeiro *Financeiro        `json:"financeiro,omitempty"`
}

// AhpConfig é a matriz de comparação pareada (escala de Saaty) e a ordem das perguntas que ela representa.
type AhpConfig struct {
	OrdemPerguntas []string    `json:"ordemPerguntas"`
	Matriz         [][]float64 `json:"matriz"`
}

// AmplitudeConfig é a raiz de amplitude-seguros-casos.json.
type AmplitudeConfig struct {
	LimiarVerde float64   `json:"limiarVerde"`
	Ahp         AhpConfig `json:"ahp"`
	Casos       []Caso    `json:"casos"`
}

// Caso busca um caso pelo id -- equivalente a `casos.find(c => c.id === id)`.
func (c *AmplitudeConfig) Caso(id string) (Caso, error) {
	for _, caso := range c.Casos {
		if caso.ID == id {
			return caso, nil
		}
	}
	return Caso{}, fmt.Errorf("caso nao encontrado: %s", id)
}

// Carregar decodifica o JSON embutido no binário.
func Carregar() (*AmplitudeConfig, error) {
	var cfg AmplitudeConfig
	if err := json.Unmarshal(casosJSON, &cfg); err != nil {
		return nil, fmt.Errorf("config: falha ao decodificar amplitude-seguros-casos.json: %w", err)
	}
	return &cfg, nil
}
