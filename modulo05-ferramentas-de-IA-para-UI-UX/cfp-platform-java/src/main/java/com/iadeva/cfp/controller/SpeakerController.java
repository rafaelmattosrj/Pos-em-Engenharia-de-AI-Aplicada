package com.iadeva.cfp.controller;

import com.iadeva.cfp.dto.CreateSpeakerDto;
import com.iadeva.cfp.model.Speaker;
import com.iadeva.cfp.service.SpeakerService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * Equivalente a speaker.controller.ts.
 */
@RestController
@RequestMapping("/speakers")
public class SpeakerController {

    private final SpeakerService speakerService;

    public SpeakerController(SpeakerService speakerService) {
        this.speakerService = speakerService;
    }

    @PostMapping
    public Speaker create(@Valid @RequestBody CreateSpeakerDto dto) {
        return speakerService.create(dto);
    }

    @GetMapping
    public List<Speaker> findAll() {
        return speakerService.findAll();
    }
}
