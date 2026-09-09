package com.psprouting.controller;

import com.psprouting.application.InvalidLlmResponseException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Traduz falhas do pipeline em respostas HTTP consistentes:
 * <ul>
 *   <li>400 — corpo da requisicao invalido (Bean Validation)</li>
 *   <li>502 — o LLM respondeu algo que nao pode ser interpretado como a
 *       recomendacao esperada (a dependencia externa falhou, nao o cliente)</li>
 * </ul>
 */
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<Map<String, Object>> handleValidation(MethodArgumentNotValidException ex) {
        Map<String, String> fieldErrors = new LinkedHashMap<>();
        ex.getBindingResult().getFieldErrors().forEach(
                error -> fieldErrors.put(error.getField(), error.getDefaultMessage())
        );

        Map<String, Object> body = new LinkedHashMap<>();
        body.put("error", "VALIDATION_ERROR");
        body.put("fields", fieldErrors);
        return ResponseEntity.badRequest().body(body);
    }

    @ExceptionHandler(InvalidLlmResponseException.class)
    public ResponseEntity<Map<String, Object>> handleInvalidLlmResponse(InvalidLlmResponseException ex) {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("error", "INVALID_LLM_RESPONSE");
        body.put("message", ex.getMessage());
        return ResponseEntity.status(HttpStatus.BAD_GATEWAY).body(body);
    }
}
