package com.recomendacao.model;

import java.util.List;

/**
 * Representa um usuario com seu historico de compras.
 * Espelho do users.json do exemplo original.
 */
public class User {

    private int id;
    private String name;
    private int age;
    private List<Product> purchases;

    public User(int id, String name, int age, List<Product> purchases) {
        this.id = id;
        this.name = name;
        this.age = age;
        this.purchases = purchases;
    }

    public int getId() { return id; }
    public String getName() { return name; }
    public int getAge() { return age; }
    public List<Product> getPurchases() { return purchases; }

    @Override
    public String toString() {
        return String.format("User{id=%d, name='%s', age=%d, purchases=%d}",
                id, name, age, purchases.size());
    }
}
