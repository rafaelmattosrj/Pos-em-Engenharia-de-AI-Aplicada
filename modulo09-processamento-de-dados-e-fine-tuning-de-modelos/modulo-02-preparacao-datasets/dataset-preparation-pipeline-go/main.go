// Demo do pipeline de preparacao de dataset (Modulo 2): gate de relevancia,
// deduplicacao+balanceamento+diversidade, e gate de higienizacao de PII.
//
// Os demos de extracao real (OCR via Tesseract e LLM multimodal via Vertex
// AI) dependem de binarios/credenciais externas nao disponiveis neste
// ambiente - chame extraction.OcrTexto/extraction.ExtrairViaLlm a partir do
// seu proprio codigo, com o binario `tesseract` instalado ou
// `gcloud auth login` feito, respectivamente. Ver README.md.
package main

import (
	"fmt"

	"dataset-preparation-pipeline/cleaning"
	"dataset-preparation-pipeline/datarelevance"
	"dataset-preparation-pipeline/pii"
)

func main() {
	demoRelevancia()
	demoLimpezaEBalanceamento()
	demoPiiScrubbing()
}

func demoRelevancia() {
	fmt.Println("===== Demo: gate de relevancia de dado =====")
	for _, candidato := range datarelevance.Candidatos {
		avaliacao := datarelevance.AvaliarCandidato(candidato.Criterios)
		status := "REJEITADO"
		if avaliacao.Aceito {
			status = "ACEITO"
		}
		fmt.Printf("[%s] %s (%s)\n", status, candidato.Nome, candidato.Caso)
	}
	fmt.Println()
}

func demoLimpezaEBalanceamento() {
	fmt.Println("===== Demo: MinHash+LSH, amostragem por temperatura, diversidade =====")
	dataset := cleaning.GerarDatasetSimulado()
	fmt.Printf("Dataset simulado: %d exemplos.\n", len(dataset))

	dedup := cleaning.EncontrarQuaseDuplicatasMinHashLSH(dataset)
	for caso, r := range dedup.ResultadosPorCaso {
		fmt.Printf("%s: %d exemplos, %d pares forca-bruta -> %d candidatos LSH (reducao %.1f%%), %d duplicata(s) confirmada(s)\n",
			caso, r.ItensNoCaso, r.ParesForcaBruta, r.CandidatosLSH, r.ReducaoPercentual, r.DuplicatasConfirmadas)
	}

	resultado := cleaning.LimparEBalancear(dataset, cleaning.AlphaTemperatura,
		map[string]int{"amplitude-auto": 20, "amplitude-saude-empresarial": 14})

	fmt.Printf("Pipeline: %d -> %d (dedup) -> %d (balanceado)\n", resultado.Original, resultado.AposDedup, resultado.Final)
	for caso, r := range resultado.RelatorioPorCaso {
		fmt.Printf("  %s: entropia antes=%.4f depois=%.4f | n efetivo antes=%.3f depois=%.3f\n",
			caso, r.EntropiaAntes, r.EntropiaDepois, r.NEfetivoAntes, r.NEfetivoDepois)
	}
	fmt.Println()
}

func demoPiiScrubbing() {
	fmt.Println("===== Demo: gate de higienizacao de PII =====")
	doc := "OFICINA ESTRELA - ORCAMENTO N. 4471\nSegurado: Marcos Vinicius Andrade Pereira\n" +
		"CPF: 111.444.777-35\nPlaca do veiculo: QJK-4F82\n\nValor total do reparo: R$ 3.210,50"
	resultado := pii.VarrerPII(doc)
	fmt.Println("Antes:\n" + doc)
	fmt.Println("Depois:\n" + resultado.TextoRedigido)
	fmt.Println()
}
