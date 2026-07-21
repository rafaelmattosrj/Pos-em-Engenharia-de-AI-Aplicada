package com.recomendacao.ml;

import java.util.Random;

/**
 * Rede Neural Feedforward implementada do zero em Java.
 *
 * Espelho da rede configurada em configureNeuralNetAndTrain() do original JS:
 *   - Camada entrada:   inputDim neurônios
 *   - Camada oculta 1: 128 neuronios, ReLU
 *   - Camada oculta 2:  64 neuronios, ReLU
 *   - Camada oculta 3:  32 neuronios, ReLU
 *   - Camada saida:      1 neuronio,  Sigmoid
 *
 * Algoritmo de treino: SGD (Stochastic Gradient Descent) com backpropagation.
 * Loss: Binary Cross-Entropy
 */
public class NeuralNetwork {

    // ====================================================================
    // Pesos e biases de cada camada
    // weights[l][i][j] = peso da camada l, neurônio i, entrada j
    // biases[l][i]     = bias da camada l, neurônio i
    // ====================================================================
    private final double[][][] weights;
    private final double[][] biases;
    private final String[] activations;
    private final double learningRate;

    // Cache para backpropagation
    private double[][] preActivations;  // valor ANTES da funcao de ativacao
    private double[][] postActivations; // valor DEPOIS da funcao de ativacao (saida)

    /**
     * @param layerSizes    tamanho de cada camada, ex: [20, 128, 64, 32, 1]
     * @param learningRate  taxa de aprendizado (ex: 0.01)
     */
    public NeuralNetwork(int[] layerSizes, double learningRate) {
        this.learningRate = learningRate;
        int numLayers = layerSizes.length - 1;

        weights     = new double[numLayers][][];
        biases      = new double[numLayers][];
        activations = new String[numLayers];
        preActivations  = new double[numLayers][];
        postActivations = new double[numLayers + 1][];

        Random rand = new Random(42); // seed fixo para reproducibilidade

        for (int l = 0; l < numLayers; l++) {
            int in  = layerSizes[l];
            int out = layerSizes[l + 1];

            weights[l] = new double[out][in];
            biases[l]  = new double[out];

            // Inicializacao He: escala = sqrt(2 / entradas)
            // Funciona bem com ReLU, evita vanishing/exploding gradients
            double scale = Math.sqrt(2.0 / in);
            for (int i = 0; i < out; i++) {
                for (int j = 0; j < in; j++) {
                    weights[l][i][j] = rand.nextGaussian() * scale;
                }
            }

            // Ultima camada usa sigmoid, demais usam relu
            activations[l] = (l == numLayers - 1) ? "sigmoid" : "relu";
        }
    }

    // ====================================================================
    // Funcoes de ativacao
    // ====================================================================

    private double relu(double x) {
        return Math.max(0.0, x);
    }

    private double sigmoid(double x) {
        return 1.0 / (1.0 + Math.exp(-x));
    }

    /** Derivada de ReLU em relacao ao pre-ativacao */
    private double reluDeriv(double preAct) {
        return preAct > 0 ? 1.0 : 0.0;
    }

    /** Derivada de Sigmoid em relacao ao pre-ativacao */
    private double sigmoidDeriv(double preAct) {
        double s = sigmoid(preAct);
        return s * (1.0 - s);
    }

    // ====================================================================
    // Forward pass: calcula a predicao dado um vetor de entrada
    // ====================================================================

    /**
     * Propaga a entrada pela rede e retorna a predicao (0-1).
     * Salva os valores intermediarios para o backpropagation.
     */
    public double forward(double[] input) {
        postActivations[0] = input;
        double[] current = input;

        for (int l = 0; l < weights.length; l++) {
            int outSize = weights[l].length;
            double[] pre  = new double[outSize];
            double[] post = new double[outSize];

            for (int i = 0; i < outSize; i++) {
                pre[i] = biases[l][i];
                for (int j = 0; j < current.length; j++) {
                    pre[i] += weights[l][i][j] * current[j];
                }
                post[i] = "relu".equals(activations[l]) ? relu(pre[i]) : sigmoid(pre[i]);
            }

            preActivations[l]      = pre;
            postActivations[l + 1] = post;
            current = post;
        }

        return current[0]; // saida binaria unica
    }

