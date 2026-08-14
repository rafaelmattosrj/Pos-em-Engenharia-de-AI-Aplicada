package com.iadeva.cfp.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;

/**
 * Equivalente a create-event.dto.ts (class-validator: IsNotEmpty, IsString,
 * IsNumber, IsDateString).
 */
public record CreateEventDto(
        @NotBlank String nome,
        @NotBlank String endereco,
        @NotNull @Positive Integer capacidade,
        @NotBlank String data
) {}
