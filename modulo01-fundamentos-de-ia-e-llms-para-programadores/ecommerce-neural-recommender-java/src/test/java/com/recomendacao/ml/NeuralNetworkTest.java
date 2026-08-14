package com.recomendacao.ml;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class NeuralNetworkTest {

    @Test
    void forwardReturnsValueBetweenZeroAndOne() {
        NeuralNetwork net = new NeuralNetwork(new int[]{4, 8, 1}, 0.01);

        double output = net.forward(new double[]{0.1, 0.2, 0.3, 0.4});

        assertThat(output).isBetween(0.0, 1.0);
    }

    @Test
    void sameSeedProducesDeterministicInitialWeights() {
        // O construtor usa seed fixo (42), entao duas redes com a mesma
        // arquitetura devem produzir a mesma saida para a mesma entrada
        // antes de qualquer treino.
        NeuralNetwork netA = new NeuralNetwork(new int[]{3, 5, 1}, 0.01);
        NeuralNetwork netB = new NeuralNetwork(new int[]{3, 5, 1}, 0.01);

        double[] input = {0.5, -0.2, 0.9};

        assertThat(netA.forward(input)).isEqualTo(netB.forward(input));
    }

    @Test
    void trainReducesLossOnLearnablePattern() {
        // Padrao simples e linearmente separavel: label = 1 quando soma das
        // entradas > 1, senao 0. A rede deve aprender a distinguir os dois casos.
        NeuralNetwork net = new NeuralNetwork(new int[]{2, 8, 1}, 0.1);

        double[][] inputs = {
                {0.9, 0.9}, {1.0, 0.8}, {0.95, 0.95}, {0.85, 0.9},
                {0.0, 0.0}, {0.1, 0.05}, {0.05, 0.1}, {0.0, 0.05}
        };
        double[] labels = {1, 1, 1, 1, 0, 0, 0, 0};

        double lossBefore = averageLoss(net, inputs, labels);

        net.train(inputs, labels, 200);

        double lossAfter = averageLoss(net, inputs, labels);

        assertThat(lossAfter).isLessThan(lossBefore);
    }

    @Test
    void forwardThenBackwardUpdatesWeightsTowardsLabel() {
        NeuralNetwork net = new NeuralNetwork(new int[]{2, 4, 1}, 0.5);

        double[] input = {1.0, 1.0};
        double firstPrediction = net.forward(input);
        net.backward(1.0);

        double secondPrediction = net.forward(input);
        net.backward(1.0);

        double thirdPrediction = net.forward(input);

        // Apos varias atualizacoes em direcao ao label 1.0, a predicao deve
        // se aproximar mais de 1.0 do que a predicao inicial.
        assertThat(Math.abs(1.0 - thirdPrediction)).isLessThan(Math.abs(1.0 - firstPrediction));
        assertThat(secondPrediction).isNotEqualTo(firstPrediction);
    }

    private double averageLoss(NeuralNetwork net, double[][] inputs, double[] labels) {
        double total = 0;
        for (int i = 0; i < inputs.length; i++) {
            double pred = net.forward(inputs[i]);
            double label = labels[i];
            total += -(label * Math.log(pred + 1e-7) + (1 - label) * Math.log(1 - pred + 1e-7));
        }
        return total / inputs.length;
    }
}
