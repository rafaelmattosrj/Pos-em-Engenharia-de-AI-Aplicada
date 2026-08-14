package com.iadeva.cfp.service;

import com.iadeva.cfp.dto.CreateEventDto;
import com.iadeva.cfp.model.Event;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Equivalente a event.service.ts — armazenamento em memória, sem
 * persistência (mesmo estágio "minimal backend" da versão NestJS original).
 */
@Service
public class EventService {

    private final List<Event> events = new CopyOnWriteArrayList<>();

    public Event create(CreateEventDto dto) {
        Event event = new Event(generateId(), dto.nome(), dto.endereco(), dto.capacidade(), dto.data());
        events.add(event);
        return event;
    }

    public List<Event> findAll() {
        return List.copyOf(events);
    }

    /**
     * Gera um id curto e aleatório — equivalente a
     * Math.random().toString(36).substring(2, 9) do TypeScript original
     * (mesmo propósito: id único e curto, sem garantias criptográficas).
     */
    private String generateId() {
        return UUID.randomUUID().toString().replace("-", "").substring(0, 7);
    }
}
