package com.recomendacao.ml;

import com.recomendacao.model.Product;
import com.recomendacao.model.User;

import java.util.*;

/**
 * Contexto do modelo: contém os metadados de codificacao calculados a partir
 * dos dados de treino. Equivalente ao objeto retornado por makeContext() no JS.
 *
 * Guarda:
 *  - Limites de normalizacao (minPrice, maxPrice, minAge, maxAge)
 *  - Indices para one-hot encoding de categorias e cores
 *  - Idade media normalizada por produto (para personalizacao)
 *  - Vetores pre-calculados de cada produto
 */
public class ModelContext {

    // Pesos de cada feature na composicao do vetor
    // (espelho do objeto WEIGHTS do JS)
    public static final double WEIGHT_CATEGORY = 0.4;
    public static final double WEIGHT_COLOR    = 0.3;
    public static final double WEIGHT_PRICE    = 0.2;
    public static final double WEIGHT_AGE      = 0.1;

    public final List<Product> products;
    public final List<User>    users;

    public final double minAge;
    public final double maxAge;
    public final double minPrice;
    public final double maxPrice;

    // Mapeamento: nome -> indice para one-hot encoding
    public final Map<String, Integer> categoriesIndex;
    public final Map<String, Integer> colorsIndex;

    public final int numCategories;
    public final int numColors;

    // dimensoes = price(1) + age(1) + categorias + cores
    public final int dimensions;

    // Idade media normalizada de compradores por produto
    // (ajuda a personalizar por faixa etaria)
    public final Map<String, Double> productAvgAgeNorm;

    // Vetores pre-calculados de cada produto para evitar reprocessamento
    // durante a recomendacao
    public double[][] productVectors;

    public ModelContext(List<Product> products, List<User> users) {
        this.products = products;
        this.users    = users;

        // --- Calcular limites de normalizacao ---
        this.minAge   = users.stream().mapToInt(User::getAge).min().orElse(0);
        this.maxAge   = users.stream().mapToInt(User::getAge).max().orElse(100);
        this.minPrice = products.stream().mapToDouble(Product::getPrice).min().orElse(0);
        this.maxPrice = products.stream().mapToDouble(Product::getPrice).max().orElse(1000);

        // --- Construir indices para one-hot encoding ---
        List<String> categories = products.stream()
                .map(Product::getCategory).distinct().sorted().toList();
        List<String> colors = products.stream()
                .map(Product::getColor).distinct().sorted().toList();

        this.categoriesIndex = new LinkedHashMap<>();
        for (int i = 0; i < categories.size(); i++) categoriesIndex.put(categories.get(i), i);

        this.colorsIndex = new LinkedHashMap<>();
        for (int i = 0; i < colors.size(); i++) colorsIndex.put(colors.get(i), i);

        this.numCategories = categories.size();
        this.numColors     = colors.size();
        this.dimensions    = 2 + numCategories + numColors;

        // --- Calcular idade media normalizada por produto ---
        Map<String, Double> ageSums   = new HashMap<>();
        Map<String, Integer> ageCounts = new HashMap<>();

        for (User user : users) {
            for (Product p : user.getPurchases()) {
                ageSums.merge(p.getName(), (double) user.getAge(), Double::sum);
                ageCounts.merge(p.getName(), 1, Integer::sum);
            }
        }

        double midAge = (minAge + maxAge) / 2.0;
        this.productAvgAgeNorm = new HashMap<>();
        for (Product product : products) {
            double avg = ageCounts.containsKey(product.getName())
                    ? ageSums.get(product.getName()) / ageCounts.get(product.getName())
                    : midAge;
            productAvgAgeNorm.put(product.getName(), normalize(avg, minAge, maxAge));
        }
    }

    /**
     * Normaliza um valor para o intervalo [0, 1].
     * Formula: (val - min) / (max - min)
     * Equivalente a normalize() do JS.
     */
    public static double normalize(double value, double min, double max) {
        double range = max - min;
        return range == 0 ? 0 : (value - min) / range;
    }
}
