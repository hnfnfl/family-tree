package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Person struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Gender             string                 `json:"gender"`
	BirthDate          string                 `json:"birthDate"`
	DeathDate          *string                `json:"deathDate,omitempty"`
	Title              *string                `json:"title,omitempty"`
	Bio                *string                `json:"bio,omitempty"`
	AddressStreet      *string                `json:"addressStreet,omitempty"`
	AddressNeighborhood *string               `json:"addressNeighborhood,omitempty"`
	AddressCity        *string                `json:"addressCity,omitempty"`
	AddressProvince    *string                `json:"addressProvince,omitempty"`
	AddressPostalCode  *string                `json:"addressPostalCode,omitempty"`
	AddressCountry     *string                `json:"addressCountry,omitempty"`
	PhonePrimary       *string                `json:"phonePrimary,omitempty"`
	PhonePrimaryType   *string                `json:"phonePrimaryType,omitempty"`
	PhoneVerified      *bool                  `json:"phoneVerified,omitempty"`
	IsDeleted          bool                   `json:"isDeleted"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

type PersonRepository struct {
	driver neo4j.DriverWithContext
}

func NewPersonRepository(driver neo4j.DriverWithContext) *PersonRepository {
	return &PersonRepository{driver: driver}
}

func (r *PersonRepository) Create(ctx context.Context, person *Person) (*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	person.ID = uuid.New().String()
	person.CreatedAt = time.Now()
	person.UpdatedAt = time.Now()
	person.IsDeleted = false

	query := `
		CREATE (p:Person {
			id: $id,
			name: $name,
			gender: $gender,
			birthDate: date($birthDate),
			deathDate: CASE WHEN $deathDate IS NULL THEN NULL ELSE date($deathDate) END,
			title: $title,
			bio: $bio,
			addressStreet: $addressStreet,
			addressNeighborhood: $addressNeighborhood,
			addressCity: $addressCity,
			addressProvince: $addressProvince,
			addressPostalCode: $addressPostalCode,
			addressCountry: $addressCountry,
			phonePrimary: $phonePrimary,
			phonePrimaryType: $phonePrimaryType,
			phoneVerified: $phoneVerified,
			isDeleted: $isDeleted,
			createdAt: datetime($createdAt),
			updatedAt: datetime($updatedAt)
		})
		RETURN p
	`

	params := map[string]interface{}{
		"id":                  person.ID,
		"name":                person.Name,
		"gender":              person.Gender,
		"birthDate":           person.BirthDate,
		"deathDate":           person.DeathDate,
		"title":               person.Title,
		"bio":                 person.Bio,
		"addressStreet":       person.AddressStreet,
		"addressNeighborhood": person.AddressNeighborhood,
		"addressCity":         person.AddressCity,
		"addressProvince":     person.AddressProvince,
		"addressPostalCode":   person.AddressPostalCode,
		"addressCountry":      person.AddressCountry,
		"phonePrimary":        person.PhonePrimary,
		"phonePrimaryType":    person.PhonePrimaryType,
		"phoneVerified":       person.PhoneVerified,
		"isDeleted":           person.IsDeleted,
		"createdAt":           person.CreatedAt.Format(time.RFC3339),
		"updatedAt":           person.UpdatedAt.Format(time.RFC3339),
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, err
		}

		node := record.AsMap()["p"].(neo4j.Node)
		return nodeToPerson(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Person), nil
}

func (r *PersonRepository) GetByID(ctx context.Context, id string) (*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (p:Person {id: $id})
		WHERE p.isDeleted = false
		RETURN p
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, err
		}

		node := record.AsMap()["p"].(neo4j.Node)
		return nodeToPerson(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Person), nil
}

func (r *PersonRepository) GetAll(ctx context.Context, limit, offset int) ([]*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (p:Person)
		WHERE p.isDeleted = false
		RETURN p
		ORDER BY p.name
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

		persons := make([]*Person, 0)
		for cursor.Next(ctx) {
			record := cursor.Record()
			node := record.AsMap()["p"].(neo4j.Node)
			person, err := nodeToPerson(node)
			if err != nil {
				return nil, err
			}
			persons = append(persons, person)
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}

		return persons, nil
	})

	if err != nil {
		return nil, err
	}

	return results.([]*Person), nil
}

func (r *PersonRepository) Update(ctx context.Context, id string, updates *Person) (*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (p:Person {id: $id})
		WHERE p.isDeleted = false
		SET p.name = COALESCE($name, p.name),
			p.gender = COALESCE($gender, p.gender),
			p.birthDate = CASE WHEN $birthDate IS NOT NULL THEN date($birthDate) ELSE p.birthDate END,
			p.deathDate = CASE WHEN $deathDate IS NOT NULL THEN date($deathDate) ELSE p.deathDate END,
			p.title = COALESCE($title, p.title),
			p.bio = COALESCE($bio, p.bio),
			p.addressStreet = COALESCE($addressStreet, p.addressStreet),
			p.addressCity = COALESCE($addressCity, p.addressCity),
			p.addressProvince = COALESCE($addressProvince, p.addressProvince),
			p.addressCountry = COALESCE($addressCountry, p.addressCountry),
			p.phonePrimary = COALESCE($phonePrimary, p.phonePrimary),
			p.phonePrimaryType = COALESCE($phonePrimaryType, p.phonePrimaryType),
			p.updatedAt = datetime()
		RETURN p
	`

	params := map[string]interface{}{
		"id":                  id,
		"name":                updates.Name,
		"gender":              updates.Gender,
		"birthDate":           updates.BirthDate,
		"deathDate":           updates.DeathDate,
		"title":               updates.Title,
		"bio":                 updates.Bio,
		"addressStreet":       updates.AddressStreet,
		"addressCity":         updates.AddressCity,
		"addressProvince":     updates.AddressProvince,
		"addressCountry":      updates.AddressCountry,
		"phonePrimary":        updates.PhonePrimary,
		"phonePrimaryType":    updates.PhonePrimaryType,
	}

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("person not found")
		}

		node := record.AsMap()["p"].(neo4j.Node)
		return nodeToPerson(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*Person), nil
}

