package com.finetuningapi;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Demo de ponta a ponta: conversão pro formato Gemini, gate de OCR,
 * validação de hiperparâmetro, escala de dataset (dedup+balanceamento),
 * reavaliação do caso Saúde Empresarial. As chamadas de rede reais (Vertex
 * AI) não são executadas aqui por padrão -- exigem `gcloud` autenticado e
 * projeto GCP com billing ativo; ver README para rodá-las manualmente.
 */
public final class Main {

    private Main() {
    }

    public static void main(String[] args) throws Exception {
        System.out.println("== Conversão para o formato Gemini (Módulo 3.2) ==");
        GeminiConverter.Exemplo exemplo = new GeminiConverter.Exemplo(
                "Extraia segurado, placa e valor do orçamento de oficina abaixo.",
                "Segurado: Camila Costa Ribeiro Placa do veiculo: AZS-6617 Valor total do reparo: R$ 1.780,50",
                GeminiConverter.exemploMap("Camila Costa Ribeiro", "AZS-6617", 1780.5));
        GeminiConverter.ConversaoGemini convertido = GeminiConverter.converter(exemplo);
        System.out.println("Turnos gerados: " + convertido.contents().size());

        System.out.println();
        System.out.println("== Gate de confiança de OCR (Módulo 3.2) ==");
        List<OcrConfidenceGate.ExemploComOcr> documentos = List.of(
                new OcrConfidenceGate.ExemploComOcr("doc-auto-1", 0.943),
                new OcrConfidenceGate.ExemploComOcr("doc-degradado", 0.62),
                new OcrConfidenceGate.ExemploComOcr("amplitude-auto-Oficina Estrela-5", null));
        var resultadoGate = OcrConfidenceGate.filtrar(documentos);
        System.out.println("Aprovados por OCR: " + resultadoGate.aprovadosPorOcr().size()
                + " | Sinalizados: " + resultadoGate.sinalizadosParaRevisao().size()
                + " | Sem OCR: " + resultadoGate.semConfianca().size());

        System.out.println();
        System.out.println("== Validação de hiperparâmetro (Módulo 3.3) ==");
        try {
            HyperparameterValidator.validar(new HyperparameterValidator.Hiperparametros(0, 5.0));
        } catch (IllegalArgumentException erro) {
            System.out.println("Bloqueado ANTES de qualquer chamada de rede: " + erro.getMessage());
        }

        System.out.println();
        System.out.println("== Escala do dataset (Módulo 3.2): 305 -> dedup -> 200 balanceado ==");
        List<DatasetScaling.ExemploDataset> bruto = DatasetScaling.gerarDatasetBruto();
        Map<String, Integer> alvos = new LinkedHashMap<>();
        alvos.put("amplitude-auto", 120);
        alvos.put("amplitude-saude-empresarial", 80);
        DatasetScaling.ResultadoPipeline resultado = DatasetScaling.limparEBalancear(bruto, alvos);
        System.out.println("Bruto: " + resultado.original() + " -> Dedup: " + resultado.aposDedup()
                + " -> Balanceado: " + resultado.total());

        System.out.println();
        System.out.println("== Reavaliação Amplitude Saúde Empresarial, 9 meses depois (Módulo 3.2) ==");
        DecisionFrameworkCore.Configuracao config = DecisionFrameworkCore.carregarConfiguracao();
        double[] pesosAHP = DecisionFrameworkCore.derivarPesosAHP(config.matrizAhp());
        DecisionFrameworkCore.Caso casoOriginal = config.caso("amplitude-saude-empresarial");
        DecisionFrameworkCore.Caso casoAtualizado = ReavaliacaoSaudeEmpresarial.construirCasoNoveMesesDepois(casoOriginal);
        var resultadoOriginal = DecisionFrameworkCore.avaliarFramework(casoOriginal.scores(), pesosAHP, config.limiarVerde());
        var resultadoAtualizado = DecisionFrameworkCore.avaliarFramework(casoAtualizado.scores(), pesosAHP, config.limiarVerde());
        System.out.println("Módulo 1.3: " + resultadoOriginal.recomendacao());
        System.out.println("Módulo 3.2 (9 meses depois): " + resultadoAtualizado.recomendacao());

        System.out.println();
        System.out.println("== Versionamento de modelo (Módulo 3.5) ==");
        System.out.println("Ver ModelVersioning.gerarFichaVersionamento / gerarModelCardMarkdown (requer job real consultado via VertexAiHttpClient).");
    }
}
