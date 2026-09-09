// Demo de ponta a ponta: conversão pro formato Gemini, gate de OCR,
// validação de hiperparâmetro, escala de dataset (dedup+balanceamento),
// reavaliação do caso Saúde Empresarial. As chamadas de rede reais (Vertex
// AI) não são executadas aqui por padrão -- exigem `gcloud` autenticado e
// projeto GCP com billing ativo (ver README para rodá-las manualmente).
// Equivalente ao main() de Main.java (fine-tuning-via-api-java).
package main

import (
	"fmt"
	"log"

	"fine-tuning-via-api/datasetscaling"
	"fine-tuning-via-api/decisionframework"
	"fine-tuning-via-api/gemini"
	"fine-tuning-via-api/hyperparam"
	"fine-tuning-via-api/ocrgate"
	"fine-tuning-via-api/reavaliacao"
)

func ptr(v float64) *float64 { return &v }

func main() {
	fmt.Println("== Conversão para o formato Gemini (Módulo 3.2) ==")
	convertido, err := gemini.Converter(
		"Extraia segurado, placa e valor do orçamento de oficina abaixo.",
		"Segurado: Camila Costa Ribeiro Placa do veiculo: AZS-6617 Valor total do reparo: R$ 1.780,50",
		map[string]any{"segurado": "Camila Costa Ribeiro", "placa": "AZS-6617", "valor": 1780.5},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Turnos gerados: %d\n", len(convertido.Contents))

	fmt.Println()
	fmt.Println("== Gate de confiança de OCR (Módulo 3.2) ==")
	documentos := []ocrgate.Exemplo{
		{ID: "doc-auto-1", ConfiancaOcr: ptr(0.943)},
		{ID: "doc-degradado", ConfiancaOcr: ptr(0.62)},
		{ID: "amplitude-auto-Oficina Estrela-5", ConfiancaOcr: nil},
	}
	resultadoGate := ocrgate.Filtrar(documentos)
	fmt.Printf("Aprovados por OCR: %d | Sinalizados: %d | Sem OCR: %d\n",
		len(resultadoGate.AprovadosPorOcr), len(resultadoGate.SinalizadosParaRevisao), len(resultadoGate.SemConfianca))

	fmt.Println()
	fmt.Println("== Validação de hiperparâmetro (Módulo 3.3) ==")
	if err := hyperparam.Validar(hyperparam.Hiperparametros{EpochCount: 0, LearningRateMultiplier: 5.0}); err != nil {
		fmt.Printf("Bloqueado ANTES de qualquer chamada de rede: %s\n", err)
	}

	fmt.Println()
	fmt.Println("== Escala do dataset (Módulo 3.2): 305 -> dedup -> 200 balanceado ==")
	bruto := datasetscaling.GerarDatasetBruto()
	alvos := map[string]int{"amplitude-auto": 120, "amplitude-saude-empresarial": 80}
	resultado := datasetscaling.LimparEBalancear(bruto, alvos)
	fmt.Printf("Bruto: %d -> Dedup: %d -> Balanceado: %d\n", resultado.Original, resultado.AposDedup, resultado.Total)

	fmt.Println()
	fmt.Println("== Reavaliação Amplitude Saúde Empresarial, 9 meses depois (Módulo 3.2) ==")
	config, err := decisionframework.CarregarConfiguracao()
	if err != nil {
		log.Fatal(err)
	}
	pesosAHP := decisionframework.DerivarPesosAHP(config.MatrizAHP)
	casoOriginal, err := config.Caso("amplitude-saude-empresarial")
	if err != nil {
		log.Fatal(err)
	}
	casoAtualizado := reavaliacao.ConstruirCasoNoveMesesDepois(casoOriginal)
	resultadoOriginal := decisionframework.AvaliarFramework(casoOriginal.Scores, pesosAHP, config.LimiarVerde)
	resultadoAtualizado := decisionframework.AvaliarFramework(casoAtualizado.Scores, pesosAHP, config.LimiarVerde)
	fmt.Printf("Módulo 1.3: %s\n", resultadoOriginal.Recomendacao)
	fmt.Printf("Módulo 3.2 (9 meses depois): %s\n", resultadoAtualizado.Recomendacao)

	fmt.Println()
	fmt.Println("== Versionamento de modelo (Módulo 3.5) ==")
	fmt.Println("Ver versioning.GerarFichaVersionamento / GerarModelCardMarkdown (requer job real consultado via vertexai.HTTPClient).")

	fmt.Println()
	fmt.Println("== Extra: dataset real alternativo Dolly-15k (Módulo 3.2/3.4/3.5) ==")
	fmt.Println("Requer baixar databricks-dolly-15k.jsonl (ver README) e chamar dolly.PrepararDatasetCompleto/RodarPipeline a partir do seu próprio main.")
}
