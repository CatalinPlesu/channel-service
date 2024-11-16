package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Channel struct {
	bun.BaseModel `bun:"table:channels"` 

	ChannelID  uuid.UUID `bun:"channel_id,pk,type:uuid" json:"uuid"`  
	Name       *string   `bun:"name" json:"name"`                      
	IsPublic   bool      `bun:"is_public" json:"is_public"`             
	OwnerID    uuid.UUID `bun:"owner_id,type:uuid" json:"owner_id"`      
	CreatedAt  *time.Time `bun:"created_at,notnull" json:"created_at"`    
	UpdatedAt  *time.Time `bun:"updated_at,notnull" json:"updated_at"`     
	UsersAcces []UserAcces `bun:"rel:has-many" json:"users_acces"`          
}

type UserAcces struct {
	bun.BaseModel `bun:"table:user_acces"` 

	UserID    uuid.UUID `bun:"user_id,type:uuid,pk" json:"user_id"`      
	ChannelID uuid.UUID `bun:"channel_id,type:uuid,pk" json:"channel_id"`
	IsAdmin   bool      `bun:"is_admin" json:"is_admin"`                 
	CanWrite  bool      `bun:"can_write" json:"can_write"`               
}
