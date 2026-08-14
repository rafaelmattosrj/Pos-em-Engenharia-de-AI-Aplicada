// Package evalgate implementa o Eval Gate de modelo: antes de promover um
// candidato pra processar tráfego real, roda ele contra um golden set e só
// promove se o score médio não regredir contra o baseline além de uma
// tolerância. Porte de model-eval-gate-prototype.js / model_eval_gate_prototype.py.
package evalgate

import (
	"context"
	"fmt"
	"math"
)

const (
	ModeloEmbedding = "nomic-embed-text"
	ModeloBaseline  = "gemma4:e2b"
	ModeloCandidato = "gemma4:e2b-mlx"

	// TOLERANCIA_REGRESSAO: candidato pode ficar até essa fração abaixo do
	// baseline sem bloquear a promoção — tolerância, não corte exato.
	ToleranciaRegressao = 0.02
)

// GoldenItem é um item do golden set: pergunta com cláusula esperada
// CONHECIDA de antemão — mede se o modelo, dado o contexto certo, produz uma
// resposta fiel à cláusula, não se o RAG acha a cláusula certa.
type GoldenItem struct {
	Pergunta string
	Clausula string
	Fonte    string
}

// GoldenSet: mesmo banco de cláusulas do TrialForge usado desde o Módulo 2.5/4.5/5.4.
var GoldenSet = []GoldenItem{
	{
		Pergunta: "Quais são as regras de assentimento pra menores nesse estudo?",
		Clausula: "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, " +
			"além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
		Fonte: "RDC ANVISA 466/2012, Art. 4º",
	},
	{
		Pergunta: "O participante pode desistir do estudo a qualquer momento?",
		Clausula: "O participante pode retirar seu consentimento a qualquer momento, sem necessidade " +
			"de justificativa e sem prejuízo ao seu tratamento (RDC ANVISA 466/2012, Art. 5º).",
		Fonte: "RDC ANVISA 466/2012, Art. 5º",
	},
	{
		Pergunta: "Qual é o critério de idade mínima pra participar desse estudo?",
		Clausula: "A idade mínima para participação no estudo é de doze anos completos na data " +
			"da assinatura do assentimento, conforme a versão vigente do protocolo aprovada " +
			"pelo comitê de ética.",
		Fonte: "Protocolo Clínico TrialForge, critério de inclusão nº 2",
	},
}

// Gateway abstrai as duas operações do Ollama usadas neste protótipo: chat
// (geração) e embeddings. *ollama.Client satisfaz esta interface. Extraída
// para permitir testar EvalGate com um dublê determinístico, sem depender de
// um Ollama local rodando durante `go test`.
type Gateway interface {
	Chat(ctx context.Context, model, systemPrompt, userPrompt string) (string, error)
	Embed(ctx context.Context, model, texto string) ([]float64, error)
}

// SimilaridadeCosseno é lógica pura, testável sem rede.
func SimilaridadeCosseno(a, b []float64) float64 {
	var produto, normaA, normaB float64
	for i := range a {
		produto += a[i] * b[i]
		normaA += a[i] * a[i]
		normaB += b[i] * b[i]
	}
	return produto / (math.Sqrt(normaA) * math.Sqrt(normaB))
}

// DecidirPromocao é a decisão pura de promoção: o candidato é promovido se a
// diferença de score em relação ao baseline não for pior que -tolerancia.
func DecidirPromocao(scoreBaseline, scoreCandidato, tolerancia float64) bool {
	diferenca := scoreCandidato - scoreBaseline
	return diferenca >= -tolerancia
}

// EvalGate roda o golden set contra um modelo, via Gateway.
type EvalGate struct {
	Ollama Gateway
}

func NewEvalGate(gateway Gateway) *EvalGate {
	return &EvalGate{Ollama: gateway}
}

func (e *EvalGate) GerarResposta(ctx context.Context, modelo, pergunta, clausula, fonte string) (string, error) {
	system := "Você redige respostas curtas e precisas sobre regras de estudos clínicos, " +
		"citando a fonte regulatória fornecida."
	user := fmt.Sprintf("Pergunta: %s\n\nCláusula regulatória relevante: %s\nFonte: %s\n\nResponda usando essa cláusula.",
		pergunta, clausula, fonte)
	return e.Ollama.Chat(ctx, modelo, system, user)
}

// GerarRespostaSemContexto simula um candidato REGREDIDO por bug de config,
// não por modelo pior: o mesmo baseline, mas sem a cláusula no contexto.
func (e *EvalGate) GerarRespostaSemContexto(ctx context.Context, modelo, pergunta string) (string, error) {
	system := "Você redige respostas curtas e precisas sobre regras de estudos clínicos."
	user := fmt.Sprintf("Pergunta: %s", pergunta)
	return e.Ollama.Chat(ctx, modelo, system, user)
}

func (e *EvalGate) AvaliarCandidato(ctx context.Context, modelo string, goldenSet []GoldenItem) (float64, error) {
	var somaScores float64
	for _, item := range goldenSet {
		texto, err := e.GerarResposta(ctx, modelo, item.Pergunta, item.Clausula, item.Fonte)
		if err != nil {
			return 0, err
		}
		score, err := e.scoreContra(ctx, texto, item.Clausula)
		if err != nil {
			return 0, err
		}
		somaScores += score
		fmt.Printf("  [Eval] \"%s...\" -> score %.3f\n", ellipsis(item.Pergunta, 55), score)
	}
	return somaScores / float64(len(goldenSet)), nil
}

func (e *EvalGate) AvaliarCandidatoSemContexto(ctx context.Context, modelo string, goldenSet []GoldenItem) (float64, error) {
	var somaScores float64
	for _, item := range goldenSet {
		texto, err := e.GerarRespostaSemContexto(ctx, modelo, item.Pergunta)
		if err != nil {
			return 0, err
		}
		score, err := e.scoreContra(ctx, texto, item.Clausula)
		if err != nil {
			return 0, err
		}
		somaScores += score
		fmt.Printf("  [Eval] \"%s...\" -> score %.3f\n", ellipsis(item.Pergunta, 55), score)
	}
	return somaScores / float64(len(goldenSet)), nil
}

func (e *EvalGate) scoreContra(ctx context.Context, texto, clausula string) (float64, error) {
	embResposta, err := e.Ollama.Embed(ctx, ModeloEmbedding, texto)
	if err != nil {
		return 0, err
	}
	embClausula, err := e.Ollama.Embed(ctx, ModeloEmbedding, clausula)
	if err != nil {
		return 0, err
	}
	return SimilaridadeCosseno(embResposta, embClausula), nil
}

func ellipsis(texto string, tamanho int) string {
	runes := []rune(texto)
	if len(runes) <= tamanho {
		return texto
	}
	return string(runes[:tamanho])
}
