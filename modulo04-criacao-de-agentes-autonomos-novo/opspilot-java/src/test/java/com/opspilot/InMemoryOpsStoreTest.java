package com.opspilot;

import com.opspilot.domain.Alert;
import com.opspilot.domain.AlertStatus;
import com.opspilot.domain.Exceptions;
import com.opspilot.domain.IncidentStatus;
import com.opspilot.domain.Runbook;
import com.opspilot.domain.Severity;
import com.opspilot.store.InMemoryOpsStore;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class InMemoryOpsStoreTest {

    @Test
    void createThenResolveIncident() {
        InMemoryOpsStore store = new InMemoryOpsStore();
        var incident = store.createIncident("Checkout fora do ar", "checkout-api", Severity.CRITICAL);

        assertThat(incident.status()).isEqualTo(IncidentStatus.OPEN);
        assertThat(store.getIncidents(IncidentStatus.OPEN)).hasSize(1);

        var resolved = store.resolveIncident(incident.id(), "Rollback aplicado");
        assertThat(resolved.status()).isEqualTo(IncidentStatus.RESOLVED);
        assertThat(store.getIncidents(IncidentStatus.OPEN)).isEmpty();
        assertThat(store.getIncidents(IncidentStatus.RESOLVED)).hasSize(1);
    }

    @Test
    void resolvingUnknownIncidentThrows() {
        InMemoryOpsStore store = new InMemoryOpsStore();
        assertThatThrownBy(() -> store.resolveIncident("nao-existe", null))
                .isInstanceOf(Exceptions.IncidentNotFoundException.class);
    }

    @Test
    void filtersAlertsByStatus() {
        InMemoryOpsStore store = new InMemoryOpsStore();
        store.seedAlert(new Alert("a1", "svc", "desc", Severity.HIGH, AlertStatus.FIRING));
        store.seedAlert(new Alert("a2", "svc", "desc", Severity.LOW, AlertStatus.RESOLVED));

        assertThat(store.getAlerts(AlertStatus.FIRING)).hasSize(1);
        assertThat(store.getAlerts(null)).hasSize(2);
    }

    @Test
    void unknownRunbookThrows() {
        InMemoryOpsStore store = new InMemoryOpsStore();
        store.seedRunbook(new Runbook("checkout-api", "conteudo"));

        assertThat(store.getRunbook("checkout-api").content()).isEqualTo("conteudo");
        assertThatThrownBy(() -> store.getRunbook("outro-servico"))
                .isInstanceOf(Exceptions.RunbookNotFoundException.class);
    }
}
