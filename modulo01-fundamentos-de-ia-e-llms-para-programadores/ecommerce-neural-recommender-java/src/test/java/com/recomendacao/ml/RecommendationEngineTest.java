package com.recomendacao.ml;

import com.recomendacao.data.DataLoader;
import com.recomendacao.model.Product;
import com.recomendacao.model.User;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInstance;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

/**
 * Cobre os mesmos cenarios demonstrados em Main.java / no README do porte Go:
 *  - recomendacoes para usuarios cadastrados com historico de compras
 *  - usuario novo sem historico (id=99, Rafael Souza)
 *  - usuario com historico so de acessorios (id=100, Marina Costa)
 */
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class RecommendationEngineTest {

    private List<Product> products;
    private List<User> users;
    private RecommendationEngine engine;

    @BeforeAll
    void trainEngineOnce() {
        products = DataLoader.loadProducts();
        users = DataLoader.loadUsers();

        engine = new RecommendationEngine();
        engine.trainModel(products, users);
    }

    @Test
    void encodeProductProducesVectorWithConfiguredDimensions() {
        double[] vector = engine.encodeProduct(products.get(0));

        assertThat(vector).hasSize(2 + 4 + 8); // price + age + 4 categorias + 8 cores
    }

    @Test
    void encodeUserWithoutPurchasesOnlyEncodesAge() {
        User newUser = new User(99, "Rafael Souza", 27, List.of());

        double[] vector = engine.encodeUser(newUser);

        // Apenas a posicao 1 (idade) deve ser diferente de zero
        assertThat(vector[0]).isZero();
        assertThat(vector[1]).isNotZero();
        for (int i = 2; i < vector.length; i++) {
            assertThat(vector[i]).isZero();
        }
    }

    @Test
    void encodeUserWithPurchasesAveragesProductVectors() {
        User ana = users.get(0); // Ana Lima: Fones + Relogio (eletronicos)

        double[] vector = engine.encodeUser(ana);

        assertThat(vector).hasSize(2 + 4 + 8);
        // Deve haver peso de categoria eletronicos (nao pode ser todo zero)
        boolean hasNonZero = false;
        for (double v : vector) {
            if (v != 0) hasNonZero = true;
        }
        assertThat(hasNonZero).isTrue();
    }

    @Test
    void recommendBeforeTrainingThrowsIllegalStateException() {
        RecommendationEngine untrainedEngine = new RecommendationEngine();
        User user = new User(1, "Teste", 30, List.of());

        assertThatThrownBy(() -> untrainedEngine.recommend(user))
                .isInstanceOf(IllegalStateException.class)
                .hasMessageContaining("nao treinado");
    }

    @Test
    void recommendReturnsAllProductsSortedByDescendingScore() {
        User ana = users.get(0);

        List<RecommendationEngine.Recommendation> recommendations = engine.recommend(ana);

        assertThat(recommendations).hasSize(products.size());
        for (int i = 0; i < recommendations.size() - 1; i++) {
            assertThat(recommendations.get(i).score())
                    .isGreaterThanOrEqualTo(recommendations.get(i + 1).score());
        }
        assertThat(recommendations).allSatisfy(r -> assertThat(r.score()).isBetween(0.0, 1.0));
    }

    @Test
    void recommendWorksForNewUserWithoutPurchaseHistory() {
        User newUser = new User(99, "Rafael Souza", 27, List.of());

        List<RecommendationEngine.Recommendation> recommendations = engine.recommend(newUser);

        assertThat(recommendations).hasSize(products.size());
        assertThat(recommendations).allSatisfy(r -> assertThat(r.score()).isBetween(0.0, 1.0));
    }

    @Test
    void recommendWorksForUserWithAccessoriesOnlyHistory() {
        User accessoriesUser = new User(100, "Marina Costa", 29, List.of(
                products.get(7), // Bone Estiloso
                products.get(8)  // Mochila Executiva
        ));

        List<RecommendationEngine.Recommendation> recommendations = engine.recommend(accessoriesUser);

        assertThat(recommendations).hasSize(products.size());
        assertThat(recommendations).extracting(r -> r.product().getName())
                .containsExactlyInAnyOrderElementsOf(products.stream().map(Product::getName).toList());
    }

    @Test
    void recommendationToStringContainsProductDetails() {
        RecommendationEngine.Recommendation rec = new RecommendationEngine.Recommendation(products.get(0), 0.75);

        assertThat(rec.toString())
                .contains("75")
                .contains(products.get(0).getName())
                .contains(products.get(0).getCategory());
    }
}
