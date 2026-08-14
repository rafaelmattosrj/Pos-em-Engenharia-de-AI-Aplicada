package com.recomendacao.data;

import com.recomendacao.model.Product;
import com.recomendacao.model.User;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class DataLoaderTest {

    @Test
    void loadProductsReturnsTenProducts() {
        List<Product> products = DataLoader.loadProducts();

        assertThat(products).hasSize(10);
        assertThat(products).extracting(Product::getName)
                .contains("Fones de Ouvido Sem Fio", "Relogio Inteligente", "Bone Estiloso", "Mochila Executiva");
    }

    @Test
    void loadProductsAreDistinctCategoriesAndColors() {
        List<Product> products = DataLoader.loadProducts();

        assertThat(products).extracting(Product::getCategory)
                .contains("eletronicos", "vestuario", "calcados", "acessorios");
        assertThat(products).allSatisfy(p -> assertThat(p.getPrice()).isPositive());
    }

    @Test
    void loadUsersReturnsFiveUsersWithPurchaseHistory() {
        List<User> users = DataLoader.loadUsers();

        assertThat(users).hasSize(5);
        assertThat(users).extracting(User::getName)
                .containsExactly("Ana Lima", "Bruno Ferreira", "Camila Souza", "Diego Almeida", "Eduarda Nunes");
        assertThat(users).allSatisfy(u -> assertThat(u.getPurchases()).isNotEmpty());
    }

    @Test
    void anaLimaHasElectronicsPurchaseHistory() {
        User ana = DataLoader.loadUsers().get(0);

        assertThat(ana.getAge()).isEqualTo(25);
        assertThat(ana.getPurchases()).extracting(Product::getCategory)
                .containsOnly("eletronicos");
    }
}
