package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/models"
)

func CreateUser(db *sql.DB, email, passwordHash, name string) (models.User, error) {
	var u models.User
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, name, password_hash, created_at, updated_at`,
		email, passwordHash, name,
	).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var u models.User
	err := db.QueryRow(
		`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
