package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"

	"github.com/CatalinPlesu/channel-service/model"
)

type RedisRepo struct {
	Client *redis.Client
}

func channelIDKey(id uuid.UUID) string {
	return fmt.Sprintf("channel:%s", id.String())
}

func (r *RedisRepo) Insert(ctx context.Context, channel model.Channel) error {
	data, err := json.Marshal(channel)
	if err != nil {
		return fmt.Errorf("failed to encode channel: %w", err)
	}

	key := channelIDKey(channel.ChannelID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set channel: %w", err)
	}

	if err := txn.SAdd(ctx, "channels", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add channel to set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("channel does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uuid.UUID) (model.Channel, error) {
	key := channelIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Channel{}, ErrNotExist
	} else if err != nil {
		return model.Channel{}, fmt.Errorf("failed to get channel: %w", err)
	}

	var channel model.Channel
	err = json.Unmarshal([]byte(value), &channel)
	if err != nil {
		return model.Channel{}, fmt.Errorf("failed to decode channel json: %w", err)
	}

	return channel, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	key := channelIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	if err := txn.SRem(ctx, "channels", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove channel from set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, channel model.Channel) error {
	data, err := json.Marshal(channel)
	if err != nil {
		return fmt.Errorf("failed to encode channel: %w", err)
	}

	key := channelIDKey(channel.ChannelID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Channels  []model.Channel
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "channels", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get channel ids: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Channels: []model.Channel{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get channels: %w", err)
	}

	channels := make([]model.Channel, len(xs))

	for i, x := range xs {
		x := x.(string)
		var channel model.Channel

		err := json.Unmarshal([]byte(x), &channel)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode channel json: %w", err)
		}

		channels[i] = channel
	}

	return FindResult{
		Channels: channels,
		Cursor: cursor,
	}, nil
}
