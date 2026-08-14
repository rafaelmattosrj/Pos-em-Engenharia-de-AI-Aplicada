package com.iadeva.cfp.controller;

import com.iadeva.cfp.dto.CreateEventDto;
import com.iadeva.cfp.model.Event;
import com.iadeva.cfp.service.EventService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * Equivalente a event.controller.ts.
 */
@RestController
@RequestMapping("/events")
public class EventController {

    private final EventService eventService;

    public EventController(EventService eventService) {
        this.eventService = eventService;
    }

    @PostMapping
    public Event create(@Valid @RequestBody CreateEventDto dto) {
        return eventService.create(dto);
    }

    @GetMapping
    public List<Event> findAll() {
        return eventService.findAll();
    }
}
