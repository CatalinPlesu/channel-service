package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/CatalinPlesu/channel-service/model"
	"github.com/CatalinPlesu/channel-service/repository/channel"
)

type Channel struct {
	RdRepo *channel.RedisRepo
	PgRepo *channel.PostgresRepo
}

func (h *Channel) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       *string           `json:"name"`
		IsPublic   bool              `json:"is_public"`
		OwnerID    uuid.UUID         `json:"owner_id"`
		UsersAcces []model.UserAcces `json:"users_acces"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	channel := model.Channel{
		ChannelID:  uuid.New(),
		Name:       body.Name,
		IsPublic:   body.IsPublic,
		OwnerID:    body.OwnerID,
		UsersAcces: body.UsersAcces,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	err := h.PgRepo.Insert(r.Context(), channel)
	if err != nil {
		fmt.Println("failed to insert channel:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(channel)
	if err != nil {
		fmt.Println("failed to marshal channel:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (h *Channel) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := h.PgRepo.FindAll(r.Context(), channel.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to find all channels:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Channel `json:"items"`
		Next  uint64          `json:"next,omitempty"`
	}
	response.Items = res.Channels
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal channels:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *Channel) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	channelID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.PgRepo.FindByID(r.Context(), channelID)
	if errors.Is(err, channel.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find channel by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		fmt.Println("failed to marshal channel:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Channel) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       *string     `json:"name,omitempty"`
		IsPublic   *bool        `json:"is_public,omitempty"`
		OwnerID    *uuid.UUID   `json:"owner_id,omitempty"`
		UsersAcces *[]model.UserAcces `json:"users_acces,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	channelID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theChannel, err := h.PgRepo.FindByID(r.Context(), channelID)
	if errors.Is(err, channel.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find channel by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	if body.Name != nil {
		theChannel.Name = body.Name
	}
	if body.IsPublic != nil {
		theChannel.IsPublic = *body.IsPublic
	}
	if body.OwnerID != nil {
		theChannel.OwnerID = *body.OwnerID
	}
	if body.UsersAcces != nil {
		theChannel.UsersAcces = *body.UsersAcces
	}
	theChannel.UpdatedAt = &now

	err = h.PgRepo.Update(r.Context(), theChannel)
	if err != nil {
		fmt.Println("failed to update channel:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theChannel); err != nil {
		fmt.Println("failed to marshal channel:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Channel) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	channelID, err := uuid.Parse(idParam) // Parse as UUID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.PgRepo.DeleteByID(r.Context(), channelID)
	if errors.Is(err, channel.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to delete channel by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
