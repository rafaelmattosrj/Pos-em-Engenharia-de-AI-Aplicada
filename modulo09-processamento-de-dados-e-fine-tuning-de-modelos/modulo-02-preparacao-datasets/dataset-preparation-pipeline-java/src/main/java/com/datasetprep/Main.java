package com.datasetprep;

import com.datasetprep.cleaning.Balancing;
import com.datasetprep.cleaning.CleaningPipeline;
import com.datasetprep.cleaning.Diversity;
import com.datasetprep.cleaning.Exemplo;
import com.datasetprep.cleaning.MinHashLsh;
import com.datasetprep.cleaning.SampleDataset;
import com.datasetprep.pii.PiiScrubbingGate;

import java.util.List;
import java.util.Map;

/**
 * Demo do pipeline de preparacao de dataset (Modulo 2): gate de relevancia,
 * deduplicacao+balanceamento+diversidade, e gate de higienizacao de PII.
 *
 * Os demos de extracao real (OCR via Tesseract e LLM multimodal via Vertex
 * AI) dependem de binarios/credenciais externas nao disponiveis neste
 * ambiente - rode-os chamando diretamente OcrExtraction/LlmMultimodalExtraction
 * a partir do seu proprio codigo, com o binario `tesseract` instalado ou
 * `gcloud auth login` feito, respectivamente. Ver README.md.
 */
public final class Main {

    private Main() {
    }

    public static void main(String[] args) {
        demoRelevancia();
        demoLimpezaEBalanceamento();
        demoPiiScrubbing();
    }

    private static void demoRelevancia() {
        System.out.println("===== Demo: gate de relevancia de dado =====");
        for (var candidato : com.datasetprep.DataRelevance.CANDIDATOS) {
            var avaliacao = com.datasetprep.DataRelevance.avaliarCandidato(candidato.criterios());
            System.out.printf("[%s] %s (%s)%n", avaliacao.aceito() ? "ACEITO" : "REJEITADO", candidato.nome(), candidato.caso());
        }
        System.out.println();
    }

    private static void demoLimpezaEBalanceamento() {
        System.out.println("===== Demo: MinHash+LSH, amostragem por temperatura, diversidade =====");
        List<Exemplo> dataset = SampleDataset.gerarDatasetSimulado();
        System.out.println("Dataset simulado: " + dataset.size() + " exemplos.");

        MinHashLsh.ResultadoDedup dedup = MinHashLsh.encontrarQuaseDuplicatasMinHashLSH(dataset);
        dedup.resultadosPorCaso().forEach((caso, r) ->
                System.out.printf("%s: %d exemplos, %d pares forca-bruta -> %d candidatos LSH (reducao %.1f%%), %d duplicata(s) confirmada(s)%n",
                        caso, r.itensNoCaso(), r.paresForcaBruta(), r.candidatosLSH(), r.reducaoPercentual(), r.duplicatasConfirmadas()));

        CleaningPipeline.ResultadoPipeline resultado = CleaningPipeline.limparEBalancear(
                dataset, Balancing.ALPHA_TEMPERATURA,
                Map.of("amplitude-auto", 20, "amplitude-saude-empresarial", 14));

        System.out.printf("Pipeline: %d -> %d (dedup) -> %d (balanceado)%n",
                resultado.original(), resultado.aposDedup(), resultado.fin());
        resultado.relatorioPorCaso().forEach((caso, r) ->
                System.out.printf("  %s: entropia antes=%.4f depois=%.4f | n efetivo antes=%.3f depois=%.3f%n",
                        caso, r.entropiaAntes(), r.entropiaDepois(), r.nEfetivoAntes(), r.nEfetivoDepois()));
        System.out.println();
    }

    private static void demoPiiScrubbing() {
        System.out.println("===== Demo: gate de higienizacao de PII =====");
        String doc = "OFICINA ESTRELA - ORCAMENTO N. 4471\nSegurado: Marcos Vinicius Andrade Pereira\n"
                + "CPF: 111.444.777-35\nPlaca do veiculo: QJK-4F82\n\nValor total do reparo: R$ 3.210,50";
        var resultado = PiiScrubbingGate.varrerPII(doc);
        System.out.println("Antes:\n" + doc);
        System.out.println("Depois:\n" + resultado.textoRedigido());
        System.out.println();
    }
}
