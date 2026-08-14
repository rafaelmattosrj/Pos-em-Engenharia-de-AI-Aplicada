package gateway

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

// fakeEmbedderPorTexto devolve um vetor fixo por texto exato — permite
// controlar precisamente similaridade de cosseno em cada cenário de teste,
// sem depender de um Ollama real.
type fakeEmbedderPorTexto struct {
	vetores map[string][]float64
	padrao  []float64
}

func (f *fakeEmbedderPorTexto) Embedar(_ context.Context, texto string) ([]float64, error) {
	if v, ok := f.vetores[texto]; ok {
		return v, nil
	}
	return f.padrao, nil
}

// fakeChatStreamer devolve uma resposta fixa (emitida num único chunk) e
// registra o modelo com que foi chamado.
type fakeChatStreamer struct {
	resposta     string
	chamadas     int
	ultimoModelo string
}

func (f *fakeChatStreamer) ChatStream(_ context.Context, model string, _ []Message, onChunk func(string)) error {
	f.chamadas++
	f.ultimoModelo = model
	onChunk(f.resposta)
	return nil
}

// fakeApprover simula o Approval Gate com uma fila de respostas pré-definida,
// sem depender de stdin.
type fakeApprover struct {
	respostas []bool
	chamadas  int
}

func (f *fakeApprover) PedirAprovacaoHumana(_ string) (bool, error) {
	if f.chamadas >= len(f.respostas) {
		return false, nil
	}
	r := f.respostas[f.chamadas]
	f.chamadas++
	return r, nil
}

// montaProcessorDeTeste prepara um Processor com embeddings controlados que
// reproduzem os 5 caminhos do roteiro de demo (main.go / trialforge-gateway-
// prototype.js) sem precisar de um Ollama real: cada cláusula real dos 3
// índices (Indices) recebe um vetor de teste ortogonal por domínio
// (icf~{1,0,0}, protocolo~{0,1,0}, csr~{0,0,1}), e cada pergunta de teste
// recebe o vetor que produz o comportamento esperado daquele cenário.
func montaProcessorDeTeste(t *testing.T, vetoresPergunta map[string][]float64, chat *fakeChatStreamer, approver *fakeApprover, out *bytes.Buffer) *Processor {
	t.Helper()

	vetores := map[string][]float64{
		Indices["icf"][0].Tema:  {1, 0, 0},
		Indices["icf"][0].Texto: {1, 0, 0},
		Indices["icf"][1].Tema:  {0.9, 0.1, 0},
		Indices["icf"][1].Texto: {0.9, 0.1, 0},

		Indices["protocolo"][0].Tema:  {0, 1, 0},
		Indices["protocolo"][0].Texto: {0, 1, 0},
		Indices["protocolo"][1].Tema:  {0, 0.9, 0.1},
		Indices["protocolo"][1].Texto: {0, 0.9, 0.1},

		Indices["csr"][0].Tema:  {0, 0, 1},
		Indices["csr"][0].Texto: {0, 0, 1},
		Indices["csr"][1].Tema:  {0, 0.1, 0.9},
		Indices["csr"][1].Texto: {0, 0.1, 0.9},
	}
	for pergunta, vetor := range vetoresPergunta {
		vetores[pergunta] = vetor
	}

	embedder := &fakeEmbedderPorTexto{vetores: vetores, padrao: []float64{-1, -1, -1}}

	preparados, err := PrepararIndices(context.Background(), embedder, nil)
	if err != nil {
		t.Fatalf("erro ao preparar índices de teste: %v", err)
	}

	return &Processor{
		Embedder:        embedder,
		Chat:            chat,
		Preparados:      preparados,
		Cache:           &SemanticCache{},
		Audit:           NewAuditTrail(filepath.Join(t.TempDir(), "audit-trail.jsonl")),
		Approval:        approver,
		Out:             out,
		ModeloBarato:    "gemma4:e2b",
		ModeloCaro:      "gemma4:latest",
		LimiarCache:     0.75,
		LimiarConfianca: 0.7,
	}
}

const (
	perguntaRotina        = "Quais são as regras de assentimento pra menores nesse estudo?"
	perguntaParafrase     = "O assentimento dos menores de idade é obrigatório nesse estudo?"
	perguntaCSR           = "Preciso da síntese do CSR final desse estudo."
	perguntaTemaDiferente = "Qual o prazo de armazenamento das amostras biológicas coletadas nesse estudo?"
	perguntaProtocolo     = "Qual é o critério de idade mínima pra participar desse estudo?"
)

// #1 do roteiro: confiança alta (embedding igual ao tema da cláusula do
// índice icf) — sem cache (cache vazio), sem Approval Gate, modelo barato.
func TestProcessarRequisicaoRotinaAltaConfianca(t *testing.T) {
	chat := &fakeChatStreamer{resposta: "resposta de rotina"}
	approver := &fakeApprover{}
	var out bytes.Buffer

	p := montaProcessorDeTeste(t, map[string][]float64{
		perguntaRotina: {1, 0, 0}, // == Indices["icf"][0] — confiança 1.0
	}, chat, approver, &out)

	resposta, err := p.ProcessarRequisicao(context.Background(), perguntaRotina)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta == nil || *resposta != "resposta de rotina" {
		t.Fatalf("resposta = %v, esperado 'resposta de rotina'", resposta)
	}
	if chat.chamadas != 1 || chat.ultimoModelo != "gemma4:e2b" {
		t.Errorf("esperado 1 chamada ao modelo barato, obtido %d chamadas, modelo %q", chat.chamadas, chat.ultimoModelo)
	}
	if approver.chamadas != 0 {
		t.Errorf("esperado 0 chamadas ao Approval Gate, obtido %d", approver.chamadas)
	}
}

