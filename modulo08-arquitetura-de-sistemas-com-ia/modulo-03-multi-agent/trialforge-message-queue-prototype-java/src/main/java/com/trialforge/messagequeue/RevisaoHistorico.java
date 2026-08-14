package com.trialforge.messagequeue;

import java.util.Map;

/** Um item do historico de revisoes do Protocolo — nunca sobrescrito, so acrescentado. */
public record RevisaoHistorico(int versaoAnterior, Map<String, Integer> criteriosAnteriores, String motivo) {
}
