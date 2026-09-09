// Package managedapi porta lora-managed-api-preview-tool.js (Modulo 4.2,
// companion). Monta a MESMA requisicao HTTP real que a API de fine-tuning da
// Together AI espera para um job LoRA (config reaproveitada do treino local
// do M4.2: rank 8, scale 20.0, dropout 0.0). So envia de verdade se
// TOGETHER_API_KEY estiver no ambiente; sem a chave, so monta e devolve a
// requisicao (modo preview), igual ao original.
//
// Ressalva de honestidade herdada do original: "scale" (MLX-LM) e
// "lora_alpha" (Together AI) nao sao garantidamente definidos de forma
// identica entre frameworks -- o mesmo valor numerico e usado aqui como ponte
// ilustrativa, nao equivalencia matematica comprovada.
package managedapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	TogetherEndpoint = "https://api.together.ai/v1/fine-tunes"
	ModeloPadrao     = "meta-llama/Meta-Llama-3.1-8B-Instruct-Reference"
)

type ConfigLora struct {
	Rank    int
	Dropout float64
	Scale   float64
}

var ConfigTreinadaLocal = ConfigLora{Rank: 8, Dropout: 0.0, Scale: 20.0}

type trainingType struct {
	Type              string  `json:"type"`
	LoraR             int     `json:"lora_r"`
	LoraAlpha         float64 `json:"lora_alpha"`
	LoraDropout       float64 `json:"lora_dropout"`
	LoraTrainableMods string  `json:"lora_trainable_modules"`
}

type requestBody struct {
	Model        string       `json:"model"`
	TrainingFile string       `json:"training_file"`
	TrainingType trainingType `json:"training_type"`
}

type Requisicao struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    requestBody
}

func MontarRequisicaoLoraGerenciada(apiKey, trainingFileID, modelo string, loraConfig ConfigLora) (Requisicao, error) {
	if trainingFileID == "" {
		return Requisicao{}, errors.New("trainingFileId é obrigatório (id do arquivo já enviado à API)")
	}

	authKey := "<TOGETHER_API_KEY>"
	if apiKey != "" {
		authKey = apiKey
	}

	return Requisicao{
		URL:    TogetherEndpoint,
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer " + authKey,
			"Content-Type":  "application/json",
		},
		Body: requestBody{
			Model:        modelo,
			TrainingFile: trainingFileID,
			TrainingType: trainingType{
				Type: "Lora", LoraR: loraConfig.Rank, LoraAlpha: loraConfig.Scale,
				LoraDropout: loraConfig.Dropout, LoraTrainableMods: "all-linear",
			},
		},
	}, nil
}

func MontarRequisicaoPadrao(apiKey, trainingFileID string) (Requisicao, error) {
	return MontarRequisicaoLoraGerenciada(apiKey, trainingFileID, ModeloPadrao, ConfigTreinadaLocal)
}

type Resultado struct {
	Modo       string // "preview" ou "real"
	Requisicao Requisicao
	Resposta   string
}

func EnviarOuPrever(trainingFileID string) (Resultado, error) {
	apiKey := os.Getenv("TOGETHER_API_KEY")
	requisicao, err := MontarRequisicaoLoraGerenciada(apiKey, trainingFileID, ModeloPadrao, ConfigTreinadaLocal)
	if err != nil {
		return Resultado{}, err
	}

	if apiKey == "" {
		return Resultado{Modo: "preview", Requisicao: requisicao}, nil
	}

	corpo, err := json.Marshal(requisicao.Body)
	if err != nil {
		return Resultado{}, err
	}
	req, err := http.NewRequest(requisicao.Method, requisicao.URL, bytes.NewReader(corpo))
	if err != nil {
		return Resultado{}, err
	}
	for chave, valor := range requisicao.Headers {
		req.Header.Set(chave, valor)
	}

	resposta, err := http.DefaultClient.Do(req)
	if err != nil {
		return Resultado{}, err
	}
	defer resposta.Body.Close()
	respostaCorpo, err := io.ReadAll(resposta.Body)
	if err != nil {
		return Resultado{}, err
	}
	if resposta.StatusCode < 200 || resposta.StatusCode >= 300 {
		return Resultado{}, fmt.Errorf("Together AI recusou a requisição (%d): %s", resposta.StatusCode, respostaCorpo)
	}
	return Resultado{Modo: "real", Requisicao: requisicao, Resposta: string(respostaCorpo)}, nil
}
