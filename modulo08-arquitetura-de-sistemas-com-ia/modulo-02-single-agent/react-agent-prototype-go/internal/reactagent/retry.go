package reactagent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ModeloHTTPError e uma resposta HTTP nao-200 do servidor do modelo
// (Ollama). Carrega o status code para que a politica de retry possa
// distinguir erro de configuracao (ex.: 404, modelo nao encontrado) de
// falha transitoria de rede.
type ModeloHTTPError struct {
	StatusCode int
	Body       string
}

func (e *ModeloHTTPError) Error() string {
	return fmt.Sprintf("ollama: erro da API (status %d): %s", e.StatusCode, e.Body)
}

// EhErroTransitorio distingue falha transitoria (rede, timeout — vale
// tentar de novo) de falha terminal (ex.: modelo nao existe no Ollama —
// retry nao resolve, e erro de configuracao).
//
// Adaptacao: no original, erro.status_code === 404 e o unico caso
// explicitamente terminal; qualquer outro codigo HTTP cairia no teste
// generico de mensagem (que nunca bate pra um erro puramente de status).
// Aqui isso vira "qualquer *ModeloHTTPError e terminal" — mesmo resultado
// observavel (nenhum status HTTP e tratado como transitorio), mais direto
// de expressar em Go.
//
// Porte 1:1 de ehErroTransitorio em react-agent-prototype.js / .py.
func EhErroTransitorio(err error) bool {
	var httpErro *ModeloHTTPError
	if errors.As(err, &httpErro) {
		return false // erro de configuração/servidor: retry não resolve
	}
	mensagem := strings.ToLower(err.Error())
	return strings.Contains(mensagem, "timeout") ||
		strings.Contains(mensagem, "connection refused") ||
		strings.Contains(mensagem, "connection reset") ||
		strings.Contains(mensagem, "econnrefused") ||
		strings.Contains(mensagem, "fetch failed")
}

// RetryingChatClient decora um ChatClient com retry + backoff exponencial
// (chamada ao modelo, nao a ferramenta). Sem isso, qualquer soneca de rede
// escalava pro Approval Gate sem necessidade.
//
// Porte 1:1 de chamarModeloComRetry em react-agent-prototype.js / .py.
type RetryingChatClient struct {
	Delegate      ChatClient
	TentativasMax int
}

// NewRetryingChatClient cria um RetryingChatClient com o default de 3
// tentativas usado no original.
func NewRetryingChatClient(delegate ChatClient) *RetryingChatClient {
	return &RetryingChatClient{Delegate: delegate, TentativasMax: 3}
}

func (c *RetryingChatClient) Chat(ctx context.Context, historico []Mensagem, tools []Mensagem) (ModeloResposta, error) {
	tentativasMax := c.TentativasMax
	if tentativasMax <= 0 {
		tentativasMax = 3
	}

	var ultimoErro error
	for tentativa := 1; tentativa <= tentativasMax; tentativa++ {
		resposta, err := c.Delegate.Chat(ctx, historico, tools)
		if err == nil {
			return resposta, nil
		}
		ultimoErro = err
		if !EhErroTransitorio(err) || tentativa == tentativasMax {
			return ModeloResposta{}, err // erro terminal, ou já esgotou as tentativas
		}

		esperaMs := 500 * (1 << (tentativa - 1)) // 500ms, 1s, 2s...
		fmt.Printf("[Retry] Falha transitória na chamada ao modelo (tentativa %d/%d): %v. Tentando de novo em %dms.\n",
			tentativa, tentativasMax, err, esperaMs)

		select {
		case <-time.After(time.Duration(esperaMs) * time.Millisecond):
		case <-ctx.Done():
			return ModeloResposta{}, ctx.Err()
		}
	}

	return ModeloResposta{}, ultimoErro
}
