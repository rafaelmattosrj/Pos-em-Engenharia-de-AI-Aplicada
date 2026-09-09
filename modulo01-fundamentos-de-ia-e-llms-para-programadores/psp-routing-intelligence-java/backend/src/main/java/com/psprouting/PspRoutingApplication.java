package com.psprouting;

import io.github.cdimascio.dotenv.Dotenv;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * PSP Routing Intelligence — IA para recomendacao de Payment Service
 * Providers, implementada a partir da especificacao em
 * {@code ../IDEIA.md} (Embeddings + RAG + LLM sobre transacoes historicas).
 *
 * Carrega o {@code .env} para variaveis de ambiente do processo antes de
 * subir o contexto Spring — mesmo padrao de
 * {@code dotenv-java} usado nos demais projetos Java do modulo — para que
 * {@code application.properties} possa referenciar
 * {@code ${OPENROUTER_API_KEY}} etc. mesmo quando essas variaveis nao estao
 * exportadas no ambiente do sistema operacional.
 */
@SpringBootApplication
public class PspRoutingApplication {

    public static void main(String[] args) {
        Dotenv dotenv = Dotenv.configure().ignoreIfMissing().load();
        dotenv.entries().forEach(entry -> {
            if (System.getProperty(entry.getKey()) == null) {
                System.setProperty(entry.getKey(), entry.getValue());
            }
        });

        SpringApplication.run(PspRoutingApplication.class, args);
    }
}
