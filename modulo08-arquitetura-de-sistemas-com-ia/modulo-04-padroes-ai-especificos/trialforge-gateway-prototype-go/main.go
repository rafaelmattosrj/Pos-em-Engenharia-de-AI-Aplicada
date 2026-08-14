// Demonstração do Gateway do TrialForge — porte Go de
// trialforge-gateway-prototype.js / trialforge_gateway_prototype.py (Módulo
// 8, Módulo 4.5 do curso de Arquitetura de Sistemas com IA). Ver README.md
// para o que foi mantido 1:1 e o que foi adaptado.
package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"trialforge-gateway-prototype/gateway"
	"trialforge-gateway-prototype/ollama"
)

func log(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// autoTeste roda os mesmos testes puros (sem rede) que os dois originais
// executam no início de main() — determinismo do classificador de intenção,
// da matemática do cosseno, do BM25 e da fusão RRF — antes de qualquer
// chamada ao Ollama. Os mesmos cenários também viram testes automatizados em
// gateway/*_test.go; aqui a função é mantida pra reproduzir a MESMA saída de
// console e o MESMO comportamento de abortar antes de tocar a rede.
func autoTeste() error {
	log("== Testes: ClassificarIntencao + SimilaridadeCosseno (puros, sem rede) ==")
	passou := 0
	total := 0

	casosIntencao := []struct {
		pergunta string
		esperado string
	}{
		{"Preciso da síntese do CSR final desse estudo.", "sintese_csr"},
		{"Quero o relatório final do estudo.", "sintese_csr"},
		{"Como os eventos adversos aparecem no relatório final?", "sintese_csr"},
		{"Qual é o critério de idade mínima pra participar desse estudo?", "consulta_protocolo"},
		{"Quais são os critérios de exclusão desse protocolo?", "consulta_protocolo"},
		{"Quais são as regras de assentimento pra menores?", "consulta_icf"},
		{"Qual o prazo de armazenamento das amostras biológicas?", "consulta_icf"},
	}
	for _, caso := range casosIntencao {
		total++
		resultado := gateway.ClassificarIntencao(caso.pergunta)
		ok := resultado == caso.esperado
		status := "FALHOU"
		if ok {
			status = "OK"
			passou++
		}
		log("  [%s] ClassificarIntencao(%q) -> %s (esperado %s)", status, caso.pergunta, resultado, caso.esperado)
	}

	vetorA := []float64{1, 0, 0}
	vetorB := []float64{1, 0, 0}
	vetorOrtogonal := []float64{0, 1, 0}
	vetorOposto := []float64{-1, 0, 0}
	casosCosseno := []struct {
		nome     string
		a, b     []float64
		esperado float64
	}{
		{"vetores idênticos", vetorA, vetorB, 1},
		{"vetores ortogonais", vetorA, vetorOrtogonal, 0},
		{"vetores opostos", vetorA, vetorOposto, -1},
	}
	for _, caso := range casosCosseno {
		total++
		resultado := gateway.SimilaridadeCosseno(caso.a, caso.b)
		ok := math.Abs(resultado-caso.esperado) < 1e-9
		status := "FALHOU"
		if ok {
			status = "OK"
			passou++
		}
		log("  [%s] SimilaridadeCosseno(%s) -> %.3f (esperado %v)", status, caso.nome, resultado, caso.esperado)
	}

	log("\n== Testes: ScoreBM25 + FusaoReciprocalRank (puros, sem rede) ==")

	corpusTeste := []string{
		"idade mínima de doze anos para participar do estudo",
		"consentimento do responsável legal é obrigatório",
		"retirada do participante a qualquer momento sem justificativa",
	}
	estatisticasTeste := gateway.ConstruirEstatisticasBM25(corpusTeste)
	queryTeste := gateway.Tokenizar("qual a idade mínima exigida")
	melhorScore := math.Inf(-1)
	vencedorBM25 := -1
	for idx := range corpusTeste {
		score := gateway.ScoreBM25Padrao(queryTeste, estatisticasTeste.TokensPorDoc[idx], estatisticasTeste)
		if score > melhorScore {
			melhorScore = score
			vencedorBM25 = idx
		}
	}
	total++
	okBM25 := vencedorBM25 == 0
	statusBM25 := "FALHOU"
	if okBM25 {
		statusBM25 = "OK"
		passou++
	}
	log("  [%s] ScoreBM25: doc sobre idade mínima vence a busca lexical (venceu doc %d)", statusBM25, vencedorBM25)

	fusaoTeste := gateway.FusaoReciprocalRank([]int{0, 1, 2}, []int{0, 2, 1}, 60)
	total++
	okFusao := fusaoTeste[0].Idx == 0
	statusFusao := "FALHOU"
	if okFusao {
		statusFusao = "OK"
		passou++
	}
	log("  [%s] FusaoReciprocalRank: doc no topo dos dois rankings vence a fusão (venceu doc %d)", statusFusao, fusaoTeste[0].Idx)

	log("Total: %d teste(s), %d passou(passaram), %d falhou(falharam).\n", total, passou, total-passou)
	if passou != total {
		return fmt.Errorf("testes puros falharam: %d/%d — corrija antes de chamar o modelo", passou, total)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log("[Erro não tratado] %s", err)
		dir, _ := os.Getwd()
		trilha := ollamaAuditPath(dir)
		_ = gateway.NewAuditTrail(trilha).Registrar(map[string]interface{}{
			"erro":         err.Error(),
			"status_final": "falha_tecnica",
		})
		os.Exit(1)
	}
}

func ollamaAuditPath(dir string) string {
	return filepath.Join(dir, "audit-trail.jsonl")
}

func run() error {
	loadDotEnv(".env")

	if err := autoTeste(); err != nil {
		return err
	}

	baseURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
	modeloBarato := getEnv("OLLAMA_MODELO_BARATO", "gemma4:e2b")
	modeloCaro := getEnv("OLLAMA_MODELO_CARO", "gemma4:latest")
	modeloEmbedding := getEnv("OLLAMA_MODELO_EMBEDDING", "nomic-embed-text")

	// Os dois limiares abaixo foram calibrados testando de verdade contra o
	// nomic-embed-text em perguntas curtas em português (Módulo 4.3: "não
	// existe limiar universal, se calibra por tipo de pergunta"). Nesse par
	// modelo+idioma, paráfrases próximas reaproveitando termos do domínio
	// ficaram em ~0.82 de similaridade, enquanto perguntas de tema totalmente
	// diferente ficaram em ~0.64-0.65 — o corte abaixo fica no meio desse
	// intervalo, não é um valor padrão de mercado.
	const limiarCache = 0.75
	const limiarConfianca = 0.7

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	trilhaAuditoria := ollamaAuditPath(dir)

	client := ollama.NewClient(baseURL, modeloEmbedding)
	client.Log = log

	ctx := context.Background()

	preparados, err := gateway.PrepararIndices(ctx, client, log)
	if err != nil {
		return err
	}

	processor := &gateway.Processor{
		Embedder:        client,
		Chat:            client,
		Preparados:      preparados,
		Cache:           &gateway.SemanticCache{},
		Audit:           gateway.NewAuditTrail(trilhaAuditoria),
		Approval:        gateway.NewApprovalGate(os.Stdin, os.Stdout),
		Out:             os.Stdout,
		ModeloBarato:    modeloBarato,
		ModeloCaro:      modeloCaro,
		LimiarCache:     limiarCache,
		LimiarConfianca: limiarConfianca,
	}

	perguntas := []string{
		// 1) Pergunta de rotina — gera de verdade, depois popula o Semantic Cache
		"Quais são as regras de assentimento pra menores nesse estudo?",
		// 2) Paráfrase da pergunta 1, reaproveitando os termos do domínio —
		// deve bater no Semantic Cache
		"O assentimento dos menores de idade é obrigatório nesse estudo?",
		// 3) Síntese de CSR — sempre modelo caro + sempre Approval Gate, nunca cache
		"Preciso da síntese do CSR final desse estudo.",
		// 4) Pergunta de tema bem diferente de qualquer índice — nenhuma das 3
		// iterações do Agentic RAG encontra confiança suficiente (nem ampliando
		// o índice, nem cruzando os 3 domínios), então esgota o limite e o
		// Confidence Threshold escala pro Approval Gate
		"Qual o prazo de armazenamento das amostras biológicas coletadas nesse estudo?",
		// 5) Pergunta sobre critério de protocolo — Multi-Index roteia pro
		// índice "protocolo" (não "icf"), e Hybrid Search acha a cláusula com
		// confiança já na 1ª iteração
		"Qual é o critério de idade mínima pra participar desse estudo?",
	}

	for _, pergunta := range perguntas {
		if _, err := processor.ProcessarRequisicao(ctx, pergunta); err != nil {
			return err
		}
	}

	return gateway.VerificarTrilhaAuditoria(trilhaAuditoria, modeloCaro, limiarConfianca, log)
}
