package com.recomendacao.model;

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

class ProductUserTest {

    @Test
    void productExposesAllFieldsViaGetters() {
        Product product = new Product(1, "Fones de Ouvido Sem Fio", "eletronicos", 129.99, "preto");

        assertThat(product.getId()).isEqualTo(1);
        assertThat(product.getName()).isEqualTo("Fones de Ouvido Sem Fio");
        assertThat(product.getCategory()).isEqualTo("eletronicos");
        assertThat(product.getPrice()).isEqualTo(129.99);
        assertThat(product.getColor()).isEqualTo("preto");
        assertThat(product.toString()).contains("Fones de Ouvido Sem Fio", "eletronicos");
    }

    @Test
    void userExposesAllFieldsAndPurchases() {
        Product p = new Product(1, "Fones", "eletronicos", 10.0, "preto");
        User user = new User(99, "Rafael Souza", 27, List.of(p));

        assertThat(user.getId()).isEqualTo(99);
        assertThat(user.getName()).isEqualTo("Rafael Souza");
        assertThat(user.getAge()).isEqualTo(27);
        assertThat(user.getPurchases()).containsExactly(p);
        assertThat(user.toString()).contains("Rafael Souza", "purchases=1");
    }

    @Test
    void newUserWithoutPurchaseHistory() {
        User newUser = new User(99, "Rafael Souza", 27, List.of());

        assertThat(newUser.getPurchases()).isEmpty();
    }
}
