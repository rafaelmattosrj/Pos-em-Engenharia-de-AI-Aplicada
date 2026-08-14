package com.iadeva.cfp.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

/**
 * Equivalente a create-speaker.dto.ts (class-validator: IsNotEmpty,
 * IsString, IsEmail, IsBoolean).
 */
public record CreateSpeakerDto(
        @NotBlank String name,
        @NotBlank @Email String email,
        @NotBlank String talkTitle,
        @NotNull Boolean isGDE
) {}
