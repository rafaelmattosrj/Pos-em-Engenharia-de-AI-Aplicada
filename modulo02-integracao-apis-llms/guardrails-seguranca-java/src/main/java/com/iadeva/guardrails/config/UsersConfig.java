package com.iadeva.guardrails.config;

import com.iadeva.guardrails.model.User;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.List;
import java.util.Map;

// Configuração de usuários com RBAC — equivalente ao users.json + config.ts do TypeScript
@Configuration
public class UsersConfig {

    @Bean
    public Map<String, User> users() {
        return Map.of(
                "admin", new User("admin", "admin",
                        List.of("read_files", "write_files", "delete_files"),
                        "Admin User"),
                "member", new User("member", "member",
                        List.of(),
                        "Regular Member")
        );
    }
}
