package com.opspilot.team;

import com.opspilot.domain.ConversationMessage;
import com.opspilot.domain.ExecutionMetrics;
import com.opspilot.domain.ReasoningStrategy;
import com.opspilot.domain.StrategyRunInput;
import com.opspilot.domain.StrategyResult;

import java.util.EnumMap;
import java.util.Map;

/** Porte de TeamStrategy (team/team-strategy.ts). */
public class TeamStrategy implements ReasoningStrategy {

    private final TeamGraph graph;

    public TeamStrategy(DecideNextFn decideNext, Map<TeamRole, RoleRunner> roleRunners, int supervisorLlmCalls) {
        this.graph = new TeamGraph(decideNext, new EnumMap<>(roleRunners), supervisorLlmCalls);
    }

    @Override
    public String name() {
        return "team";
    }

    @Override
    public StrategyResult run(StrategyRunInput input) {
        long startedAt = System.currentTimeMillis();
        TeamGraph.Result result = graph.run(composeMessage(input));
        return new StrategyResult(
                result.answer(),
                result.trace(),
                new ExecutionMetrics(result.llmCalls(), System.currentTimeMillis() - startedAt));
    }

    private static String composeMessage(StrategyRunInput input) {
        if (input.history().isEmpty()) {
            return input.message();
        }
        StringBuilder sb = new StringBuilder("Historico recente da conversa:\n");
        for (ConversationMessage message : input.history()) {
            sb.append(message.role()).append(": ").append(message.content()).append('\n');
        }
        sb.append('\n').append(input.message());
        return sb.toString();
    }
}
