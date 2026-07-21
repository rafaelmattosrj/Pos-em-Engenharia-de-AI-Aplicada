package com.iadeva.agendamento.service;

import com.iadeva.agendamento.model.Appointment;
import com.iadeva.agendamento.model.Professional;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;

// Serviço de agendamentos em memória — equivalente ao appointmentService.ts
@Service
public class AppointmentService {

    public static final List<Professional> PROFESSIONALS = List.of(
            new Professional(1, "Dr. Alicio da Silva", "Cardiologia"),
            new Professional(2, "Dra. Ana Pereira", "Dermatologia"),
            new Professional(3, "Dra. Carol Gomes", "Neurologia")
    );

    private final List<Appointment> appointments = new ArrayList<>();

    public AppointmentService() {
        // Pre-populate with sample appointments
        Instant todayAt11 = Instant.now().truncatedTo(ChronoUnit.DAYS).plus(11, ChronoUnit.HOURS);
        Instant tomorrowAt14 = todayAt11.plus(1, ChronoUnit.DAYS).minus(11, ChronoUnit.HOURS).plus(14, ChronoUnit.HOURS);

        appointments.add(new Appointment(todayAt11, "Joao da Silva", "check-up regular", 1));
        appointments.add(new Appointment(tomorrowAt14, "Luana Costa", "Erupção cutânea", 2));
    }

    public boolean checkAvailability(int professionalId, Instant date) {
        return appointments.stream()
                .noneMatch(a -> a.professionalId() == professionalId && a.date().equals(date));
    }

    public Appointment bookAppointment(int professionalId, Instant date, String patientName, String reason) {
        if (!checkAvailability(professionalId, date)) {
            throw new IllegalStateException("Horário indisponível para este profissional");
        }
        Appointment newAppointment = new Appointment(date, patientName, reason, professionalId);
        appointments.add(newAppointment);
        return newAppointment;
    }

    public void cancelAppointment(int professionalId, String patientName, Instant date) {
        Appointment found = appointments.stream()
                .filter(a -> a.professionalId() == professionalId
                        && a.date().equals(date)
                        && a.patientName().equalsIgnoreCase(patientName))
                .findFirst()
                .orElseThrow(() -> new IllegalStateException("Agendamento não encontrado para cancelamento"));
        appointments.remove(found);
    }

    public List<Professional> getProfessionals() {
        return PROFESSIONALS;
    }
}
