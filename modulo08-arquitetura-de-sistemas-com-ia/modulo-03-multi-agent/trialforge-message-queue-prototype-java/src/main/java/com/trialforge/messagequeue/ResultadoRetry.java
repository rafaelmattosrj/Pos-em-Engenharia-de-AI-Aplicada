package com.trialforge.messagequeue;

/** Resultado de executarICFComRetry — sucesso com o valor, ou falha apos MAX_TENTATIVAS. */
record ResultadoRetry(boolean sucesso, ResultadoICF resultado, String erro, int tentativasFeitas) {
}
