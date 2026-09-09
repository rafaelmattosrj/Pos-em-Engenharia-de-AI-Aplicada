// Package neo4jservice encapsula o acesso ao Neo4j — equivalente a
// Neo4jService.java (execução de Cypher, schema e seed de dados).
package neo4jservice

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Service executa queries Cypher e gerencia o schema/seed do banco.
type Service struct {
	driver neo4j.DriverWithContext
}

// NewService conecta ao Neo4j em uri com as credenciais informadas.
func NewService(ctx context.Context, uri, user, password string) (*Service, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("conectando ao Neo4j em %s: %w", uri, err)
	}
	return &Service{driver: driver}, nil
}

// Close libera a conexão com o Neo4j.
func (s *Service) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// Query executa uma query Cypher e retorna cada registro como um mapa
// coluna->valor — equivalente a Neo4jService.query.
func (s *Service) Query(ctx context.Context, cypherQuery string) ([]map[string]any, error) {
	result, err := neo4j.ExecuteQuery(ctx, s.driver, cypherQuery, nil, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}
	records := make([]map[string]any, 0, len(result.Records))
	for _, rec := range result.Records {
		records = append(records, rec.AsMap())
	}
	return records, nil
}

// ValidateQuery confirma que a query é sintaticamente válida via EXPLAIN,
// sem executá-la de fato — equivalente a Neo4jService.validateQuery.
func (s *Service) ValidateQuery(ctx context.Context, cypherQuery string) bool {
	_, err := neo4j.ExecuteQuery(ctx, s.driver, "EXPLAIN "+cypherQuery, nil, neo4j.EagerResultTransformer)
	return err == nil
}

// GetSchema tenta obter o schema via APOC; se indisponível, cai para um
// resumo com labels e tipos de relação — equivalente a
// Neo4jService.getSchema/describeSchema.
func (s *Service) GetSchema(ctx context.Context) string {
	result, err := neo4j.ExecuteQuery(ctx, s.driver, "CALL apoc.meta.schema()", nil, neo4j.EagerResultTransformer)
	if err != nil {
		return s.describeSchema(ctx)
	}
	return fmt.Sprintf("%v", result.Records)
}

func (s *Service) describeSchema(ctx context.Context) string {
	labels, err := neo4j.ExecuteQuery(ctx, s.driver, "CALL db.labels()", nil, neo4j.EagerResultTransformer)
	if err != nil {
		return "Schema unavailable"
	}
	relTypes, err := neo4j.ExecuteQuery(ctx, s.driver, "CALL db.relationshipTypes()", nil, neo4j.EagerResultTransformer)
	if err != nil {
		return "Schema unavailable"
	}
	return fmt.Sprintf("Labels: %v\nRelationships: %v", labels.Records, relTypes.Records)
}

// SeedData recria os cursos/alunos/matrículas de exemplo — equivalente a
// Neo4jService.seedData.
func (s *Service) SeedData(ctx context.Context) error {
	statements := []string{
		"MATCH (n) DETACH DELETE n",
		`CREATE (c1:Course {id: 1, name: 'Java com Spring Boot', price: 299.0, category: 'Backend'})
		 CREATE (c2:Course {id: 2, name: 'React Avançado', price: 249.0, category: 'Frontend'})
		 CREATE (c3:Course {id: 3, name: 'Machine Learning com Python', price: 399.0, category: 'IA'})
		 CREATE (c4:Course {id: 4, name: 'Docker e Kubernetes', price: 199.0, category: 'DevOps'})
		 CREATE (c5:Course {id: 5, name: 'TypeScript para Devs', price: 149.0, category: 'Frontend'})`,
		`MATCH (c1:Course {id: 1}), (c2:Course {id: 2}), (c3:Course {id: 3})
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
		 CREATE (s1)-[:ENROLLED_IN {enrolledAt: '2024-04-01'}]->(c2)`,
	}
	for _, stmt := range statements {
		if _, err := neo4j.ExecuteQuery(ctx, s.driver, stmt, nil, neo4j.EagerResultTransformer); err != nil {
			return err
		}
	}
	return nil
}
