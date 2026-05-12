package com.iadeva.customer.service;

import com.iadeva.customer.model.Customer;
import org.springframework.stereotype.Service;

import java.util.*;

/**
 * Serviço de gerenciamento de clientes em memória.
 * Equivalente ao customerStore / array de clientes nos módulos 06 e 07 do curso JS/TS.
 * Usa HashMap para O(1) em busca por ID e ArrayList para listagem ordenada.
 */
@Service
public class CustomerService {

    private final Map<UUID, Customer> store = new LinkedHashMap<>();

    public CustomerService() {
        // 5 clientes pré-cadastrados para demonstração
        seed("Alice Silva",    "(11) 91234-5678");
        seed("Bruno Oliveira", "(21) 98765-4321");
        seed("Carla Souza",    "(31) 97654-3210");
        seed("Diego Lima",     "(41) 96543-2109");
        seed("Eva Costa",      "(51) 95432-1098");
    }

    private void seed(String name, String phone) {
        Customer c = Customer.create(name, phone);
        store.put(c.id(), c);
    }

    /** Retorna todos os clientes. */
    public List<Customer> findAll() {
        return List.copyOf(store.values());
    }

    /** Retorna um cliente por ID ou Optional vazio se não encontrado. */
    public Optional<Customer> findById(UUID id) {
        return Optional.ofNullable(store.get(id));
    }

    /** Cria um novo cliente e armazena em memória. */
    public Customer create(String name, String phone) {
        Customer customer = Customer.create(name, phone);
        store.put(customer.id(), customer);
        return customer;
    }

    /**
     * Atualiza um cliente existente.
     * Retorna o cliente atualizado ou Optional vazio se ID não encontrado.
     */
    public Optional<Customer> update(UUID id, String name, String phone) {
        return findById(id).map(existing -> {
            Customer updated = existing.withUpdated(name, phone);
            store.put(id, updated);
            return updated;
        });
    }

    /**
     * Remove um cliente por ID.
     * Retorna true se removido, false se não encontrado.
     */
    public boolean delete(UUID id) {
        return store.remove(id) != null;
    }
}
