// Porte Go de model-eval-gate-prototype.js / model_eval_gate_prototype.py.
// Ver README.md deste projeto para detalhes de paridade e adaptações.
package main

import (
	"context"
	"fmt"
	"os"

	"model-eval-gate-prototype/evalgate"
	"model-eval-gate-prototype/ollama"
)

func main() {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := ollama.NewClient(baseURL)
	gate := evalgate.NewEvalGate(client)
	ctx := context.Background()

	if err := run(ctx, gate); err != nil {
		fmt.Fprintln(os.Stderr, "[Erro não tratado]", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, gate *evalgate.EvalGate) error {
	fmt.Printf("\n[Eval Gate] Avaliando BASELINE (%s) contra golden set de %d perguntas...\n",
		evalgate.ModeloBaseline, len(evalgate.GoldenSet))
	scoreBaseline, err := gate.AvaliarCandidato(ctx, evalgate.ModeloBaseline, evalgate.GoldenSet)
	if err != nil {
		return err
	}
	fmt.Printf("[Eval Gate] Baseline: score médio %.3f\n", scoreBaseline)

	fmt.Println("\n== Cenário 1: candidato real (variante MLX do mesmo modelo) ==")
	fmt.Printf("[Eval Gate] Avaliando CANDIDATO (%s) contra o mesmo golden set...\n", evalgate.ModeloCandidato)
	scoreCandidato1, err := gate.AvaliarCandidato(ctx, evalgate.ModeloCandidato, evalgate.GoldenSet)
	if err != nil {
		return err
	}
	fmt.Printf("[Eval Gate] Candidato: score médio %.3f\n", scoreCandidato1)
	diferenca1 := scoreCandidato1 - scoreBaseline
	promove1 := evalgate.DecidirPromocao(scoreBaseline, scoreCandidato1, evalgate.ToleranciaRegressao)
	fmt.Printf("[Eval Gate] Diferença: %s%.3f | Tolerância: -%v\n", sinal(diferenca1), diferenca1, evalgate.ToleranciaRegressao)
	fmt.Printf("[Eval Gate] Decisão: %s — caso limite, dois modelos reais e parecidos; pode mudar entre "+
		"execuções por variância do próprio modelo, por isso a decisão nunca deve ser no olho, sempre pelo gate.\n",
		decisaoTexto(promove1))

	fmt.Println("\n== Cenário 2: candidato regredido por config, não por modelo pior ==")
	fmt.Printf("[Eval Gate] Avaliando %s SEM a cláusula no contexto (simula bug de RAG/config, mesmo modelo)...\n",
		evalgate.ModeloBaseline)
	scoreCandidato2, err := gate.AvaliarCandidatoSemContexto(ctx, evalgate.ModeloBaseline, evalgate.GoldenSet)
	if err != nil {
		return err
	}
	fmt.Printf("[Eval Gate] Candidato regredido: score médio %.3f\n", scoreCandidato2)
	diferenca2 := scoreCandidato2 - scoreBaseline
	promove2 := evalgate.DecidirPromocao(scoreBaseline, scoreCandidato2, evalgate.ToleranciaRegressao)
	fmt.Printf("[Eval Gate] Diferença: %s%.3f | Tolerância: -%v\n", sinal(diferenca2), diferenca2, evalgate.ToleranciaRegressao)
	fmt.Printf("[Eval Gate] Decisão: %s — mesmo modelo, contexto perdido; o gate pega uma regressão de "+
		"config que nenhuma troca de modelo causou.\n", decisaoTexto(promove2))

	return nil
}

func sinal(v float64) string {
	if v >= 0 {
		return "+"
	}
	return ""
}

func decisaoTexto(promove bool) string {
	if promove {
		return "PROMOVE"
	}
	return "BLOQUEIA"
}
