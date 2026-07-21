package com.iadeva.neo4j.service;

import org.neo4j.driver.Driver;
import org.neo4j.driver.Result;
import org.neo4j.driver.Session;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

// Serviço Neo4j — equivalente ao neo4jService.ts
@Service
public class Neo4jService {

    private final Driver driver;

    public Neo4jService(Driver driver) {
        this.driver = driver;
    }

    public List<Map<String, Object>> query(String cypherQuery) {
        try (Session session = driver.session()) {
            Result result = session.run(cypherQuery);
            return result.list(record -> record.asMap());
        }
    }

    public boolean validateQuery(String cypherQuery) {
        try (Session session = driver.session()) {
            session.run("EXPLAIN " + cypherQuery);
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    public String getSchema() {
        try (Session session = driver.session()) {
            Result result = session.run("CALL apoc.meta.schema()");
            return result.list().toString();
        } catch (Exception e) {
            // Fallback: describe node labels and relationships
            return describeSchema();
        }
    }

    private String describeSchema() {
        try (Session session = driver.session()) {
            String labels = session.run("CALL db.labels()").list().toString();
            String relTypes = session.run("CALL db.relationshipTypes()").list().toString();
            return "Labels: " + labels + "\nRelationships: " + relTypes;
        } catch (Exception e) {
            return "Schema unavailable";
        }
    }

    public void seedData() {
        try (Session session = driver.session()) {
            // Limpa dados existentes
            session.run("MATCH (n) DETACH DELETE n");

            // Cria cursos
            session.run("""
                    CREATE (c1:Course {id: 1, name: 'Java com Spring Boot', price: 299.0, category: 'Backend'})
                    CREATE (c2:Course {id: 2, name: 'React Avançado', price: 249.0, category: 'Frontend'})
                    CREATE (c3:Course {id: 3, name: 'Machine Learning com Python', price: 399.0, category: 'IA'})
                    CREATE (c4:Course {id: 4, name: 'Docker e Kubernetes', price: 199.0, category: 'DevOps'})
                    CREATE (c5:Course {id: 5, name: 'TypeScript para Devs', price: 149.0, category: 'Frontend'})
                    """);

            // Cria alunos e relacionamentos de compra
            session.run("""
                    MATCH (c1:Course {id: 1}), (c2:Course {id: 2}), (c3:Course {id: 3})
                    MATCH (c4:Course {id: 4}), (c5:Course {id: 5})
                    CREATE (s1:Student {id: 1, name: 'Ana Silva', email: 'ana@email.com'})
                    CREATE (s2:Student {id: 2, name: 'Bruno Costa', email: 'bruno@email.com'})
                    CREATE (s3:Student {id: 3, name: 'Carla Nunes', email: 'carla@email.com'})
                    CREATE (s1)-[:ENROLLED_IN {enrolledAt: '2024-01-15'}]->(c1)
                    CREATE (s1)-[:ENROLLED_IN {enrolledAt: '2024-02-10'}]->(c3)
                    CREATE (s2)-[:ENROLLED_IN {enrolledAt: '2024-01-20'}]->(c1)
                    CREATE (s2)-[:ENROLLED_IN {enrolledAt: '2024-03-05'}]->(c4)
                    CREATE (s3)-[:ENROLLED_IN {enrolledAt: '2024-02-15'}]->(c2)
                    CREATE (s3)-[:ENROLLED_IN {enrolledAt: '2024-02-20'}]->(c5)
                    CREATE (s1)-[:ENROLLED_IN {enrolledAt: '2024-04-01'}]->(c2)
                    """);

            System.out.println("Neo4j seed data created");
        }
    }
}
