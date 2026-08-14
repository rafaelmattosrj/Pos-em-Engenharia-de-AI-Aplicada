package com.trialforge.reactagent;

import java.util.List;

/**
 * Mesma disciplina do Módulo 1.3 (decision-framework-tool): o que é determinístico
 * no nosso próprio código ganha teste automatizado; a redação exata do modelo, e se
 * ele decide chamar a ferramenta ou não, fica pra observação ao vivo na simulação,
 * não pra assert aqui. Estes casos cobrem os dois bugs reais encontrados testando
 * contra o modelo de verdade na sessão original: a grafia "jurisdicicao" (parâmetro
 * mal formado) e o tema formulado com "adolescente"/"pediátrico" em vez de "menor".
 *
 * Porte 1:1 de CASOS_TESTE_FERRAMENTA em react-agent-prototype.js / .py. Compartilhada
 * entre {@link Main} (roda em console, como no original) e os testes JUnit.
 */
public final class CasosTesteFerramenta {

    private CasosTesteFerramenta() {
    }

    public record Caso(String tema, String jurisdicao, boolean esperaAchar) {
    }

    public static final List<Caso> CASOS = List.of(
            new Caso("Assentimento para menores de idade em estudos clínicos", "ANVISA", true),
            new Caso("Consentimento de adolescentes em pesquisa", "ANVISA", true),
            new Caso("Cuidados pediátricos em ensaio clínico", "ANVISA", true),
            new Caso("Consentimento informado de população adulta", "ANVISA", false),
            new Caso("Assentimento para menores de idade", "FDA", false), // jurisdição errada
            new Caso("Termo de Consentimento Livre e Esclarecido (TCLE)", "ANVISA", false)
    );
}
