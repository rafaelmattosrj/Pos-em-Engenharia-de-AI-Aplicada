// Package automation implementa o processo completo de fine-tuning via API
// (Modulo 3.4): upload, criacao de job com trava de confirmacao explicita
// (incidente real do Modulo 3.3: hiperparametro invalido nao gera erro
// rapido na Vertex AI), acompanhamento sozinho ate o job terminar, com
// backoff exponencial e retry limitado por falha transiente de rede.
package automation

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var estadosTerminais = map[string]bool{
	"JOB_STATE_SUCCEEDED": true, "JOB_STATE_FAILED": true, "JOB_STATE_CANCELLED": true,
}

func MontarComandoUpload(caminhoLocal, uriGcs string) (string, error) {
	if !strings.HasSuffix(caminhoLocal, ".jsonl") {
		return "", errors.New("dataset precisa ser .jsonl")
	}
	if !strings.HasPrefix(uriGcs, "gs://") {
		return "", errors.New("destino precisa ser um URI gs://")
	}
	return fmt.Sprintf(`gsutil cp "%s" "%s"`, caminhoLocal, uriGcs), nil
}

func ExigirConfirmacao(confirmar bool) error {
	if !confirmar {
		return errors.New("criarJobFineTuning bloqueado: passe confirmar=true explicitamente pra criar job de verdade. " +
			"Trava adicionada depois do incidente do Módulo 3.3, onde hiperparâmetro inválido criou job real sem aviso.")
	}
	return nil
}

func CalcularProximoIntervalo(atualMs int64, fator float64, maximoMs int64) int64 {
	proximo := int64(math.Round(float64(atualMs) * fator))
	if proximo > maximoMs {
		return maximoMs
	}
	return proximo
}

type ConsultarFn func(nomeJob string) (map[string]any, error)
type EsperarFn func(d time.Duration)

var EsperarReal EsperarFn = time.Sleep

// ConsultarComRetry absorve falha transiente de rede: uma automacao rodando
// sozinha por ~1h40min nao pode cair por causa de um timeout isolado.
func ConsultarComRetry(consultarFn ConsultarFn, nomeJob string, tentativas int, atraso time.Duration, esperarFn EsperarFn) (map[string]any, error) {
	var ultimoErro error
	for tentativa := 1; tentativa <= tentativas; tentativa++ {
		job, err := consultarFn(nomeJob)
		if err == nil {
			return job, nil
		}
		ultimoErro = err
		if tentativa < tentativas {
			esperarFn(atraso)
		}
	}
	return nil, ultimoErro
}

type OpcoesAcompanhamento struct {
	IntervaloInicial  time.Duration
	FatorBackoff      float64
	IntervaloMaximo   time.Duration
	AoAtualizar       func(job map[string]any)
	TentativasConsulta int
	AtrasoRetry       time.Duration
}

func OpcoesPadrao() OpcoesAcompanhamento {
	return OpcoesAcompanhamento{
		IntervaloInicial:   5 * time.Second,
		FatorBackoff:       1.5,
		IntervaloMaximo:    60 * time.Second,
		AoAtualizar:        func(map[string]any) {},
		TentativasConsulta: 3,
		AtrasoRetry:        3 * time.Second,
	}
}

// AcompanharAteFinalizar consulta repetidamente ate o job chegar num estado
// terminal, com backoff exponencial entre tentativas.
func AcompanharAteFinalizar(nomeJob string, consultarFn ConsultarFn, esperarFn EsperarFn, opcoes OpcoesAcompanhamento) (map[string]any, error) {
	intervaloMs := opcoes.IntervaloInicial.Milliseconds()
	job, err := ConsultarComRetry(consultarFn, nomeJob, opcoes.TentativasConsulta, opcoes.AtrasoRetry, esperarFn)
	if err != nil {
		return nil, err
	}
	opcoes.AoAtualizar(job)

	for !estadosTerminais[fmt.Sprintf("%v", job["state"])] {
		esperarFn(time.Duration(intervaloMs) * time.Millisecond)
		intervaloMs = CalcularProximoIntervalo(intervaloMs, opcoes.FatorBackoff, opcoes.IntervaloMaximo.Milliseconds())
		job, err = ConsultarComRetry(consultarFn, nomeJob, opcoes.TentativasConsulta, opcoes.AtrasoRetry, esperarFn)
		if err != nil {
			return nil, err
		}
		opcoes.AoAtualizar(job)
	}
	return job, nil
}
