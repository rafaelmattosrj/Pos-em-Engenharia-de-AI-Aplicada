package com.trialforge.tiering;

/** Matemática de vetores usada pelo RAG e pela confiança de resposta — lógica pura, sem rede. */
public final class VectorMath {

    private VectorMath() {
    }

    public static double similaridadeCosseno(double[] a, double[] b) {
        double produto = 0, normaA = 0, normaB = 0;
        for (int i = 0; i < a.length; i++) {
            produto += a[i] * b[i];
            normaA += a[i] * a[i];
            normaB += b[i] * b[i];
        }
        return produto / (Math.sqrt(normaA) * Math.sqrt(normaB));
    }
}
