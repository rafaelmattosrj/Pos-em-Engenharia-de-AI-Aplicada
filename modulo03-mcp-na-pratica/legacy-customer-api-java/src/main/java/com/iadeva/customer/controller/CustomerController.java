package com.iadeva.customer.controller;

import com.iadeva.customer.model.Customer;
import com.iadeva.customer.service.CustomerService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * Controller principal de clientes + endpoint de health check.
 * Equivalente ao customerRouter.ts dos módulos 06 e 07 do curso JS/TS.
 *
 * RBAC (aplicado via SecurityConfig):
 * - GET    /customers       → MEMBER + ADMIN
 * - GET    /customers/{id}  → MEMBER + ADMIN
 * - POST   /customers       → ADMIN apenas
 * - PUT    /customers/{id}  → ADMIN apenas
 * - DELETE /customers/{id}  → ADMIN apenas
 * - GET    /health          → público
 */
@RestController
public class CustomerController {

    private final CustomerService customerService;

    public CustomerController(CustomerService customerService) {
        this.customerService = customerService;
    }

    // -------------------------------------------------------------------------
    // Health check
    // -------------------------------------------------------------------------

    /**
     * GET /health
     * Retorna status da aplicação — sem autenticação.
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of(
                "status", "ok",
                "service", "legacy-customer-api-java"
        ));
    }

    // -------------------------------------------------------------------------
    // CRUD de Customers
    // -------------------------------------------------------------------------

    /**
     * GET /customers
     * Retorna lista de todos os clientes.
     */
    @GetMapping("/customers")
    public ResponseEntity<List<Customer>> listAll() {
        return ResponseEntity.ok(customerService.findAll());
    }

    /**
     * GET /customers/{id}
     * Retorna um cliente específico ou 404.
     */
    @GetMapping("/customers/{id}")
    public ResponseEntity<?> getById(@PathVariable String id) {
        try {
            UUID uuid = UUID.fromString(id);
            return customerService.findById(uuid)
                    .<ResponseEntity<?>>map(ResponseEntity::ok)
                    .orElseGet(() -> ResponseEntity
                            .status(HttpStatus.NOT_FOUND)
                            .body(Map.of("error", "Cliente não encontrado: " + id)));
        } catch (IllegalArgumentException e) {
            return ResponseEntity
                    .status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "UUID inválido: " + id));
        }
    }

    /**
     * POST /customers
     * Cria um novo cliente. Body: { "name": "...", "phone": "..." }
     * Requer role ADMIN.
     */
    @PostMapping("/customers")
    public ResponseEntity<?> create(@RequestBody Map<String, String> body) {
        String name = body.get("name");
        String phone = body.get("phone");

        if (name == null || name.isBlank()) {
            return ResponseEntity
                    .status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Campo 'name' é obrigatório"));
        }
        if (phone == null || phone.isBlank()) {
            return ResponseEntity
                    .status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Campo 'phone' é obrigatório"));
        }

        Customer created = customerService.create(name, phone);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    /**
     * PUT /customers/{id}
     * Atualiza nome e/ou telefone. Retorna 404 se não encontrado.
     * Requer role ADMIN.
     */
    @PutMapping("/customers/{id}")
    public ResponseEntity<?> update(@PathVariable String id,
                                    @RequestBody Map<String, String> body) {
        try {
            UUID uuid = UUID.fromString(id);
            String name  = body.get("name");
            String phone = body.get("phone");

            return customerService.update(uuid, name, phone)
                    .<ResponseEntity<?>>map(ResponseEntity::ok)
                    .orElseGet(() -> ResponseEntity
                            .status(HttpStatus.NOT_FOUND)
                            .body(Map.of("error", "Cliente não encontrado: " + id)));
        } catch (IllegalArgumentException e) {
            return ResponseEntity
                    .status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "UUID inválido: " + id));
        }
    }

    /**
     * DELETE /customers/{id}
     * Remove o cliente. Retorna 204 se removido, 404 se não encontrado.
     * Requer role ADMIN.
     */
    @DeleteMapping("/customers/{id}")
    public ResponseEntity<?> delete(@PathVariable String id) {
        try {
            UUID uuid = UUID.fromString(id);
            if (customerService.delete(uuid)) {
                return ResponseEntity.noContent().build();
            }
            return ResponseEntity
                    .status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Cliente não encontrado: " + id));
        } catch (IllegalArgumentException e) {
            return ResponseEntity
                    .status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "UUID inválido: " + id));
        }
    }
}
