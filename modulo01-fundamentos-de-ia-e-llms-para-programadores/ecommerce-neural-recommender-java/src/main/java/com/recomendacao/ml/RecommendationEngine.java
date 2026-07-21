package com.recomendacao.ml;

import com.recomendacao.model.Product;
import com.recomendacao.model.User;

import java.util.*;

/**
 * Motor de recomendacao. Orquestra:
 *  1. Criacao do contexto (makeContext)
 *  2. Codificacao de features (encodeProduct / encodeUser)
 *  3. Criacao dos dados de treino (createTrainingData)
 *  4. Treinamento da rede neural (trainModel)
 *  5. Geracao de recomendacoes (recommend)
 *
 * Equivalente ao modelTrainingWorker.js do exemplo original.
 */
public class RecommendationEngine {

    private NeuralNetwork model;
    private ModelContext context;

    // ====================================================================
    // 1. Inicializar o contexto e treinar o modelo
    // ====================================================================

    public void trainModel(List<Product> products, List<User> users) {
        System.out.println("Construindo contexto de codificacao...");
        this.context = new ModelContext(products, users);

        System.out.printf("Categorias encontradas: %s%n", context.categoriesIndex.keySet());
        System.out.printf("Cores encontradas:      %s%n", context.colorsIndex.keySet());
        System.out.printf("Dimensoes do vetor:     %d (price + age + %d categorias + %d cores)%n%n",
                context.dimensions, context.numCategories, context.numColors);

        // Pre-calcular vetores dos produtos para uso na recomendacao
        context.productVectors = new double[products.size()][];
        for (int i = 0; i < products.size(); i++) {
            context.productVectors[i] = encodeProduct(products.get(i));
        }

        // Criar dados de treino
        TrainingData trainData = createTrainingData();

        // Configurar a rede neural: [inputDim, 128, 64, 32, 1]
        // Mesma arquitetura do exemplo JS
        int inputDim = context.dimensions * 2; // user vector + product vector
        this.model = new NeuralNetwork(
                new int[]{ inputDim, 128, 64, 32, 1 },
                0.01 // learning rate
        );

        // Treinar por 100 epocas (mesmo que epochs: 100 do TF.js)
        model.train(trainData.inputs, trainData.labels, 100);
    }

    // ====================================================================
    // 2. Codificacao de features (encodeProduct / encodeUser)
    //    Equivalente a encodeProduct() e encodeUser() do JS
    // ====================================================================

    /**
     * Codifica um produto como vetor numerico:
     * [price_norm * 0.2, avgAge_norm * 0.1, cat_one_hot * 0.4..., color_one_hot * 0.3...]
     *
     * ====================================================================
     * Exemplo de como um produto fica ANTES da codificacao:
     *   { name: "Fones de Ouvido", category: "eletronicos", price: 129.99, color: "preto" }
     *
     * APOS a codificacao (suponha 4 categorias e 7 cores):
     *   [
     *     0.5600,           <- preco normalizado * peso 0.2
     *     0.4200,           <- idade media normalizada * peso 0.1
     *     0, 0.4, 0, 0,     <- one-hot de categoria (eletronicos=ativo) * peso 0.4
     *     0.3, 0, 0, 0, ... <- one-hot de cor (preto=ativo) * peso 0.3
     *   ]
     * ====================================================================
     */
    public double[] encodeProduct(Product product) {
        double[] vector = new double[context.dimensions];
        int idx = 0;

        // Feature 1: preco normalizado com peso
        vector[idx++] = ModelContext.normalize(product.getPrice(), context.minPrice, context.maxPrice)
                        * ModelContext.WEIGHT_PRICE;

        // Feature 2: idade media de compradores normalizada com peso
        double avgAgeNorm = context.productAvgAgeNorm.getOrDefault(product.getName(), 0.5);
        vector[idx++] = avgAgeNorm * ModelContext.WEIGHT_AGE;

        // Features 3..N: one-hot de categoria com peso
        int catIndex = context.categoriesIndex.getOrDefault(product.getCategory(), -1);
        for (int i = 0; i < context.numCategories; i++) {
            vector[idx++] = (i == catIndex ? 1.0 : 0.0) * ModelContext.WEIGHT_CATEGORY;
        }

        // Features N+1..M: one-hot de cor com peso
        int colorIndex = context.colorsIndex.getOrDefault(product.getColor(), -1);
        for (int i = 0; i < context.numColors; i++) {
            vector[idx++] = (i == colorIndex ? 1.0 : 0.0) * ModelContext.WEIGHT_COLOR;
        }

        return vector;
    }

