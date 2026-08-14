package reactagent

import (
	"context"
	"errors"
	"testing"
)

// Cobre a mesma disciplina de ehErroTransitorio / chamarModeloComRetry em
// react-agent-prototype.js / .py: falha transitória tenta de novo (com
// backoff), falha terminal (ex.: 404 / modelo não encontrado) sobe direto.

func respostaFinalFake(texto string) ModeloResposta {
	return ModeloResposta{
		MensagemBruta: Mensagem{"role": "assistant", "content": texto},
		Content:       texto,
	}
}

func TestEhErroTransitorio_ConexaoRecusada_EhTransitorio(t *testing.T) {
	if !EhErroTransitorio(errors.New("dial tcp: connection refused")) {
		t.Error("esperava erro de conexão recusada como transitório")
	}
}

func TestEhErroTransitorio_MensagemDeTimeout_EhTransitorio(t *testing.T) {
	if !EhErroTransitorio(errors.New("context deadline exceeded (timeout)")) {
		t.Error("esperava erro de timeout como transitório")
	}
}

func TestEhErroTransitorio_ErroHttp404_NaoEhTransitorio(t *testing.T) {
	if EhErroTransitorio(&ModeloHTTPError{StatusCode: 404, Body: "model not found"}) {
		t.Error("esperava erro HTTP 404 como não-transitório")
	}
}

func TestEhErroTransitorio_ErroHttpGenerico_NaoEhTransitorio(t *testing.T) {
	if EhErroTransitorio(&ModeloHTTPError{StatusCode: 500, Body: "internal error"}) {
		t.Error("esperava erro HTTP 500 como não-transitório")
	}
}

func TestEhErroTransitorio_MensagemSemPadraoConhecido_NaoEhTransitorio(t *testing.T) {
	if EhErroTransitorio(errors.New("algo inesperado aconteceu")) {
		t.Error("esperava mensagem desconhecida como não-transitória")
	}
}

func TestRetryingChatClient_FalhaTransitoriaSeguidaDeSucesso_Retenta(t *testing.T) {
	fake := (&fakeChatClient{}).
		comFalha(errors.New("connection refused")).
		comResposta(respostaFinalFake("ok na segunda tentativa"))

	client := &RetryingChatClient{Delegate: fake, TentativasMax: 3}
	resposta, err := client.Chat(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("não esperava erro, obteve %v", err)
	}
	if resposta.Content != "ok na segunda tentativa" {
		t.Errorf("conteúdo inesperado: %q", resposta.Content)
	}
	if fake.Chamadas != 2 {
		t.Errorf("esperava 2 chamadas, obteve %d", fake.Chamadas)
	}
}

func TestRetryingChatClient_ErroTerminal_NaoRetenta(t *testing.T) {
	fake := (&fakeChatClient{}).comFalha(&ModeloHTTPError{StatusCode: 404, Body: "model not found"})

	client := &RetryingChatClient{Delegate: fake, TentativasMax: 3}
	_, err := client.Chat(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("esperava erro terminal propagado")
	}
	if fake.Chamadas != 1 {
		t.Errorf("esperava 1 chamada (sem retry), obteve %d", fake.Chamadas)
	}
}

func TestRetryingChatClient_EsgotaTentativas_PropagaUltimoErro(t *testing.T) {
	fake := (&fakeChatClient{}).
		comFalha(errors.New("connection refused")).
		comFalha(errors.New("connection refused")).
		comFalha(errors.New("connection refused"))

	client := &RetryingChatClient{Delegate: fake, TentativasMax: 3}
	_, err := client.Chat(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("esperava erro propagado após esgotar tentativas")
	}
	if fake.Chamadas != 3 {
		t.Errorf("esperava 3 chamadas, obteve %d", fake.Chamadas)
	}
}
