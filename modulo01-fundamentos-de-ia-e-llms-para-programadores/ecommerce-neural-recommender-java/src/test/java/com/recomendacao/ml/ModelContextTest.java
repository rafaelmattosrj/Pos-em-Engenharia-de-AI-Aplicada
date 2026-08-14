package com.recomendacao.ml;

import com.recomendacao.data.DataLoader;
import com.recomendacao.model.Product;
import com.recomendacao.model.User;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.offset;

class ModelContextTest {

    @Test
    void normalizeMapsValueToZeroOneRange() {
        assertThat(ModelContext.normalize(50, 0, 100)).isEqualTo(0.5, offset(1e-9));
        assertThat(ModelContext.normalize(0, 0, 100)).isEqualTo(0.0, offset(1e-9));
        assertThat(ModelContext.normalize(100, 0, 100)).isEqualTo(1.0, offset(1e-9));
    }

    @Test
    void normalizeReturnsZeroWhenRangeIsZero() {
        assertThat(ModelContext.normalize(42, 10, 10)).isEqualTo(0.0);
    }

    @Test
    void dimensionsMatchCategoriesAndColorsFoundInCatalog() {
        List<Product> products = DataLoader.loadProducts();
        List<User> users = DataLoader.loadUsers();

        ModelContext context = new ModelContext(products, users);

        // catalogo real: 4 categorias (eletronicos, vestuario, calcados, acessorios)
        // e 7 cores distintas (preto, prata, azul, branco, vermelho, bege, cinza, marrom -> 8)
        long expectedCategories = products.stream().map(Product::getCategory).distinct().count();
        long expectedColors = products.stream().map(Product::getColor).distinct().count();

        assertThat(context.numCategories).isEqualTo((int) expectedCategories);
        assertThat(context.numColors).isEqualTo((int) expectedColors);
        assertThat(context.dimensions).isEqualTo(2 + context.numCategories + context.numColors);
    }

    @Test
    void minMaxAgeAndPriceAreComputedFromInputData() {
        List<Product> products = DataLoader.loadProducts();
        List<User> users = DataLoader.loadUsers();

        ModelContext context = new ModelContext(products, users);

        assertThat(context.minAge).isEqualTo(22); // Diego Almeida
        assertThat(context.maxAge).isEqualTo(30); // Camila Souza
        assertThat(context.minPrice).isEqualTo(39.99); // Bone Estiloso
        assertThat(context.maxPrice).isEqualTo(199.99); // Relogio Inteligente
    }

    @Test
    void categoriesAndColorsIndexesAreZeroBasedAndUnique() {
        ModelContext context = new ModelContext(DataLoader.loadProducts(), DataLoader.loadUsers());

        assertThat(context.categoriesIndex.values()).containsExactlyInAnyOrderElementsOf(
                java.util.stream.IntStream.range(0, context.numCategories).boxed().toList());
        assertThat(context.colorsIndex.values()).containsExactlyInAnyOrderElementsOf(
                java.util.stream.IntStream.range(0, context.numColors).boxed().toList());
    }
}
