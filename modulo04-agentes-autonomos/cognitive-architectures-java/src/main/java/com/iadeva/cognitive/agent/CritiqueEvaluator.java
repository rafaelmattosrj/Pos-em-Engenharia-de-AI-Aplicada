package com.iadeva.cognitive.agent;

import com.iadeva.cognitive.model.CritiqueResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

// Equivalente ao critique_evaluator.py — avalia qualidade do output usando LLM como juiz.
// O LLM pontua o output em três dimensões independentes: correção, completude e qualidade.
@Component
public class CritiqueEvaluator {

    private static final Logger log = LoggerFactory.getLogger(CritiqueEvaluator.class);

    private final ChatClient chatClient;

    @Value("${agent.reflection.threshold:0.7}")
    private double threshold;

    public CritiqueEvaluator(ChatClient.Builder chatClientBuilder) {
        this.chatClient = chatClientBuilder.build();
    }

    /**
     * Avalia um output usando LLM como juiz em três dimensões:
     * - correctness:  correção factual e lógica
     * - completeness: cobertura completa da tarefa
     * - quality:      clareza e qualidade de escrita
     *
     * @param originalTask tarefa original que foi solicitada ao agente
     * @param output       resposta produzida pelo agente que será avaliada
     * @return CritiqueResult com score global, flag passed e feedback detalhado
     */
    public CritiqueResult evaluate(String originalTask, String output) {
        log.info("[CritiqueEvaluator] Avaliando output (tarefa: {})", originalTask);

        String prompt = buildCritiquePrompt(originalTask, output);
        String response = chatClient.prompt()
                .user(prompt)
                .call()
                .content();

        log.debug("[CritiqueEvaluator] Resposta de avaliação: {}", response);

        return parseCritiqueResponse(response);
    }

    /**
     * Monta o prompt que instrui o LLM a retornar pontuações estruturadas.
     */
    private String buildCritiquePrompt(String task, String output) {
        return """
                Você é um avaliador crítico de qualidade. Avalie o OUTPUT abaixo considerando a TAREFA original.
                
                TAREFA: %s
                
                OUTPUT GERADO:
                %s
                
                Avalie em três dimensões com pontuação de 0.0 a 1.0 e retorne EXATAMENTE neste formato:
                CORRECTNESS: <0.0-1.0>
                COMPLETENESS: <0.0-1.0>
                QUALITY: <0.0-1.0>
                FEEDBACK: <texto com sugestões de melhoria>
                
                Critérios:
                - CORRECTNESS: o output é factualmente correto e logicamente consistente?
                - COMPLETENESS: o output cobre todos os aspectos da tarefa solicitada?
                - QUALITY: o output é claro, bem estruturado e fácil de entender?
                """.formatted(task, output);
    }

    /**
     * Faz parse da resposta estruturada do LLM avaliador.
     * Em caso de falha no parse, retorna score neutro com feedback de erro.
     */
    private CritiqueResult parseCritiqueResponse(String response) {
        try {
            double correctness  = extractScore(response, "CORRECTNESS");
            double completeness = extractScore(response, "COMPLETENESS");
            double quality      = extractScore(response, "QUALITY");
            String feedback     = extractFeedback(response);

            // Score global = média simples das três dimensões
            double score = (correctness + completeness + quality) / 3.0;
            boolean passed = score >= threshold;

            log.info("[CritiqueEvaluator] Score: {:.2f} (correctness={}, completeness={}, quality={}) — {}",
                    score, correctness, completeness, quality, passed ? "APROVADO" : "REPROVADO");

            return new CritiqueResult(score, passed, feedback, correctness, completeness, quality);

        } catch (Exception e) {
            log.warn("[CritiqueEvaluator] Falha ao parsear resposta de avaliação: {}", e.getMessage());
            // Fallback conservador: score baixo para forçar nova iteração
            return new CritiqueResult(0.5, false,
                    "Erro ao processar avaliação: " + e.getMessage(),
                    0.5, 0.5, 0.5);
        }
    }

    /**
     * Extrai o valor numérico de uma linha "FIELD: 0.X" via regex.
     */
    private double extractScore(String text, String field) {
        Pattern pattern = Pattern.compile(field + ":\\s*([0-9]*\\.?[0-9]+)", Pattern.CASE_INSENSITIVE);
        Matcher matcher = pattern.matcher(text);
        if (matcher.find()) {
            double value = Double.parseDouble(matcher.group(1));
            // Garante que está no intervalo válido
            return Math.max(0.0, Math.min(1.0, value));
        }
        // Valor padrão neutro quando o campo não é encontrado
        return 0.5;
    }

    /**
     * Extrai o texto de feedback da resposta estruturada.
     */
    private String extractFeedback(String text) {
        int idx = text.indexOf("FEEDBACK:");
        if (idx == -1) return "Nenhum feedback fornecido.";
        String feedback = text.substring(idx + "FEEDBACK:".length()).trim();
        return feedback.isEmpty() ? "Nenhum feedback fornecido." : feedback;
    }
}
