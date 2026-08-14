package com.iadeva.cfp.model;

/**
 * Equivalente à interface SpeakerDTO de shared-types/src/lib/speaker.dto.ts.
 */
public record Speaker(String id, String name, String email, String talkTitle, boolean isGDE) {}
