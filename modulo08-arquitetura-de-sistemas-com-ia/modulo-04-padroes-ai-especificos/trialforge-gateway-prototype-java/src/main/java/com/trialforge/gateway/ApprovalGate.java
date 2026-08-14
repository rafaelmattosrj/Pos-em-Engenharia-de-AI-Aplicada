package com.trialforge.gateway;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

/**
 * Approval Gate (Modulo 4.4): pede aprovacao humana antes de um rascunho
 * virar oficial. Os originais em JS/Python precisam de um cuidado especial
 * com stdin nao-interativo (readline recriado a cada chamada, ou EOF
 * prematuro — ver comentario no .js) porque o runtime deles fecha a fonte de
 * entrada sozinho quando ela e lida aos poucos. Em Java, um unico
 * BufferedReader criado uma vez sobre o InputStream (interativo ou
 * tubo/pipe) e reutilizado a cada chamada nao sofre desse problema — cada
 * readLine() retoma de onde a anterior parou, entao basta NAO recriar o
 * reader a cada chamada. Mesma disciplina descrita no .js, so que sem
 * precisar detectar TTY (mesmo motivo pelo qual a versao Python, que usa
 * input() simples, tambem nao precisou do workaround).
 */
public class ApprovalGate implements Approver {

    private final BufferedReader reader;
    private final PrintStream saida;

    public ApprovalGate(InputStream entrada, PrintStream saida) {
        this.reader = new BufferedReader(new InputStreamReader(entrada, StandardCharsets.UTF_8));
        this.saida = saida;
    }

    /**
     * Mostra o rascunho e le uma linha de resposta — "s" (case-insensitive,
     * com espacos em volta ignorados) aprova, qualquer outra coisa (inclusive
     * EOF/linha vazia) rejeita.
     */
    @Override
    public boolean pedirAprovacaoHumana(String rascunho) throws IOException {
        saida.println("\n[Approval Gate] Rascunho aguardando aprovação antes de virar oficial:");
        saida.println("    " + rascunho);
        saida.print("\nAprovar? (s/n) ");
        saida.flush();

        String linha = reader.readLine();
        if (linha == null) {
            return false; // sem mais entrada (EOF) — trata como reprovado, nao como erro fatal
        }
        return linha.trim().equalsIgnoreCase("s");
    }
}
