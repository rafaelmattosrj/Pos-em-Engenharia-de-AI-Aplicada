package com.psprouting.application;

import com.psprouting.domain.SimilarCase;
import com.psprouting.domain.Transaction;
import org.springframework.core.io.ClassPathResource;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.List;

/**
 * Monta o prompt RAG enviado ao LLM a partir do template fixo
 * ({@code resources/prompts/routing-prompt.txt}, extraido 1:1 do IDEIA.md),
 * da transacao atual e dos casos historicos similares recuperados no Neo4j.
 *
 * Extraida do fluxo principal para ser testavel sem depender do Neo4j, do
 * modelo de embeddings ou de uma chamada HTTP real ao OpenRouter — mesmo
 * espirito do {@code PromptBuilder} de {@code pdf-rag-knowledge-base-java}.
 */
@Component
public class RoutingPromptBuilder {

    private static final String TEMPLATE_PATH = "prompts/routing-prompt.txt";

    private final String template;

    public RoutingPromptBuilder() {
        this.template = loadTemplate();
    }

    private static String loadTemplate() {
        try {
            return new String(
                    new ClassPathResource(TEMPLATE_PATH).getInputStream().readAllBytes(),
                    StandardCharsets.UTF_8
            );
        } catch (IOException e) {
            throw new UncheckedIOException("Nao foi possivel carregar " + TEMPLATE_PATH, e);
        }
    }

    public String build(Transaction transaction, List<SimilarCase> similarCases) {
        return template
                .replace("{transaction}", describeTransaction(transaction))
                .replace("{similarCases}", describeSimilarCases(similarCases));
    }

    private static String describeTransaction(Transaction transaction) {
        return TransactionTextSerializer.serialize(transaction);
    }

    private static String describeSimilarCases(List<SimilarCase> similarCases) {
        if (similarCases.isEmpty()) {
            return "(nenhum caso historico similar encontrado)";
        }

        StringBuilder text = new StringBuilder();
        for (SimilarCase c : similarCases) {
            text.append("- PSP=").append(c.psp())
                    .append(", status=").append(c.status())
                    .append(", valor=R$").append(c.amount())
                    .append(", metodo=").append(c.method())
                    .append(", similaridade=").append(String.format("%.2f", c.similarity()))
                    .append("\n");
        }
        return text.toString().stripTrailing();
    }
}
