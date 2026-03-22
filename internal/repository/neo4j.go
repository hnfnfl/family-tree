package repository

import (
	"context"
	"fmt"

	"github.com/hnfnfl/keluarga-tree/internal/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jRepository struct {
	Driver neo4j.DriverWithContext
}

func NewNeo4jDriver(cfg config.Neo4jConfig) (neo4j.DriverWithContext, error) {
	auth := neo4j.BasicAuth(cfg.Username, cfg.Password, "")
	
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		auth,
		func(config *neo4j.Config) {
			config.DatabaseName = cfg.Database
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	return driver, nil
}

func (r *Neo4jRepository) ExecuteQuery(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		records, err := cursor.All(ctx)
		if err != nil {
			return nil, err
		}

		data := make([]map[string]interface{}, len(records))
		for i, record := range records {
			data[i] = record.AsMap()
		}

		return data, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}
