package com.psprouting.application;

import com.psprouting.domain.SimilarCase;
import com.psprouting.domain.Transaction;
import dev.langchain4j.model.chat.ChatModel;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * Passo (3) do pipeline: monta o prompt RAG e chama o LLM via OpenRouter
 * (LangChain4j {@link ChatModel}), devolvendo a recomendacao ja decodificada.
 */
@Service
public class ReasoningService {

    private final ChatModel chatModel;
    private final RoutingPromptBuilder promptBuilder;
    private final RecommendationParser parser;

    public ReasoningService(ChatModel chatModel, RoutingPromptBuilder promptBuilder, RecommendationParser parser) {
        this.chatModel = chatModel;
        this.promptBuilder = promptBuilder;
        this.parser = parser;
    }

    public ParsedRecommendation reason(Transaction transaction, List<SimilarCase> similarCases) {
        String prompt = promptBuilder.build(transaction, similarCases);
        String rawResponse = chatModel.chat(prompt);
        return parser.parse(rawResponse);
    }
}