    /**
     * Codifica um usuario como vetor numerico.
     * Se o usuario tem compras: media dos vetores dos produtos comprados.
     * Se nao tem compras: apenas a idade e encodada (os demais campos ficam 0).
     *
     * ====================================================================
     * Exemplo para Rafael (idade 27, compras: Bone + Mochila - acessorios):
     *
     * Vetor resultante (simplificado):
     *   [
     *     0.45,         <- preco medio normalizado * 0.2
     *     0.60,         <- idade normalizada * 0.1
     *     0.4, 0, 0, 0, <- one-hot categoria (acessorios) * 0.4
     *     0.3, 0, 0, 0, <- one-hot cor media * 0.3
     *     ...
     *   ]
     * ====================================================================
     */
    public double[] encodeUser(User user) {
        if (user.getPurchases().isEmpty()) {
            // Sem historico: apenas a idade contribui
            double[] vector = new double[context.dimensions];
            vector[1] = ModelContext.normalize(user.getAge(), context.minAge, context.maxAge)
                        * ModelContext.WEIGHT_AGE;
            return vector;
        }

        // Media dos vetores dos produtos comprados
        double[] sum = new double[context.dimensions];
        for (Product purchase : user.getPurchases()) {
            double[] prodVector = encodeProduct(purchase);
            for (int i = 0; i < context.dimensions; i++) {
                sum[i] += prodVector[i];
            }
        }

        int n = user.getPurchases().size();
        for (int i = 0; i < context.dimensions; i++) {
            sum[i] /= n;
        }

        return sum;
    }

    // ====================================================================
    // 3. Criar dados de treino
    //    Equivalente a createTrainingData() do JS
    // ====================================================================

    private TrainingData createTrainingData() {
        List<double[]> inputs = new ArrayList<>();
        List<Double> labels = new ArrayList<>();

        Set<String> productNames = new HashSet<>();
        context.products.forEach(p -> productNames.add(p.getName()));

        for (User user : context.users) {
            if (user.getPurchases().isEmpty()) continue;

            double[] userVector = encodeUser(user);

            // Criar pares (usuario, produto) e rotular se o usuario comprou ou nao
            for (int pi = 0; pi < context.products.size(); pi++) {
                Product product = context.products.get(pi);
                double[] productVector = context.productVectors[pi];

                // Label = 1 se o usuario comprou este produto, 0 caso contrario
                boolean purchased = user.getPurchases().stream()
                        .anyMatch(p -> p.getName().equals(product.getName()));

                // Combinar vetores do usuario e produto em uma unica entrada
                double[] combined = new double[context.dimensions * 2];
                System.arraycopy(userVector,    0, combined, 0,                    context.dimensions);
                System.arraycopy(productVector, 0, combined, context.dimensions,   context.dimensions);

                inputs.add(combined);
                labels.add(purchased ? 1.0 : 0.0);
            }
        }

        double[][] inputMatrix = inputs.toArray(new double[0][]);
        double[] labelArray = labels.stream().mapToDouble(Double::doubleValue).toArray();

        System.out.printf("Dados de treino criados: %d exemplos%n", inputMatrix.length);
        return new TrainingData(inputMatrix, labelArray);
    }

    // ====================================================================
    // 4. Gerar recomendacoes para um usuario
    //    Equivalente a recommend() do JS
    // ====================================================================

    /**
     * Recomenda produtos para um usuario dado.
     * Retorna a lista de produtos ordenada do mais para o menos relevante.
     *
     * Fluxo:
     *  1. Codifica o usuario em um vetor numerico
     *  2. Para cada produto, concatena [userVector, productVector]
     *  3. Passa pelo modelo treinado -> score entre 0 e 1
     *  4. Ordena pelos scores
     */
    public List<Recommendation> recommend(User user) {
        if (model == null) {
            throw new IllegalStateException("Modelo nao treinado. Chame trainModel() primeiro.");
        }

        double[] userVector = encodeUser(user);
        List<Recommendation> recommendations = new ArrayList<>();

        for (int pi = 0; pi < context.products.size(); pi++) {
            Product product = context.products.get(pi);
            double[] productVector = context.productVectors[pi];

            // Concatenar user + product (mesmo processo do createTrainingData)
            double[] combined = new double[context.dimensions * 2];
            System.arraycopy(userVector,    0, combined, 0,                  context.dimensions);
            System.arraycopy(productVector, 0, combined, context.dimensions, context.dimensions);

            // Score do modelo: probabilidade de o usuario querer este produto
            double score = model.forward(combined);
            recommendations.add(new Recommendation(product, score));
        }

        // Ordenar do maior score para o menor
        recommendations.sort((a, b) -> Double.compare(b.score(), a.score()));
        return recommendations;
    }

    // ====================================================================
    // Classes auxiliares
    // ====================================================================

    private record TrainingData(double[][] inputs, double[] labels) {}

    public record Recommendation(Product product, double score) {
        @Override
        public String toString() {
            return String.format("  [%.4f] %s (R$ %.2f | %s | %s)",
                    score, product.getName(), product.getPrice(),
                    product.getCategory(), product.getColor());
        }
    }
}
