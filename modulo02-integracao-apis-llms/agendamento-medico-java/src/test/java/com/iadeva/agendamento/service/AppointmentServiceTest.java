package com.iadeva.agendamento.service;

import com.iadeva.agendamento.model.Appointment;
import org.junit.jupiter.api.Test;

import java.time.Instant;
import java.time.temporal.ChronoUnit;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class AppointmentServiceTest {

    private final AppointmentService service = new AppointmentService();
    private final Instant date = Instant.parse("2030-01-01T10:00:00Z").truncatedTo(ChronoUnit.SECONDS);

    @Test
    void bookAppointment_sucesso() {
        Appointment appointment = service.bookAppointment(1, date, "Rafael Souza", "consulta de rotina");

        assertThat(appointment.patientName()).isEqualTo("Rafael Souza");
        assertThat(service.checkAvailability(1, date)).isFalse();
    }

    @Test
    void bookAppointment_horarioIndisponivelLancaExcecao() {
        service.bookAppointment(1, date, "Paciente A", "motivo");

        assertThatThrownBy(() -> service.bookAppointment(1, date, "Paciente B", "outro motivo"))
                .isInstanceOf(IllegalStateException.class)
                .hasMessage("Horário indisponível para este profissional");
    }

    @Test
    void cancelAppointment_sucesso() {
        service.bookAppointment(1, date, "Rafael Souza", "consulta");

        service.cancelAppointment(1, "rafael souza", date);

        assertThat(service.checkAvailability(1, date)).isTrue();
    }

    @Test
    void cancelAppointment_naoEncontradoLancaExcecao() {
        assertThatThrownBy(() -> service.cancelAppointment(1, "ninguem", date))
                .isInstanceOf(IllegalStateException.class)
                .hasMessage("Agendamento não encontrado para cancelamento");
    }

    @Test
    void getProfessionals_retornaOsTresProfissionaisFixos() {
        assertThat(service.getProfessionals()).hasSize(3);
    }
}
