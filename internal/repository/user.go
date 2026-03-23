package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	PersonID     *string   `json:"person_id,omitempty"`
	IsVerified   bool      `json:"is_verified"`
	RefreshToken *string   `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository struct {
	driver neo4j.DriverWithContext
}

func NewUserRepository(driver neo4j.DriverWithContext) *UserRepository {
	return &UserRepository{driver: driver}
}

// Create creates a new user in Neo4j
func (r *UserRepository) Create(ctx context.Context, user *User, password string) (*User, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Check if email already exists
	exists, err := r.EmailExists(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.ID = uuid.New().String()
	user.PasswordHash = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsVerified = false

	// Generate refresh token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	user.RefreshToken = &refreshToken

	query := `
		CREATE (u:User {
			id: $id,
			email: $email,
			passwordHash: $passwordHash,
			role: $role,
			personId: $personId,
			isVerified: $isVerified,
			refreshToken: $refreshToken,
			createdAt: datetime($createdAt),
			updatedAt: datetime($updatedAt)
		})
		RETURN u
	`

	params := map[string]interface{}{
		"id":           user.ID,
		"email":        user.Email,
		"passwordHash": user.PasswordHash,
		"role":         user.Role,
		"personId":     user.PersonID,
		"isVerified":   user.IsVerified,
		"refreshToken": user.RefreshToken,
		"createdAt":    user.CreatedAt.Format(time.RFC3339),
		"updatedAt":    user.UpdatedAt.Format(time.RFC3339),
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

		node := record.AsMap()["u"].(neo4j.Node)
		return nodeToUser(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*User), nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {email: $email})
		RETURN u
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"email": email})
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("user not found")
		}

		node := record.AsMap()["u"].(neo4j.Node)
		return nodeToUser(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*User), nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $id})
		RETURN u
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, errors.New("user not found")
		}

		node := record.AsMap()["u"].(neo4j.Node)
		return nodeToUser(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*User), nil
}

// UpdateRefreshToken updates the refresh token for a user
func (r *UserRepository) UpdateRefreshToken(ctx context.Context, userID, refreshToken string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $id})
		SET u.refreshToken = $refreshToken, u.updatedAt = datetime()
		RETURN u
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"id":           userID,
			"refreshToken": refreshToken,
		})
		if err != nil {
			return nil, err
		}
		return cursor.Single(ctx)
	})

	return err
}

// VerifyPassword checks if the provided password matches the hash
func (r *UserRepository) VerifyPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {email: $email})
		RETURN count(u) > 0 as exists
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{"email": email})
		if err != nil {
			return false, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return false, err
		}

		return record.AsMap()["exists"].(bool), nil
	})

	if err != nil {
		return false, err
	}

	return result.(bool), nil
}

// UpdateProfile updates user profile
func (r *UserRepository) UpdateProfile(ctx context.Context, userID string, personID *string) (*User, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (u:User {id: $id})
		SET u.personId = $personId, u.updatedAt = datetime()
		RETURN u
	`

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cursor, err := tx.Run(ctx, query, map[string]interface{}{
			"id":       userID,
			"personId": personID,
		})
		if err != nil {
			return nil, err
		}

		record, err := cursor.Single(ctx)
		if err != nil {
			return nil, err
		}

		node := record.AsMap()["u"].(neo4j.Node)
		return nodeToUser(node)
	})

	if err != nil {
		return nil, err
	}

	return result.(*User), nil
}

// Helper function to convert Neo4j node to User
func nodeToUser(node neo4j.Node) (*User, error) {
	props := node.Props
	user := &User{}

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

	getBool := func(key string) bool {
		if v, ok := props[key].(bool); ok {
			return v
		}
		return false
	}

	getTime := func(key string) time.Time {
		switch v := props[key].(type) {
		case time.Time:
			return v
		case dbtype.LocalDateTime:
			return v.Time()
		default:
			return time.Time{}
		}
	}

	user.ID = getString("id")
	user.Email = getString("email")
	user.PasswordHash = getString("passwordHash")
	user.Role = getString("role")
	user.PersonID = getOptionalString("personId")
	user.IsVerified = getBool("isVerified")
	user.RefreshToken = getOptionalString("refreshToken")
	user.CreatedAt = getTime("createdAt")
	user.UpdatedAt = getTime("updatedAt")

	return user, nil
}

// generateRefreshToken generates a cryptographically secure refresh token
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
