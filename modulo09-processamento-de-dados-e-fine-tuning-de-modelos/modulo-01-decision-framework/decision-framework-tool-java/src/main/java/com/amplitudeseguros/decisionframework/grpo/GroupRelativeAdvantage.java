package com.amplitudeseguros.decisionframework.grpo;

/**
 * A_i = (r_i - média(r)) / desvio_padrão(r) -- o coração do GRPO: o próprio
 * grupo vira a linha de base, sem precisar de um modelo de valor (critic)
 * separado como no PPO clássico. Equivalente a vantagemRelativaAoGrupo().
 */
public final class GroupRelativeAdvantage {

    /** Guarda contra divisão por zero quando o grupo empata (desvio padrão = 0). */
    private static final double EPS = 1e-4;

    private GroupRelativeAdvantage() {
    }

    public record Resultado(double[] vantagens, double media, double desvio) {
    }

    public static Resultado calcular(double[] recompensas) {
        double media = 0;
        for (double r : recompensas) {
            media += r;
        }
        media /= recompensas.length;

        double variancia = 0;
        for (double r : recompensas) {
            variancia += Math.pow(r - media, 2);
        }
        variancia /= recompensas.length;
        double desvio = Math.sqrt(variancia);
        double desvioEfetivo = desvio > 0 ? desvio : EPS;

        double[] vantagens = new double[recompensas.length];
        for (int i = 0; i < recompensas.length; i++) {
            vantagens[i] = (recompensas[i] - media) / desvioEfetivo;
        }

        return new Resultado(vantagens, media, desvio);
    }
}
