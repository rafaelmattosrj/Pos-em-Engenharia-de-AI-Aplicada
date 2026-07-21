package com.recomendacao.data;

import com.recomendacao.model.Product;
import com.recomendacao.model.User;

import java.util.List;

/**
 * Carrega os dados de produtos e usuarios.
 * Equivalente aos arquivos data/products.json e data/users.json do original.
 */
public class DataLoader {

    public static List<Product> loadProducts() {
        return List.of(
            new Product(1,  "Fones de Ouvido Sem Fio",  "eletronicos", 129.99, "preto"),
            new Product(2,  "Relogio Inteligente",       "eletronicos", 199.99, "prata"),
            new Product(3,  "Caixa de Som Bluetooth",    "eletronicos",  89.99, "azul"),
            new Product(4,  "Camiseta Estampada",        "vestuario",    49.99, "branco"),
            new Product(5,  "Calca Jeans Slim",          "vestuario",    99.99, "azul"),
            new Product(6,  "Tenis Esportivo",           "calcados",    149.99, "vermelho"),
            new Product(7,  "Sandalia Casual",           "calcados",     69.99, "bege"),
            new Product(8,  "Bone Estiloso",             "acessorios",   39.99, "preto"),
            new Product(9,  "Mochila Executiva",         "acessorios",  159.99, "cinza"),
            new Product(10, "Oculos de Sol",             "acessorios",   89.99, "marrom")
        );
    }

    public static List<User> loadUsers() {
        List<Product> products = loadProducts();

        // Ana Lima - gosta de eletronicos
        User ana = new User(1, "Ana Lima", 25, List.of(
            products.get(0), // Fones de Ouvido
            products.get(1)  // Relogio Inteligente
        ));

        // Bruno Ferreira - gosta de eletronicos
        User bruno = new User(2, "Bruno Ferreira", 27, List.of(
            products.get(0), // Fones de Ouvido
            products.get(2)  // Caixa de Som
        ));

        // Camila Souza - gosta de vestuario
        User camila = new User(3, "Camila Souza", 30, List.of(
            products.get(3), // Camiseta
            products.get(4)  // Calca Jeans
        ));

        // Diego Almeida - misto: eletronicos + calcados
        User diego = new User(4, "Diego Almeida", 22, List.of(
            products.get(1), // Relogio
            products.get(2), // Caixa de Som
            products.get(5)  // Tenis
        ));

        // Eduarda Nunes - misto: eletronicos + calcados + vestuario
        User eduarda = new User(5, "Eduarda Nunes", 28, List.of(
            products.get(0), // Fones de Ouvido
            products.get(5), // Tenis
            products.get(4)  // Calca Jeans
        ));

        return List.of(ana, bruno, camila, diego, eduarda);
    }
}
