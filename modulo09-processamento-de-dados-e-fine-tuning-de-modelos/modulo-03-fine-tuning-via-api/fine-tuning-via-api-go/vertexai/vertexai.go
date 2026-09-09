// Package vertexai fala com a API REST da Vertex AI (aiplatform.googleapis.com)
// pra consultar e criar job de fine-tuning. So o HTTPClient real (JobClient)
// toca rede/processo externo -- tudo o mais no projeto e logica pura,
// testavel com uma implementacao falsa de JobClient.
package vertexai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// JobClient e' o contrato injetavel (mesmo principio de consultarFn/criarJobFn
// injetaveis nos originais JS): a implementacao real faz chamada de rede de
// verdade, a de teste devolve estado fixo sem tocar rede.
type JobClient interface {
	ConsultarJob(nomeJob string) (map[string]any, error)
	CriarJob(corpo map[string]any) (map[string]any, error)
}

type HTTPClient struct {
	Projeto string
	Regiao  string
	http    *http.Client
}

func NewHTTPClient(projeto, regiao string) *HTTPClient {
	return &HTTPClient{Projeto: projeto, Regiao: regiao, http: &http.Client{}}
}

func (c *HTTPClient) obterTokenAcesso() (string, error) {
	out, err := exec.Command("gcloud", "auth", "print-access-token").Output()
	if err != nil {
		return "", fmt.Errorf("gcloud auth print-access-token falhou: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *HTTPClient) ConsultarJob(nomeJob string) (map[string]any, error) {
	token, err := c.obterTokenAcesso()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", c.Regiao, nomeJob)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("falha ao consultar job: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var job map[string]any
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *HTTPClient) CriarJob(corpo map[string]any) (map[string]any, error) {
	token, err := c.obterTokenAcesso()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/tuningJobs", c.Regiao, c.Projeto, c.Regiao)
	corpoJSON, err := json.Marshal(corpo)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(corpoJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("falha ao criar job: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var job map[string]any
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, err
	}
	return job, nil
}
