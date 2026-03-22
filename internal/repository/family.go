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
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FamilyTree struct {
	Family      *Family          `json:"family"`
	Members     []*Person        `json:"members"`
	Relationships []Relationship `json:"relationships"`
}

type Relationship struct {
	FromPersonID   string    `json:"from_person_id"`
	FromPersonName string    `json:"from_person_name"`
	ToPersonID     string    `json:"to_person_id"`
	ToPersonName   string    `json:"to_person_name"`
	Type           string    `json:"type"`
	CreatedAt      time.Time `json:"created_at"`
}

type FamilyRepository struct {
	driver neo4j.DriverWithContext
}

func NewFamilyRepository(driver neo4j.DriverWithContext) *FamilyRepository {
	return &FamilyRepository{driver: driver}
}

func (r *FamilyRepository) Create(ctx context.Context, family *Family, createdBy string) (*Family, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
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
		"description": family.Description,
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
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
		"description": updates.Description,
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Get family
	family, err := r.GetByID(ctx, familyID)
	if err != nil {
		return nil, err
	}

	// Get all members and relationships in one query
	query := `
		MATCH (f:Family {id: $familyId})<-[:BELONGS_TO]-(p:Person)
		WHERE p.isDeleted = false
		OPTIONAL MATCH (p)-[r]-(related:Person)
		WHERE related.isDeleted = false
		RETURN p, r, related
	`

	results, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"familyId": familyID})
		if err != nil {
			return nil, err
		}

		members := make([]*Person, 0)
		relationships := make([]Relationship, 0)
		seenMembers := make(map[string]bool)

		for cursor.Next(ctx) {
			record := cursor.Record()
			
			// Add person
			pNode := record.AsMap()["p"].(neo4j.Node)
			person, err := nodeToPerson(pNode)
			if err != nil {
				continue
			}

			if !seenMembers[person.ID] {
				members = append(members, person)
				seenMembers[person.ID] = true
			}

			// Add relationship if exists
			rRel := record.AsMap()["r"]
			if rRel != nil {
				if rel, ok := rRel.(neo4j.Relationship); ok {
					// Get related person
					relatedNode := record.AsMap()["related"].(neo4j.Node)
					relatedPerson, err := nodeToPerson(relatedNode)
					if err != nil {
						continue
					}

					// Get relationship type
					relType := rel.Type
					if len(relType) > 0 && relType[0] == ':' {
						relType = relType[1:]
					}

					relationships = append(relationships, Relationship{
						FromPersonID:   person.ID,
						FromPersonName: person.Name,
						ToPersonID:     relatedPerson.ID,
						ToPersonName:   relatedPerson.Name,
						Type:           relType,
						CreatedAt:      time.Now(),
					})
				}
			}
		}

		return &FamilyTree{
			Family:      family,
			Members:     members,
			Relationships: relationships,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return results.(*FamilyTree), nil
}

func (r *FamilyRepository) AddMember(ctx context.Context, familyID, personID string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (f:Family {id: $familyId})
		MATCH (p:Person {id: $personId})
		WHERE p.isDeleted = false
		MERGE (p)-[:BELONGS_TO]->(f)
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"familyId": familyID,
			"personId": personID,
		})
		if err != nil {
			return nil, err
		}
		return cursor.Consume(ctx)
	})

	return err
}

func nodeToFamily(node neo4j.Node) (*Family, error) {
	props := node.Props
	family := &Family{}

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

	family.ID = getString("id")
	family.Name = getString("name")
	family.Description = getOptionalString("description")
	family.CreatedBy = getString("createdBy")

	return family, nil
}
