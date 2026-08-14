package com.trialforge.tiering;

import java.util.concurrent.ConcurrentHashMap;
import java.util.Map;

/**
 * Orçamento por estudo (Módulo 5.1 Uber/Vitalis Platform, Módulo 5.2 LiteLLM).
 *
 * {@link #reservarOrcamento} é uma reserva otimista — checa E debita atomicamente
 * (bloco {@code synchronized} na conta do estudo), fechando a mesma janela de
 * corrida discutida no comentário do original em JS: lá, o eventloop de Node
 * garante que não existe {@code await} entre checar e debitar dentro de uma
 * única chamada; aqui, como Java roda threads de verdade (ver
 * {@code simularVolumeConcorrente} em {@link CascadeGateway}), a mesma garantia
 * exige um lock explícito — sem ele, N threads concorrentes do mesmo estudo
 * poderiam passar todas pela checagem antes de qualquer uma debitar.
 */
public class OrcamentoManager {

    private static final class Conta {
        final double limite;
        double gasto;

        Conta(double limite, double gasto) {
            this.limite = limite;
            this.gasto = gasto;
        }
    }

    private final Map<String, Conta> contas = new ConcurrentHashMap<>();

    public void definirOrcamento(String estudoId, double limite) {
        contas.put(estudoId, new Conta(limite, 0));
    }

    public boolean verificarOrcamento(String estudoId, double custoEstimado) {
        Conta conta = obter(estudoId);
        synchronized (conta) {
            return conta.gasto + custoEstimado <= conta.limite;
        }
    }

    public void registrarGasto(String estudoId, double custo) {
        Conta conta = obter(estudoId);
        synchronized (conta) {
            conta.gasto += custo;
        }
    }

    /** Reserva o pior caso ANTES de qualquer chamada de modelo — checa e debita no mesmo passo. */
    public boolean reservarOrcamento(String estudoId, double custoReservado) {
        Conta conta = obter(estudoId);
        synchronized (conta) {
            if (conta.gasto + custoReservado > conta.limite) {
                return false;
            }
            conta.gasto += custoReservado;
            return true;
        }
    }

    /** Devolve a diferença entre o que foi reservado (pior caso) e o custo real, se sobrar algo. */
    public void liberarSobra(String estudoId, double valorASobrar) {
        if (valorASobrar <= 0) {
            return;
        }
        Conta conta = obter(estudoId);
        synchronized (conta) {
            conta.gasto -= valorASobrar;
        }
    }

    public double gastoAtual(String estudoId) {
        Conta conta = obter(estudoId);
        synchronized (conta) {
            return conta.gasto;
        }
    }

    public double limite(String estudoId) {
        return obter(estudoId).limite;
    }

    private Conta obter(String estudoId) {
        Conta conta = contas.get(estudoId);
        if (conta == null) {
            throw new IllegalArgumentException("Estudo desconhecido: " + estudoId);
        }
        return conta;
    }
}
