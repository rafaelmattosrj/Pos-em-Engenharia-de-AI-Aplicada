// Package client fala HTTP com a legacy customer API — equivalente a
// CustomerHttpClient.java.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mcp-customers-server/model"
)

// CustomerHttpClient é responsável por toda comunicação HTTP com a legacy
// customer API. Autentica via Bearer token.
type CustomerHttpClient struct {
	BaseURL      string
	ServiceToken string
	HTTPClient   *http.Client
}

// NewCustomerHttpClient cria um CustomerHttpClient com timeout de 30s.
func NewCustomerHttpClient(baseURL, serviceToken string) *CustomerHttpClient {
	return &CustomerHttpClient{
		BaseURL:      baseURL,
		ServiceToken: serviceToken,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ListAll lista todos os clientes cadastrados na API legada.
func (c *CustomerHttpClient) ListAll(ctx context.Context) ([]model.Customer, error) {
	var customers []model.Customer
	if err := c.doJSON(ctx, http.MethodGet, "/customers", nil, &customers); err != nil {
		return nil, fmt.Errorf("falha ao listar clientes da API legada: %w", err)
	}
	return customers, nil
}

// FindByID busca um cliente específico pelo ID. Retorna nil se não
// encontrado ou se ocorrer qualquer erro (mesmo comportamento "silencioso"
// de CustomerHttpClient.findById, que apenas loga e retorna null).
func (c *CustomerHttpClient) FindByID(ctx context.Context, id string) *model.Customer {
	var customer model.Customer
	if err := c.doJSON(ctx, http.MethodGet, "/customers/"+id, nil, &customer); err != nil {
		return nil
	}
	return &customer
}

// Create cria um novo cliente na API legada.
func (c *CustomerHttpClient) Create(ctx context.Context, name, phone string) model.MutationResult {
	payload := map[string]string{"name": name, "phone": phone}

	var created model.Customer
	if err := c.doJSON(ctx, http.MethodPost, "/customers", payload, &created); err != nil {
		return model.MutationResult{Success: false, Message: "Falha ao criar cliente: " + err.Error()}
	}
	return model.MutationResult{Success: true, Message: "Cliente criado com sucesso", Customer: &created}
}

// Update atualiza os dados de um cliente existente. name/phone vazios não
// são enviados no corpo (equivalente ao payload parcial da versão Java).
func (c *CustomerHttpClient) Update(ctx context.Context, id, name, phone string) model.MutationResult {
	payload := map[string]string{}
	if name != "" {
		payload["name"] = name
	}
	if phone != "" {
		payload["phone"] = phone
	}
	if len(payload) == 0 {
		return model.MutationResult{Success: false, Message: "Nenhum campo fornecido para atualização"}
	}

	var updated model.Customer
	if err := c.doJSON(ctx, http.MethodPut, "/customers/"+id, payload, &updated); err != nil {
		return model.MutationResult{Success: false, Message: "Falha ao atualizar cliente: " + err.Error()}
	}
	return model.MutationResult{Success: true, Message: "Cliente atualizado com sucesso", Customer: &updated}
}

// Delete remove um cliente da API legada pelo ID.
func (c *CustomerHttpClient) Delete(ctx context.Context, id string) model.MutationResult {
	if err := c.doJSON(ctx, http.MethodDelete, "/customers/"+id, nil, nil); err != nil {
		return model.MutationResult{Success: false, Message: "Falha ao remover cliente: " + err.Error()}
	}
	return model.MutationResult{Success: true, Message: "Cliente removido com sucesso"}
}

// doJSON executa uma requisição HTTP com corpo JSON opcional e desserializa
// a resposta em out (se não nil).
func (c *CustomerHttpClient) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.ServiceToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
