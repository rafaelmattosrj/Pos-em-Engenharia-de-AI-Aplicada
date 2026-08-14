package com.trialforge.tiering;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

/**
 * Implementação real de {@link ApprovalPrompt}: lê uma linha de stdin (funciona
 * tanto em terminal interativo quanto com stdin redirecionado, ex.:
 * {@code echo s | mvn exec:java} — mesma correção do "readline recriado a cada
 * chamada" documentada no original em JS, aqui resolvida com um único
 * {@link BufferedReader} reaproveitado por toda a execução).
 */
public class StdinApprovalPrompt implements ApprovalPrompt {

    private final BufferedReader reader = new BufferedReader(new InputStreamReader(System.in, StandardCharsets.UTF_8));

    @Override
    public boolean approve(String rascunho) throws IOException {
        System.out.println("\n[Approval Gate] Rascunho aguardando aprovação antes de virar oficial:");
        System.out.println("    " + rascunho);
        System.out.print("\nAprovar? (s/n) ");
        String resposta = reader.readLine();
        if (resposta == null) {
            resposta = "";
        }
        return resposta.trim().equalsIgnoreCase("s");
    }
}
