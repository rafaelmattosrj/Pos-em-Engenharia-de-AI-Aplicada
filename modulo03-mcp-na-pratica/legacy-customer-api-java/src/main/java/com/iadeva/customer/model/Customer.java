package com.iadeva.customer.model;

import java.util.UUID;

/**
 * Entidade principal do domínio.
 * Equivalente à interface/tipo Customer nos módulos 06 e 07 do curso JS/TS.
 * Usa Java record (imutável por padrão) — adequado para dados simples de domínio.
 */
public record Customer(UUID id, String name, String phone) {

    /**
     * Cria um Customer com um novo UUID gerado automaticamente.
     */
    public static Customer create(String name, String phone) {
        return new Customer(UUID.randomUUID(), name, phone);
    }

    /**
     * Retorna um Customer com os dados atualizados, mantendo o mesmo id.
     */
    public Customer withUpdated(String name, String phone) {
        return new Customer(
                this.id,
                name != null ? name : this.name,
                phone != null ? phone : this.phone
        );
    }
}
