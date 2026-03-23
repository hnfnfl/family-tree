package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	neo4jtime "github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

type Person struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name" binding:"required"`
	Gender              string    `json:"gender" binding:"required,oneof=male female"`
	BirthDate           string    `json:"birthDate"`
	DeathDate           *string   `json:"deathDate,omitempty"`
	Title               *string   `json:"title,omitempty"`
	Bio                 *string   `json:"bio,omitempty"`
	AddressStreet       *string   `json:"addressStreet,omitempty"`
	AddressNeighborhood *string   `json:"addressNeighborhood,omitempty"`
	AddressCity         *string   `json:"addressCity,omitempty"`
	AddressProvince     *string   `json:"addressProvince,omitempty"`
	AddressPostalCode   *string   `json:"addressPostalCode,omitempty"`
	AddressCountry      *string   `json:"addressCountry,omitempty"`
	PhonePrimary        *string   `json:"phonePrimary,omitempty"`
	PhonePrimaryType    *string   `json:"phonePrimaryType,omitempty"`
	PhoneVerified       *bool     `json:"phoneVerified,omitempty"`
	IsDeleted           bool      `json:"isDeleted"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type PersonRepository struct {
	driver   neo4j.DriverWithContext
	database string
}

func NewPersonRepository(driver neo4j.DriverWithContext) *PersonRepository {
	return &PersonRepository{driver: driver, database: "neo4j"}
}

func (r *PersonRepository) Create(ctx context.Context, person *Person) (*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	person.ID = uuid.New().String()
	person.CreatedAt = time.Now()
	person.UpdatedAt = time.Now()
	person.IsDeleted = false

	// BirthDate required validation
	if person.BirthDate == "" {
		return nil, errors.New("birthDate is required")
	}

	query := `
		CREATE (p:Person {
			id: $id,
			name: $name,
			gender: $gender,
			birthDate: date($birthDate),
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
		"title":               nilString(person.Title),
		"bio":                 nilString(person.Bio),
		"addressStreet":       nilString(person.AddressStreet),
		"addressNeighborhood": nilString(person.AddressNeighborhood),
		"addressCity":         nilString(person.AddressCity),
		"addressProvince":     nilString(person.AddressProvince),
		"addressPostalCode":   nilString(person.AddressPostalCode),
		"addressCountry":      nilString(person.AddressCountry),
		"phonePrimary":        nilString(person.PhonePrimary),
		"phonePrimaryType":    nilString(person.PhonePrimaryType),
		"phoneVerified":       nilBool(person.PhoneVerified),
		"isDeleted":           false,
		"createdAt":           person.CreatedAt.Format(time.RFC3339),
		"updatedAt":           person.UpdatedAt.Format(time.RFC3339),
	}

	// Fix: use ExecuteWrite for CREATE
	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
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

func (r *PersonRepository) GetAll(ctx context.Context, limit, offset int) ([]*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	query := `
		MATCH (p:Person {id: $id})
		WHERE p.isDeleted = false
		SET p.name = COALESCE($name, p.name),
			p.gender = COALESCE($gender, p.gender),
			p.birthDate = CASE WHEN $birthDate <> '' THEN date($birthDate) ELSE p.birthDate END,
			p.deathDate = CASE WHEN $deathDate IS NOT NULL THEN date($deathDate) ELSE p.deathDate END,
			p.title = COALESCE($title, p.title),
			p.bio = COALESCE($bio, p.bio),
			p.addressStreet = COALESCE($addressStreet, p.addressStreet),
			p.addressNeighborhood = COALESCE($addressNeighborhood, p.addressNeighborhood),
			p.addressCity = COALESCE($addressCity, p.addressCity),
			p.addressProvince = COALESCE($addressProvince, p.addressProvince),
			p.addressPostalCode = COALESCE($addressPostalCode, p.addressPostalCode),
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
		"deathDate":           nilString(updates.DeathDate),
		"title":               nilString(updates.Title),
		"bio":                 nilString(updates.Bio),
		"addressStreet":       nilString(updates.AddressStreet),
		"addressNeighborhood": nilString(updates.AddressNeighborhood),
		"addressCity":         nilString(updates.AddressCity),
		"addressProvince":     nilString(updates.AddressProvince),
		"addressPostalCode":   nilString(updates.AddressPostalCode),
		"addressCountry":      nilString(updates.AddressCountry),
		"phonePrimary":        nilString(updates.PhonePrimary),
		"phonePrimaryType":    nilString(updates.PhonePrimaryType),
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	query := `
		MATCH (p:Person {id: $id})
		WHERE p.isDeleted = false
		SET p.isDeleted = true, p.updatedAt = datetime()
		RETURN p.id
	`

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}
		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("person not found")
		}
		return record, nil
	})

	if err != nil {
		return err
	}
	_ = result
	return nil
}

func (r *PersonRepository) Search(ctx context.Context, query string, limit int) ([]*Person, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

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
	if personID == targetPersonID {
		return errors.New("cannot create relationship with self")
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer session.Close(ctx)

	reverseRelationship := getReverseRelationship(relationshipType)

	// Use MERGE to avoid duplicate relationships
	query := fmt.Sprintf(`
		MATCH (p:Person {id: $personId})
		MATCH (t:Person {id: $targetId})
		WHERE p.isDeleted = false AND t.isDeleted = false
		MERGE (p)-[r:%s]->(t)
		ON CREATE SET r.createdAt = datetime()
		MERGE (t)-[r2:%s]->(p)
		ON CREATE SET r2.createdAt = datetime()
		RETURN p.id, t.id
	`, relationshipType, reverseRelationship)

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"personId": personID,
			"targetId": targetPersonID,
		})
		if err != nil {
			return nil, err
		}
		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("one or both persons not found")
		}
		return record, nil
	})

	if err != nil {
		return err
	}
	_ = result
	return nil
}

func getReverseRelationship(relationship string) string {
	reverseMap := map[string]string{
		"PARENT_OF":   "CHILD_OF",
		"CHILD_OF":    "PARENT_OF",
		"SPOUSE_OF":   "SPOUSE_OF",
		"SIBLING_OF":  "SIBLING_OF",
		"PARTNER_OF":  "PARTNER_OF",
		"ADOPTED_BY":  "ADOPTS",
		"ADOPTS":      "ADOPTED_BY",
		"STEP_PARENT": "STEP_CHILD",
		"STEP_CHILD":  "STEP_PARENT",
	}
	if reverse, ok := reverseMap[relationship]; ok {
		return reverse
	}
	return relationship
}

// nodeToPerson safely converts a Neo4j node to Person struct.
// Handles Neo4j date/datetime types properly.
func nodeToPerson(node neo4j.Node) (*Person, error) {
	props := node.Props
	person := &Person{}

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

	// Neo4j stores date() as neo4jtime.Date, not string
	getDateString := func(key string) string {
		switch v := props[key].(type) {
		case string:
			return v
		case neo4jtime.Date:
			return fmt.Sprintf("%04d-%02d-%02d", v.Time().Year(), v.Time().Month(), v.Time().Day())
		default:
			return ""
		}
	}

	getOptionalDateString := func(key string) *string {
		switch v := props[key].(type) {
		case string:
			if v != "" {
				return &v
			}
		case neo4jtime.Date:
			s := fmt.Sprintf("%04d-%02d-%02d", v.Time().Year(), v.Time().Month(), v.Time().Day())
			return &s
		}
		return nil
	}

	// Neo4j datetime() returns time.Time directly
	getTime := func(key string) time.Time {
		switch v := props[key].(type) {
		case time.Time:
			return v
		case neo4jtime.LocalDateTime:
			return v.Time()
		default:
			return time.Time{}
		}
	}

	person.ID = getString("id")
	person.Name = getString("name")
	person.Gender = getString("gender")
	person.BirthDate = getDateString("birthDate")
	person.DeathDate = getOptionalDateString("deathDate")
	person.Title = getOptionalString("title")
	person.Bio = getOptionalString("bio")
	person.AddressStreet = getOptionalString("addressStreet")
	person.AddressNeighborhood = getOptionalString("addressNeighborhood")
	person.AddressCity = getOptionalString("addressCity")
	person.AddressProvince = getOptionalString("addressProvince")
	person.AddressPostalCode = getOptionalString("addressPostalCode")
	person.AddressCountry = getOptionalString("addressCountry")
	person.PhonePrimary = getOptionalString("phonePrimary")
	person.PhonePrimaryType = getOptionalString("phonePrimaryType")
	person.CreatedAt = getTime("createdAt")
	person.UpdatedAt = getTime("updatedAt")

	if v, ok := props["phoneVerified"].(bool); ok {
		person.PhoneVerified = &v
	}
	if v, ok := props["isDeleted"].(bool); ok {
		person.IsDeleted = v
	}

	return person, nil
}

// nilString safely dereferences a *string for use in Cypher params
func nilString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

// nilBool safely dereferences a *bool for use in Cypher params
func nilBool(b *bool) interface{} {
	if b == nil {
		return nil
	}
	return *b
}
