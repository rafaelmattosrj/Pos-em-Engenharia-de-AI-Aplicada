package com.recomendacao;

import com.recomendacao.data.DataLoader;
import com.recomendacao.ml.RecommendationEngine;
import com.recomendacao.model.Product;
import com.recomendacao.model.User;

import java.util.List;

/**
 * Ponto de entrada do sistema de recomendacao de e-commerce.
 *
 * Traducao do Exemplo 01 (JavaScript/TensorFlow.js) para Java puro.
 *
 * Fluxo principal:
 *  1. Carregar produtos e usuarios
 *  2. Treinar a rede neural
 *  3. Gerar e exibir recomendacoes para cada usuario
 *  4. Testar com um usuario novo (sem historico de compras)
 */
public class Main {

    public static void main(String[] args) {
        System.out.println("=====================================================");
        System.out.println("  Sistema de Recomendacao de E-commerce - Java");
        System.out.println("  (Traducao do Exemplo 01 - JavaScript/TensorFlow.js)");
        System.out.println("=====================================================\n");

        // ---- 1. Carregar dados ----
        List<Product> products = DataLoader.loadProducts();
        List<User>    users    = DataLoader.loadUsers();

        System.out.printf("Produtos carregados: %d%n", products.size());
        System.out.printf("Usuarios carregados: %d%n%n", users.size());

        // ---- 2. Treinar o modelo ----
        RecommendationEngine engine = new RecommendationEngine();
        engine.trainModel(products, users);

        // ---- 3. Recomendar para cada usuario ----
        System.out.println("=====================================================");
        System.out.println("  RECOMENDACOES POR USUARIO");
        System.out.println("=====================================================");

        for (User user : users) {
            printRecommendations(engine, user, 3);
        }

        // ---- 4. Testar com usuario novo (sem historico) ----
        System.out.println("=====================================================");
        System.out.println("  USUARIO NOVO (sem historico de compras)");
        System.out.println("=====================================================");

        User newUser = new User(99, "Rafael Souza", 27, List.of());
        printRecommendations(engine, newUser, 3);

        // ---- 5. Testar com usuario que comprou acessorios ----
        System.out.println("=====================================================");
        System.out.println("  USUARIO COM HISTORICO DE ACESSORIOS");
        System.out.println("=====================================================");

        User accessoriesUser = new User(100, "Marina Costa", 29, List.of(
            products.get(7), // Bone Estiloso
            products.get(8)  // Mochila Executiva
        ));
        printRecommendations(engine, accessoriesUser, 3);
    }

    private static void printRecommendations(
            RecommendationEngine engine,
            User user,
            int topN) {

        System.out.printf("%n Usuario: %s (idade %d)%n", user.getName(), user.getAge());

        if (user.getPurchases().isEmpty()) {
            System.out.println("  Historico: nenhuma compra");
        } else {
            System.out.print("  Historico: ");
            System.out.println(user.getPurchases().stream()
                    .map(p -> p.getName())
                    .reduce((a, b) -> a + ", " + b)
                    .orElse("-"));
        }

        System.out.printf("  Top %d Recomendacoes:%n", topN);
        engine.recommend(user)
              .stream()
              .limit(topN)
              .forEach(rec -> System.out.println("  " + rec));
    }
}