func (r *PersonRepository) Delete(ctx context.Context, id string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Soft delete - set isDeleted = true
	query := `
		MATCH (p:Person {id: $id})
		WHERE p.isDeleted = false
		SET p.isDeleted = true, p.updatedAt = datetime()
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}
		return cursor.Consume(ctx)
	})

	return err
}

func (r *PersonRepository) Search(ctx context.Context, query string, limit int) ([]*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Search by name (case-insensitive)
	searchQuery := `
		MATCH (p:Person)
		WHERE p.isDeleted = false
		AND toLower(p.name) CONTAINS toLower($query)
		RETURN p
		ORDER BY p.name
		LIMIT $limit
	`

	results, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, searchQuery, map[string]interface{}{
			"query": query,
			"limit": limit,
		})
		if err != nil {
			return nil, err
		}

		persons := make([]*Person, 0)
		for cursor.Next(ctx) {
			record := cursor.Record()
			node := record.AsMap()["p"].(neo4j.Node)
			person, err := nodeToPerson(node)
			if err != nil {
				return nil, err
			}
			persons = append(persons, person)
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}

		return persons, nil
	})

	if err != nil {
		return nil, err
	}

	return results.([]*Person), nil
}

func (r *PersonRepository) AddRelationship(ctx context.Context, personID, targetPersonID, relationshipType string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Determine the reverse relationship
	reverseRelationship := getReverseRelationship(relationshipType)

	query := `
		MATCH (p:Person {id: $personId})
		MATCH (t:Person {id: $targetId})
		WHERE p.isDeleted = false AND t.isDeleted = false
		CREATE (p)-[r:` + relationshipType + ` {createdAt: datetime()}]->(t)
		WITH p, t
		MATCH (p)-[rel:` + relationshipType + `]->(t)
		WITH p, t
		CREATE (t)-[r2:` + reverseRelationship + ` {createdAt: datetime()}]->(p)
		RETURN count(r) + count(r2) as relationshipsCreated
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"personId":   personID,
			"targetId":   targetPersonID,
		})
		if err != nil {
			return nil, err
		}
		return cursor.Consume(ctx)
	})

	return err
}

func getReverseRelationship(relationship string) string {
	reverseMap := map[string]string{
		"PARENT_OF":   "CHILD_OF",
		"CHILD_OF":    "PARENT_OF",
		"SPOUSE_OF":   "SPOUSE_OF",
		"SIBLING_OF":  "SIBLING_OF",
		"PARTNER_OF":  "PARTNER_OF",
		"ADOPTED_BY":  "ADOPT_PARENT",
		"STEP_PARENT": "STEP_CHILD",
		"STEP_CHILD":  "STEP_PARENT",
	}

	if reverse, ok := reverseMap[relationship]; ok {
		return reverse
	}
	return relationship
}

func nodeToPerson(node neo4j.Node) (*Person, error) {
	props := node.Props
	person := &Person{}

	// Helper function to safely get string props
	getString := func(key string) string {
		if v, ok := props[key].(string); ok {
			return v
		}
		return ""
	}

	// Helper to get optional string
	getOptionalString := func(key string) *string {
		if v, ok := props[key].(string); ok && v != "" {
			return &v
		}
		return nil
	}

	person.ID = getString("id")
	person.Name = getString("name")
	person.Gender = getString("gender")
	person.BirthDate = getString("birthDate")
	person.DeathDate = getOptionalString("deathDate")
	person.Title = getOptionalString("title")
	person.Bio = getOptionalString("bio")
	person.AddressStreet = getOptionalString("addressStreet")
	person.AddressCity = getOptionalString("addressCity")
	person.AddressProvince = getOptionalString("addressProvince")
	person.AddressCountry = getOptionalString("addressCountry")
	person.PhonePrimary = getOptionalString("phonePrimary")
	person.PhonePrimaryType = getOptionalString("phonePrimaryType")

	if v, ok := props["phoneVerified"].(bool); ok {
		person.PhoneVerified = &v
	}

	if v, ok := props["isDeleted"].(bool); ok {
		person.IsDeleted = v
	}

	return person, nil
}
