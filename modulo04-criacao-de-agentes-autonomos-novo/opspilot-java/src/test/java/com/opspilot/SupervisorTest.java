package com.opspilot;

import com.opspilot.team.SupervisorDecision;
import com.opspilot.team.TeamRole;
import org.junit.jupiter.api.Test;

import static com.opspilot.team.Supervisor.parseDecision;
import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class SupervisorTest {

    @Test
    void parsesDelegationDecision() {
        SupervisorDecision decision = parseDecision("{\"next\":\"analista\",\"brief\":\"diagnostique\"}");

        assertThat(decision.isDone()).isFalse();
        assertThat(decision.next()).isEqualTo(TeamRole.ANALISTA);
        assertThat(decision.brief()).isEqualTo("diagnostique");
    }

    @Test
    void parsesDoneDecision() {
        SupervisorDecision decision = parseDecision("{\"next\":\"done\",\"brief\":\"resumo final\"}");

        assertThat(decision.isDone()).isTrue();
        assertThat(decision.brief()).isEqualTo("resumo final");
    }

    @Test
    void malformedJsonThrows() {
        assertThatThrownBy(() -> parseDecision("nao e json"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("decisao invalida do supervisor");
    }

    @Test
    void unknownRoleThrows() {
        assertThatThrownBy(() -> parseDecision("{\"next\":\"estagiario\",\"brief\":\"x\"}"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("papel desconhecido");
    }
}
