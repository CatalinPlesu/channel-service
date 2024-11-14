package model

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ChannelID  uuid.UUID   `json:"uuid"`
	Name       *string     `json:"name"`
	IsPublic   bool        `json:"is_public"`
	OwnerID    uuid.UUID   `json:"owner_id"`
	UsersAcces []UserAcces `json:"users_acces"`
	CreatedAt  *time.Time   `json:"created_at"`
	UpdatedAt  *time.Time   `json:"updated_at"`
}

type UserAcces struct {
	UserID   uuid.UUID `json:"user_id"`
	IsAdmin  bool      `json:"is_admin"`
	CanWrite bool      `json:"can_write"`
}
