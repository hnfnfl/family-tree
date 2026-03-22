package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Family struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FamilyTree struct {
	Family        *Family        `json:"family"`
	Members       []*Person      `json:"members"`
	Relationships []Relationship `json:"relationships"`
}

type Relationship struct {
	FromPersonID   string `json:"from_person_id"`
	FromPersonName string `json:"from_person_name"`
	ToPersonID     string `json:"to_person_id"`
	ToPersonName   string `json:"to_person_name"`
	Type           string `json:"type"`
}

type FamilyRepository struct {
	driver   neo4j.DriverWithContext
	database string
}

func NewFamilyRepository(driver neo4j.DriverWithContext) *FamilyRepository {
	return &FamilyRepository{driver: driver, database: "neo4j"}
}

func (r *FamilyRepository) Create(ctx context.Context, family *Family, createdBy string) (*Family, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	family.ID = uuid.New().String()
	family.CreatedBy = createdBy
	family.CreatedAt = time.Now()
	family.UpdatedAt = time.Now()

	query := `
		CREATE (f:Family {
			id: $id,
			name: $name,
			description: $description,
			createdBy: $createdBy,
			createdAt: datetime($createdAt),
			updatedAt: datetime($updatedAt)
		})
		RETURN f
	`

	params := map[string]interface{}{
		"id":          family.ID,
		"name":        family.Name,
		"description": nilString(family.Description),
		"createdBy":   createdBy,
		"createdAt":   family.CreatedAt.Format(time.RFC3339),
		"updatedAt":   family.UpdatedAt.Format(time.RFC3339),
	}

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, err
		}

		node := record.AsMap()["f"].(neo4j.Node)
		return nodeToFamily(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Family), nil
}

func (r *FamilyRepository) GetByID(ctx context.Context, id string) (*Family, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	query := `
		MATCH (f:Family {id: $id})
		RETURN f
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("family not found")
		}

		node := record.AsMap()["f"].(neo4j.Node)
		return nodeToFamily(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Family), nil
}

func (r *FamilyRepository) GetAll(ctx context.Context, limit, offset int) ([]*Family, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	query := `
		MATCH (f:Family)
		RETURN f
		ORDER BY f.name
		SKIP $offset LIMIT $limit
	`

	results, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		})
		if err != nil {
			return nil, err
		}

		families := make([]*Family, 0)
		for cursor.Next(ctx) {
			record := cursor.Record()
			node := record.AsMap()["f"].(neo4j.Node)
			family, err := nodeToFamily(node)
			if err != nil {
				return nil, err
			}
			families = append(families, family)
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}

		return families, nil
	})

	if err != nil {
		return nil, err
	}

	return results.([]*Family), nil
}

func (r *FamilyRepository) Update(ctx context.Context, id string, updates *Family) (*Family, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	query := `
		MATCH (f:Family {id: $id})
		SET f.name = COALESCE($name, f.name),
			f.description = COALESCE($description, f.description),
			f.updatedAt = datetime()
		RETURN f
	`

	params := map[string]interface{}{
		"id":          id,
		"name":        updates.Name,
		"description": nilString(updates.Description),
	}

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("family not found")
		}

		node := record.AsMap()["f"].(neo4j.Node)
		return nodeToFamily(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Family), nil
}

func (r *FamilyRepository) GetFamilyTree(ctx context.Context, familyID string) (*FamilyTree, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	// Single query: get family + all its members + their relationships
	query := `
		MATCH (f:Family {id: $familyId})
		OPTIONAL MATCH (p:Person)-[:BELONGS_TO]->(f)
		WHERE p.isDeleted = false
		OPTIONAL MATCH (p)-[r]->(related:Person)
		WHERE related.isDeleted = false
			AND type(r) <> 'BELONGS_TO'
		RETURN f, p, r, related
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"familyId": familyID})
		if err != nil {
			return nil, err
		}

		var family *Family
		members := make([]*Person, 0)
		relationships := make([]Relationship, 0)
		seenMembers := make(map[string]bool)
		seenRels := make(map[string]bool)

		for cursor.Next(ctx) {
			record := cursor.Record()
			m := record.AsMap()

			// Parse family (only once)
			if family == nil {
				if fNode, ok := m["f"].(neo4j.Node); ok {
					family, err = nodeToFamily(fNode)
					if err != nil {
						return nil, err
					}
				}
			}

			// Parse person (may be nil if family has no members)
			pVal := m["p"]
			if pVal == nil {
				continue
			}
			pNode, ok := pVal.(neo4j.Node)
			if !ok {
				continue
			}
			person, err := nodeToPerson(pNode)
			if err != nil {
				continue
			}
			if !seenMembers[person.ID] {
				members = append(members, person)
				seenMembers[person.ID] = true
			}

			// Parse relationship (optional)
			rVal := m["r"]
			relatedVal := m["related"]
			if rVal == nil || relatedVal == nil {
				continue
			}
			rel, ok := rVal.(neo4j.Relationship)
			if !ok {
				continue
			}
			relatedNode, ok := relatedVal.(neo4j.Node)
			if !ok {
				continue
			}
			relatedPerson, err := nodeToPerson(relatedNode)
			if err != nil {
				continue
			}

			// Deduplicate relationships
			relKey := person.ID + ":" + rel.Type + ":" + relatedPerson.ID
			if !seenRels[relKey] {
				seenRels[relKey] = true
				relationships = append(relationships, Relationship{
					FromPersonID:   person.ID,
					FromPersonName: person.Name,
					ToPersonID:     relatedPerson.ID,
					ToPersonName:   relatedPerson.Name,
					Type:           rel.Type,
				})
			}
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}

		if family == nil {
			return nil, errors.New("family not found")
		}

		return &FamilyTree{
			Family:        family,
			Members:       members,
			Relationships: relationships,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*FamilyTree), nil
}

func (r *FamilyRepository) AddMember(ctx context.Context, familyID, personID string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	query := `
		MATCH (f:Family {id: $familyId})
		MATCH (p:Person {id: $personId})
		WHERE p.isDeleted = false
		MERGE (p)-[:BELONGS_TO]->(f)
		RETURN f.id
	`

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"familyId": familyID,
			"personId": personID,
		})
		if err != nil {
			return nil, err
		}
		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("family or person not found")
		}
		return record, nil
	})

	if err != nil {
		return err
	}
	_ = result
	return nil
}

func nodeToFamily(node neo4j.Node) (*Family, error) {
	props := node.Props

	getString := func(key string) string {
		if v, ok := props[key].(string); ok {
			return v
		}
		return ""
	}

	getOptionalString := func(key string) *string {
		if v, ok := props[key].(string); ok && v != "" {
			return &v
		}
		return nil
	}

	return &Family{
		ID:          getString("id"),
		Name:        getString("name"),
		Description: getOptionalString("description"),
		CreatedBy:   getString("createdBy"),
	}, nil
}
