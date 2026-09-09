package com.opspilot;

import com.opspilot.team.BlackboardEntry;
import com.opspilot.team.RoleRunInput;
import com.opspilot.team.RoleRunResult;
import com.opspilot.team.RoleRunner;
import com.opspilot.team.SupervisorDecision;
import com.opspilot.team.TeamGraph;
import com.opspilot.team.TeamRole;
import com.opspilot.tools.Tool;
import org.junit.jupiter.api.Test;

import java.util.EnumMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;

class TeamGraphTest {

    private static RoleRunner fakeRunner(TeamRole role, BlackboardEntry.Kind kind, String content) {
        return new RoleRunner() {
            public TeamRole role() {
                return role;
            }

            public List<Tool> tools() {
                return List.of();
            }

            public RoleRunResult run(RoleRunInput input) {
                return new RoleRunResult(new BlackboardEntry(role, kind, input.brief(), content), List.of(), 1);
            }
        };
    }

    private static Map<TeamRole, RoleRunner> allRunners() {
        Map<TeamRole, RoleRunner> runners = new EnumMap<>(TeamRole.class);
        runners.put(TeamRole.ANALISTA, fakeRunner(TeamRole.ANALISTA, BlackboardEntry.Kind.FACTS, "fatos coletados"));
        runners.put(TeamRole.PLANEJADOR, fakeRunner(TeamRole.PLANEJADOR, BlackboardEntry.Kind.PLAN, "plano de 3 passos"));
        runners.put(TeamRole.EXECUTOR, fakeRunner(TeamRole.EXECUTOR, BlackboardEntry.Kind.EXECUTION, "acao executada"));
        return runners;
    }

    @Test
    void delegatesThenFinishesWithBriefAsAnswer() {
        TeamGraph graph = new TeamGraph(
                (message, blackboard, handoffCount) -> blackboard.isEmpty()
                        ? new SupervisorDecision(TeamRole.ANALISTA, "levante os fatos")
                        : SupervisorDecision.done("Resumo final do supervisor"),
                allRunners(), 1);

        TeamGraph.Result result = graph.run("o que esta acontecendo?");

        assertThat(result.answer()).isEqualTo("Resumo final do supervisor");
        assertThat(result.llmCalls()).isEqualTo(3); // 2 chamadas ao supervisor + 1 ao papel
    }

    @Test
    void fallsBackToBlackboardWhenDoneBriefIsEmpty() {
        AtomicInteger calls = new AtomicInteger();
        TeamGraph graph = new TeamGraph(
                (message, blackboard, handoffCount) -> {
                    if (calls.getAndIncrement() == 0) {
                        return new SupervisorDecision(TeamRole.ANALISTA, "levante os fatos");
                    }
                    return SupervisorDecision.done("");
                },
                allRunners(), 1);

        TeamGraph.Result result = graph.run("mensagem");

        assertThat(result.answer()).contains("Resumo do blackboard").contains("fatos coletados");
    }

    @Test
    void stopsAtHandoffCapEvenIfSupervisorNeverSaysDone() {
        TeamGraph graph = new TeamGraph(
                (message, blackboard, handoffCount) -> new SupervisorDecision(TeamRole.ANALISTA, "de novo"),
                allRunners(), 1);

        TeamGraph.Result result = graph.run("mensagem");

        long handoffEvents = result.trace().stream()
                .filter(event -> event.type() == com.opspilot.domain.TraceEvent.Type.HANDOFF)
                .count();
        assertThat(handoffEvents).isEqualTo(TeamGraph.MAX_HANDOFFS + 1); // +1 e o handoff final pra "done"
        assertThat(result.answer()).contains("Resumo do blackboard");
    }

    @Test
    void invalidSupervisorDecisionEndsGracefully() {
        TeamGraph graph = new TeamGraph(
                (message, blackboard, handoffCount) -> {
                    throw new IllegalArgumentException("json malformado");
                },
                allRunners(), 1);

        TeamGraph.Result result = graph.run("mensagem");

        assertThat(result.answer()).contains("A equipe encerrou sem contribuicoes");
        assertThat(result.trace().get(0).content()).contains(TeamGraph.INVALID_DECISION_PREFIX);
    }

    @Test
    void emptyMessageWithNoBlackboardStillProducesAnswer() {
        TeamGraph graph = new TeamGraph(
                (message, blackboard, handoffCount) -> SupervisorDecision.done(""),
                allRunners(), 1);

        TeamGraph.Result result = graph.run("mensagem vazia de contribuicoes");

        assertThat(result.answer()).contains("A equipe encerrou sem contribuicoes no blackboard para: mensagem vazia de contribuicoes");
    }
}
