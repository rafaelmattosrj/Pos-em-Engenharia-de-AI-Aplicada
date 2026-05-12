package com.iadeva.evals.evaluator;

// Equivalente ao memory_eval.py da aula 15 — compara execuções com e sem memória

import com.iadeva.evals.model.MemoryImpactReport;
import org.springframework.stereotype.Component;

import java.util.Random;

/**
 * Avalia o impacto do uso de memória no desempenho do agente.
 * Simula múltiplas execuções com e sem memória e calcula as métricas de comparação.
 *
 * As métricas são simuladas com valores realistas para demonstrar o framework.
 * Em produção, substituir pela execução real do agente com e sem memória ativa.
 */
@Component
public class MemoryEvaluator {

    // Número padrão de execuções simuladas para cada configuração
    private static final int DEFAULT_RUNS = 10;

    // Semente fixa para reprodutibilidade
    private final Random random = new Random(99);

    /**
     * Executa a avaliação de impacto da memória.
     * Simula DEFAULT_RUNS execuções com memória e sem memória e calcula delta.
     *
     * @return relatório com métricas comparativas
     */
    public MemoryImpactReport evaluate() {
        return evaluate(DEFAULT_RUNS);
    }

    /**
     * Executa a avaliação com número configurável de execuções.
     *
     * @param totalRuns número de execuções a simular
     * @return relatório com métricas comparativas
     */
    public MemoryImpactReport evaluate(int totalRuns) {
        // Simulação de recuperação de memória: precision e recall
        // Valores realistas esperados para um sistema RAG bem calibrado
        double retrievalPrecision = simularMetrica(0.84, 0.06);
        double retrievalRecall = simularMetrica(0.79, 0.07);

        // Percentual de execuções onde memória foi consultada e utilizada
        double memoryUtilization = simularMetrica(0.71, 0.08);

        // Taxa de alucinações geradas a partir de memória incorreta ou desatualizada
        // Deve ser baixo — esperado < 0.10 para sistema saudável
        double hallucinationFromMemory = simularMetrica(0.06, 0.03);

        // Melhoria percentual nas decisões com memória vs sem memória
        // Calculado como: (acertos_com_memoria - acertos_sem_memoria) / acertos_sem_memoria
        double decisionImprovement = simularMetrica(0.23, 0.09);

        // Qualidade média das lições aprendidas (score 0-1, avaliado por rubrica)
        double lessonQuality = simularMetrica(0.78, 0.05);

        return new MemoryImpactReport(
                totalRuns,
                round(retrievalPrecision),
                round(retrievalRecall),
                round(memoryUtilization),
                round(hallucinationFromMemory),
                round(decisionImprovement),
                round(lessonQuality)
        );
    }

    /**
     * Simula uma métrica com valor base e variação aleatória controlada.
     *
     * @param base      valor central esperado
     * @param variance  amplitude máxima de variação
     * @return valor simulado entre 0 e 1
     */
    private double simularMetrica(double base, double variance) {
        double value = base + (random.nextDouble() - 0.5) * variance * 2;
        return Math.max(0.0, Math.min(1.0, value));
    }

    private double round(double value) {
        return Math.round(value * 10000.0) / 10000.0;
    }
}
