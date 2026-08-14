// Package ollama é um cliente mínimo para a API NATIVA do Ollama
// (http://localhost:11434/api/...) — usa os mesmos dois endpoints que os
// pacotes `ollama` do npm e do PyPI usam por baixo dos panos:
// POST /api/embeddings (embedding de um texto) e POST /api/chat (chat, com
// streaming NDJSON quando stream=true). Diferente de
// ollama-local-llm-chat-go (que fala com a API OpenAI-compatible em /v1,
// só chat, sem embeddings), este gateway precisa de embeddings reais e de
// controle fino sobre o streaming — por isso fala direto com a API nativa.
package ollama

import (
	"fmt"
	"time"
)

const (
	maxTentativasPadrao = 3                // MAX_TENTATIVAS_OLLAMA nos dois originais
	timeoutPadrao       = 20 * time.Second // TIMEOUT_OLLAMA_MS / TIMEOUT_OLLAMA_S
)

// ErroDeTimeout replica ErroDeTimeoutOllama (JS) / ErroDeTimeoutOllama
// (Python): o Ollama não respondeu dentro do timeout configurado.
type ErroDeTimeout struct {
	Operacao string
	Timeout  time.Duration
}

func (e *ErroDeTimeout) Error() string {
	return fmt.Sprintf("%s: timeout de %s excedido — Ollama não respondeu a tempo.", e.Operacao, e.Timeout)
}

// comTimeout executa fn numa goroutine e devolve seu resultado, ou um
// ErroDeTimeout se `timeout` esgotar primeiro — mesma receita do
// Promise.race usado em comTimeout() no JS e do
// ThreadPoolExecutor(...).result(timeout=...) usado em com_retry() no
// Python. Limitação honesta idêntica à da versão Python: isso desiste de
// ESPERAR a resposta, não cancela a chamada em si — Go não tem como matar
// uma goroutine à força; se o Ollama travar de verdade, a goroutine só
// termina quando (ou se) ele eventualmente responder.
func comTimeout[T any](fn func() (T, error), timeout time.Duration, operacao string) (T, error) {
	type resultado struct {
		valor T
		err   error
	}
	ch := make(chan resultado, 1)
	go func() {
		valor, err := fn()
		ch <- resultado{valor, err}
	}()
	select {
	case r := <-ch:
		return r.valor, r.err
	case <-time.After(timeout):
		var zero T
		return zero, &ErroDeTimeout{Operacao: operacao, Timeout: timeout}
	}
}

// comRetry chama fn até maxTentativas vezes, aplicando comTimeout em cada
// tentativa — mesmo padrão de retry com limite do Módulo 3.5, aqui aplicado
// às duas chamadas ao Ollama (embedding e geração) que antes não tinham
// nenhuma proteção contra travamento ou instabilidade momentânea do modelo local.
func comRetry[T any](fn func() (T, error), maxTentativas int, timeout time.Duration, operacao string, logf func(format string, args ...interface{})) (T, error) {
	var ultimoErro error
	for tentativa := 1; tentativa <= maxTentativas; tentativa++ {
		valor, err := comTimeout(fn, timeout, operacao)
		if err == nil {
			return valor, nil
		}
		ultimoErro = err
		if logf != nil {
			logf("[Retry] %s falhou na tentativa %d/%d: %s", operacao, tentativa, maxTentativas, err)
		}
	}
	var zero T
	return zero, fmt.Errorf("%s falhou após %d tentativas: %w", operacao, maxTentativas, ultimoErro)
}
