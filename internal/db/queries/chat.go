package queries

import (
	"database/sql"

	"github.com/yourname/pocket-api/internal/models"
)

func SaveChatMessage(db *sql.DB, userID, role, content string) error {
	_, err := db.Exec(
		`INSERT INTO chat_messages (user_id, role, content) VALUES ($1, $2, $3)`,
		userID, role, content,
	)
	return err
}

func GetChatHistory(db *sql.DB, userID string, limit int) ([]models.ChatMessage, error) {
	rows, err := db.Query(
		`SELECT id, user_id, role, content, created_at
		 FROM (
		   SELECT id, user_id, role, content, created_at
		   FROM chat_messages
		   WHERE user_id = $1
		   ORDER BY created_at DESC
		   LIMIT $2
		 ) sub
		 ORDER BY created_at ASC`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []models.ChatMessage
	for rows.Next() {
		var m models.ChatMessage
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
