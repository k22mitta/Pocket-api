package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/models"
)

func CreateUser(db *sql.DB, email, passwordHash string) (models.User, error) {
	var u models.User
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash, created_at, updated_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var u models.User
	err := db.QueryRow(
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
