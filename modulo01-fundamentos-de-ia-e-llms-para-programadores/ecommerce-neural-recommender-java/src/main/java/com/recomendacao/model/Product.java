package com.recomendacao.model;

/**
 * Representa um produto do catalogo do e-commerce.
 * Espelho do products.json do exemplo original.
 */
public class Product {

    private int id;
    private String name;
    private String category;
    private double price;
    private String color;

    public Product(int id, String name, String category, double price, String color) {
        this.id = id;
        this.name = name;
        this.category = category;
        this.price = price;
        this.color = color;
    }

    public int getId() { return id; }
    public String getName() { return name; }
    public String getCategory() { return category; }
    public double getPrice() { return price; }
    public String getColor() { return color; }

    @Override
    public String toString() {
        return String.format("Product{id=%d, name='%s', category='%s', price=%.2f, color='%s'}",
                id, name, category, price, color);
    }
}
