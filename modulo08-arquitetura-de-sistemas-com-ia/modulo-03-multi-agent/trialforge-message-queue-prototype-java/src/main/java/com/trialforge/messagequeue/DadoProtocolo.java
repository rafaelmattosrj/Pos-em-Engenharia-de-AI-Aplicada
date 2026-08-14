package com.trialforge.messagequeue;

import java.util.Map;

/**
 * Evento rico publicado em "protocolo:pronto" — carrega a versao e os
 * criterios estruturados, nao so uma notificacao vazia (paragrafos 62-64 do
 * TP). E uma COPIA do estado no momento da publicacao: revisoes futuras no
 * {@link EstadoProtocolo} nao mudam retroativamente o que ja foi publicado
 * neste evento.
 */
public record DadoProtocolo(int versao, Map<String, Integer> criterios, String estudo) {
}
