package com.iadeva.agendamento.model;

import java.time.Instant;

public record Appointment(Instant date, String patientName, String reason, int professionalId) {}