// #2 do roteiro: paráfrase da #1 — mesmo embedding da pergunta anterior já
// cacheada, deve bater no Semantic Cache e NUNCA chamar o modelo.
func TestProcessarRequisicaoParafraseUsaCache(t *testing.T) {
	chat := &fakeChatStreamer{resposta: "resposta de rotina"}
	approver := &fakeApprover{}
	var out bytes.Buffer

	p := montaProcessorDeTeste(t, map[string][]float64{
		perguntaRotina:    {1, 0, 0},
		perguntaParafrase: {1, 0, 0}, // mesmo vetor -> similaridade 1.0 com o cache
	}, chat, approver, &out)

	ctx := context.Background()
	if _, err := p.ProcessarRequisicao(ctx, perguntaRotina); err != nil {
		t.Fatalf("erro inesperado na #1: %v", err)
	}
	chamadasAntes := chat.chamadas

	resposta, err := p.ProcessarRequisicao(ctx, perguntaParafrase)
	if err != nil {
		t.Fatalf("erro inesperado na #2: %v", err)
	}
	if resposta == nil || *resposta != "resposta de rotina" {
		t.Fatalf("resposta = %v, esperado a resposta cacheada", resposta)
	}
	if chat.chamadas != chamadasAntes {
		t.Errorf("esperado NENHUMA chamada nova ao modelo (cache hit), obtido %d novas chamadas", chat.chamadas-chamadasAntes)
	}
}

// #3 do roteiro: síntese de CSR — sempre modelo caro, sempre Approval Gate,
// mesmo com confiança alta.
func TestProcessarRequisicaoSinteseCSRSempreGate(t *testing.T) {
	chat := &fakeChatStreamer{resposta: "resposta de csr"}
	approver := &fakeApprover{respostas: []bool{true}}
	var out bytes.Buffer

	p := montaProcessorDeTeste(t, map[string][]float64{
		perguntaCSR: {0, 0, 1}, // == Indices["csr"][0] — confiança 1.0
	}, chat, approver, &out)

	resposta, err := p.ProcessarRequisicao(context.Background(), perguntaCSR)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta == nil || *resposta != "resposta de csr" {
		t.Fatalf("resposta = %v, esperado 'resposta de csr'", resposta)
	}
	if chat.ultimoModelo != "gemma4:latest" {
		t.Errorf("modelo usado = %q, esperado modelo caro", chat.ultimoModelo)
	}
	if approver.chamadas != 1 {
		t.Errorf("esperado 1 chamada ao Approval Gate (CSR é sempre gated), obtido %d", approver.chamadas)
	}
}

// #4 do roteiro: pergunta de tema bem diferente de qualquer índice —
// Agentic RAG esgota as 3 iterações sem atingir o limiar, Confidence
// Threshold escala pro Approval Gate por confiança baixa.
func TestProcessarRequisicaoTemaDiferenteEscalaGate(t *testing.T) {
	chat := &fakeChatStreamer{resposta: "resposta incerta"}
	approver := &fakeApprover{respostas: []bool{false}}
	var out bytes.Buffer

	p := montaProcessorDeTeste(t, map[string][]float64{
		perguntaTemaDiferente: {0, 0, -1}, // ortogonal/oposto a todos os índices
	}, chat, approver, &out)

	resposta, err := p.ProcessarRequisicao(context.Background(), perguntaTemaDiferente)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta != nil {
		t.Errorf("esperado rascunho REJEITADO (nil), obtido %v", *resposta)
	}
	if approver.chamadas != 1 {
		t.Errorf("esperado 1 chamada ao Approval Gate por confiança baixa, obtido %d", approver.chamadas)
	}
}

// #5 do roteiro: pergunta de critério de protocolo — Multi-Index roteia pro
// índice "protocolo" e Hybrid Search converge já na 1ª iteração, sem gate.
func TestProcessarRequisicaoProtocoloConvergePrimeiraIteracao(t *testing.T) {
	chat := &fakeChatStreamer{resposta: "resposta de protocolo"}
	approver := &fakeApprover{}
	var out bytes.Buffer

	p := montaProcessorDeTeste(t, map[string][]float64{
		perguntaProtocolo: {0, 1, 0}, // == Indices["protocolo"][0] — confiança 1.0
	}, chat, approver, &out)

	resposta, err := p.ProcessarRequisicao(context.Background(), perguntaProtocolo)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta == nil || *resposta != "resposta de protocolo" {
		t.Fatalf("resposta = %v, esperado 'resposta de protocolo'", resposta)
	}
	if approver.chamadas != 0 {
		t.Errorf("esperado 0 chamadas ao Approval Gate, obtido %d", approver.chamadas)
	}
}

// Rejeição no Approval Gate: a resposta NÃO deve ser adicionada ao Semantic
// Cache (senão uma rejeição vazaria pra respostas futuras).
func TestProcessarRequisicaoRejeicaoNaoAlimentaCache(t *testing.T) {
	chat := &fakeChatStreamer{resposta: "rascunho rejeitado"}
	approver := &fakeApprover{respostas: []bool{false}}
	var out bytes.Buffer

	p := montaProcessorDeTeste(t, map[string][]float64{
		perguntaTemaDiferente: {0, 0, -1},
	}, chat, approver, &out)

	ctx := context.Background()
	if _, err := p.ProcessarRequisicao(ctx, perguntaTemaDiferente); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(p.Cache.entradas) != 0 {
		t.Errorf("esperado cache vazio após rejeição, obtido %d entrada(s)", len(p.Cache.entradas))
	}
}