    // ====================================================================
    // Backward pass: calcula gradientes e atualiza os pesos
    // ====================================================================

    /**
     * Backpropagation com SGD.
     * Chame sempre APOS um forward() com o mesmo exemplo.
     *
     * @param label valor real (0 ou 1)
     */
    public void backward(double label) {
        int numLayers = weights.length;
        double[][] deltas = new double[numLayers][];

        // --- Camada de saida ---
        // delta = (predicao - label) * sigmoid'(preAct)
        // derivada da binary cross-entropy + sigmoid juntas simplifica para: (pred - label)
        double pred   = postActivations[numLayers][0];
        double preOut = preActivations[numLayers - 1][0];
        deltas[numLayers - 1] = new double[]{ (pred - label) * sigmoidDeriv(preOut) };

        // --- Camadas ocultas (de tras para frente) ---
        for (int l = numLayers - 2; l >= 0; l--) {
            int size = weights[l].length;
            deltas[l] = new double[size];

            for (int i = 0; i < size; i++) {
                double sum = 0;
                for (int k = 0; k < deltas[l + 1].length; k++) {
                    sum += weights[l + 1][k][i] * deltas[l + 1][k];
                }
                deltas[l][i] = sum * reluDeriv(preActivations[l][i]);
            }
        }

        // --- Atualizar pesos e biases ---
        for (int l = 0; l < numLayers; l++) {
            double[] prevOutput = postActivations[l];
            for (int i = 0; i < weights[l].length; i++) {
                for (int j = 0; j < weights[l][i].length; j++) {
                    weights[l][i][j] -= learningRate * deltas[l][i] * prevOutput[j];
                }
                biases[l][i] -= learningRate * deltas[l][i];
            }
        }
    }

    // ====================================================================
    // Treino completo (equivalente ao model.fit() do TensorFlow.js)
    // ====================================================================

    /**
     * Treina a rede neural.
     *
     * @param inputs  matriz de entrada [numExemplos][numFeatures]
     * @param labels  rotulos binarios [numExemplos]
     * @param epochs  numero de epocas
     */
    public void train(double[][] inputs, double[] labels, int epochs) {
        System.out.println("\n=== Iniciando Treinamento da Rede Neural ===");
        System.out.printf("Exemplos: %d | Epocas: %d | Learning Rate: %.3f%n%n",
                inputs.length, epochs, learningRate);

        Random rand = new Random();
        int n = inputs.length;

        for (int epoch = 0; epoch < epochs; epoch++) {
            double totalLoss = 0;
            int correct = 0;

            // Embaralhar indices a cada epoca (equivalente ao shuffle: true do TF.js)
            int[] indices = createShuffledIndices(n, rand);

            for (int idx : indices) {
                double pred  = forward(inputs[idx]);
                double label = labels[idx];

                backward(label);

                // Binary cross-entropy: -[y*log(p) + (1-y)*log(1-p)]
                double loss = -(label * Math.log(pred + 1e-7)
                             + (1 - label) * Math.log(1 - pred + 1e-7));
                totalLoss += loss;

                if ((pred >= 0.5 && label == 1.0) || (pred < 0.5 && label == 0.0)) {
                    correct++;
                }
            }

            if ((epoch + 1) % 10 == 0) {
                double avgLoss = totalLoss / n;
                double accuracy = (double) correct / n * 100;
                System.out.printf("Epoca %3d | Loss: %.4f | Acuracia: %.1f%%%n",
                        epoch + 1, avgLoss, accuracy);
            }
        }

        System.out.println("\n=== Treinamento Concluido ===\n");
    }

    private int[] createShuffledIndices(int n, Random rand) {
        int[] indices = new int[n];
        for (int i = 0; i < n; i++) indices[i] = i;
        for (int i = n - 1; i > 0; i--) {
            int j = rand.nextInt(i + 1);
            int tmp = indices[i]; indices[i] = indices[j]; indices[j] = tmp;
        }
        return indices;
    }
}
