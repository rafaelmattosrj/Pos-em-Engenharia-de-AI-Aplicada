package com.opspilot.team;

import com.opspilot.domain.TraceEvent;

import java.util.ArrayList;
import java.util.EnumMap;
import java.util.List;
import java.util.Map;

/**
 * Porte de team-graph.ts: a mesma maquina de estado (sem depender do
 * StateGraph do LangGraph) -- supervisor decide o proximo papel, cada papel
 * contribui uma entrada no blackboard, ate a decisao "done" ou o teto de
 * handoffs (MAX_HANDOFFS).
 */
public final class TeamGraph {

    public static final int MAX_HANDOFFS = 8;
    public static final String CAP_REACHED_PREFIX = "teto de handoffs atingido";
    public static final String INVALID_DECISION_PREFIX = "decisao invalida do supervisor";

    private final DecideNextFn decideNext;
    private final Map<TeamRole, RoleRunner> roleRunners;
    private final int supervisorLlmCalls;

    public TeamGraph(DecideNextFn decideNext, Map<TeamRole, RoleRunner> roleRunners, int supervisorLlmCalls) {
        this.decideNext = decideNext;
        this.roleRunners = new EnumMap<>(roleRunners);
        this.supervisorLlmCalls = supervisorLlmCalls;
    }

    public record Result(String answer, List<TraceEvent> trace, int llmCalls) {
    }

    public Result run(String message) {
        List<BlackboardEntry> blackboard = new ArrayList<>();
        List<TraceEvent> trace = new ArrayList<>();
        int handoffCount = 0;
        int llmCalls = 0;
        String lastBrief = "";

        while (true) {
            if (handoffCount >= MAX_HANDOFFS) {
                String content = CAP_REACHED_PREFIX + ": encerrando com o conteudo do blackboard";
                trace.add(TraceEvent.handoff("supervisor", "done", content));
                lastBrief = "";
                break;
            }

            SupervisorDecision decision;
            try {
                decision = decideNext.decideNext(message, blackboard, handoffCount);
            } catch (RuntimeException e) {
                String content = INVALID_DECISION_PREFIX + ": " + e.getMessage();
                trace.add(TraceEvent.handoff("supervisor", "done", content));
                llmCalls += supervisorLlmCalls;
                lastBrief = "";
                break;
            }

            trace.add(TraceEvent.handoff(
                    "supervisor", decision.isDone() ? "done" : decision.next().name().toLowerCase(), decision.brief()));
            llmCalls += supervisorLlmCalls;

            if (decision.isDone()) {
                lastBrief = decision.brief();
                break;
            }

            handoffCount++;
            RoleRunner runner = roleRunners.get(decision.next());
            try {
                RoleRunResult result = runner.run(new RoleRunInput(message, decision.brief(), List.copyOf(blackboard)));
                blackboard.add(result.entry());
                trace.addAll(result.trace());
                llmCalls += result.llmCalls();
            } catch (RuntimeException e) {
                String content = "erro no papel " + decision.next().name().toLowerCase() + ": " + e.getMessage();
                blackboard.add(new BlackboardEntry(decision.next(), BlackboardEntry.Kind.ERROR, decision.brief(), content));
                trace.add(TraceEvent.observation(decision.next().name().toLowerCase(), content));
            }
        }

        String answer;
        String trimmedBrief = lastBrief.trim();
        if (!trimmedBrief.isEmpty()) {
            answer = trimmedBrief;
        } else if (!blackboard.isEmpty()) {
            answer = "Resumo do blackboard:\n" + Blackboard.render(blackboard);
        } else {
            answer = "A equipe encerrou sem contribuicoes no blackboard para: " + message;
        }
        trace.add(TraceEvent.answer("supervisor", answer));

        return new Result(answer, trace, llmCalls);
    }
}
