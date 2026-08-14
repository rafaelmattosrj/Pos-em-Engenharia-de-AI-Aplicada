package com.iadeva.cfp.service;

import com.iadeva.cfp.dto.CreateSpeakerDto;
import com.iadeva.cfp.model.Speaker;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Equivalente a speaker.service.ts — armazenamento em memória, sem
 * persistência.
 */
@Service
public class SpeakerService {

    private final List<Speaker> speakers = new CopyOnWriteArrayList<>();

    public Speaker create(CreateSpeakerDto dto) {
        Speaker speaker = new Speaker(generateId(), dto.name(), dto.email(), dto.talkTitle(), dto.isGDE());
        speakers.add(speaker);
        return speaker;
    }

    public List<Speaker> findAll() {
        return List.copyOf(speakers);
    }

    private String generateId() {
        return UUID.randomUUID().toString().replace("-", "").substring(0, 7);
    }
}
