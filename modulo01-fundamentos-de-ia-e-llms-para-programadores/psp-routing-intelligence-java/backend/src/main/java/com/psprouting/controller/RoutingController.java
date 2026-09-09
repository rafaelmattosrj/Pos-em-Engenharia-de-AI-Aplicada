package com.psprouting.controller;

import com.psprouting.application.RoutingService;
import com.psprouting.application.SeedService;
import com.psprouting.domain.RoutingRecommendation;
import com.psprouting.dto.RecommendRequest;
import com.psprouting.dto.SeedResponse;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * Endpoints descritos em IDEIA.md: {@code POST /api/routing/recommend} e
 * {@code POST /api/routing/seed}.
 */
@RestController
@RequestMapping("/api/routing")
public class RoutingController {

    private final RoutingService routingService;
    private final SeedService seedService;

    public RoutingController(RoutingService routingService, SeedService seedService) {
        this.routingService = routingService;
        this.seedService = seedService;
    }

    @PostMapping("/recommend")
    public ResponseEntity<RoutingRecommendation> recommend(@Valid @RequestBody RecommendRequest request) {
        RoutingRecommendation recommendation = routingService.recommend(request.toDomain());
        return ResponseEntity.ok(recommendation);
    }

    @PostMapping("/seed")
    public ResponseEntity<SeedResponse> seed() {
        int seeded = seedService.seed();
        return ResponseEntity.ok(new SeedResponse(seeded));
    }
}
