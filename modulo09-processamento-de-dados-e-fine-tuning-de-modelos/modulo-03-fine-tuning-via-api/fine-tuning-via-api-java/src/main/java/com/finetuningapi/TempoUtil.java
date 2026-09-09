package com.finetuningapi;

import java.time.Duration;
import java.time.Instant;
import java.util.Locale;

/** Formatacao de duracao real de job (Modulo 3.2/3.5), porte de formatarDuracao. */
public final class TempoUtil {

    private TempoUtil() {
    }

    /** Formato do Modulo 3.2: "45min41.95s". */
    public static String formatarDuracaoCompacta(String inicioISO, String fimISO) {
        long totalMs = millisEntre(inicioISO, fimISO);
        long minutos = totalMs / 60000;
        double segundos = (totalMs % 60000) / 1000.0;
        return String.format(Locale.US, "%dmin%.2fs", minutos, segundos);
    }

    /** Formato do Modulo 3.5: "45min 42s" (com espaco, segundos arredondados). */
    public static String formatarDuracaoComEspaco(double duracaoSegundos) {
        long minutos = (long) Math.floor(duracaoSegundos / 60);
        long segundos = Math.round(duracaoSegundos - minutos * 60);
        return minutos + "min " + segundos + "s";
    }

    public static double duracaoSegundos(String criadoEmISO, String concluidoEmISO) {
        return millisEntre(criadoEmISO, concluidoEmISO) / 1000.0;
    }

    private static long millisEntre(String inicioISO, String fimISO) {
        Instant inicio = Instant.parse(inicioISO);
        Instant fim = Instant.parse(fimISO);
        return Duration.between(inicio, fim).toMillis();
    }
}
