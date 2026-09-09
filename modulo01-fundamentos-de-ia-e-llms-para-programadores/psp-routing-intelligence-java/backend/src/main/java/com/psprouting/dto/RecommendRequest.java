package com.psprouting.dto;

import com.psprouting.domain.PaymentMethod;
import com.psprouting.domain.Transaction;
import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;

import java.math.BigDecimal;

/**
 * Corpo de {@code POST /api/routing/recommend}, exatamente no shape do
 * exemplo em IDEIA.md:
 *
 * <pre>{@code
 * {
 *   "amount": 1500.00,
 *   "method": "CREDIT_CARD",
 *   "brand": "VISA",
 *   "userRegion": "SP",
 *   "hour": 20,
 *   "merchantCategory": "STREAMING"
 * }
 * }</pre>
 *
 * {@code brand} nao e obrigatorio — PIX e WALLET normalmente nao tem
 * bandeira de cartao.
 */
public record RecommendRequest(
        @NotNull @Positive BigDecimal amount,
        @NotNull PaymentMethod method,
        String brand,
        @NotBlank String userRegion,
        @NotNull @Min(0) @Max(23) Integer hour,
        @NotBlank String merchantCategory
) {
    public Transaction toDomain() {
        return new Transaction(amount, method, brand, userRegion, hour, merchantCategory);
    }
}
