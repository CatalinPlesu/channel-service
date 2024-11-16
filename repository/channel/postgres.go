package channel

import (
	"context"
	"fmt"

	"github.com/CatalinPlesu/channel-service/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PostgresRepo struct {
	DB *bun.DB
}

func NewPostgresRepo(db *bun.DB) *PostgresRepo {
	return &PostgresRepo{DB: db}
}

func (p *PostgresRepo) Migrate(ctx context.Context) error {
	_, err := p.DB.NewCreateTable().
		Model((*model.Channel)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create channels table: %w", err)
	}
	return nil
}

func (p *PostgresRepo) Insert(ctx context.Context, channel model.Channel) error {
	_, err := p.DB.NewInsert().Model(&channel).Exec(ctx)
	if err != nil {
		p.Migrate(ctx)
		return fmt.Errorf("failed to insert channel: %w", err)
	}
	return nil
}

func (p *PostgresRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	var channel model.Channel
	err := p.DB.NewSelect().Model(&channel).Where("channel_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel by ID: %w", err)
	}
	return &channel, nil
}

func (p *PostgresRepo) FindByName(ctx context.Context, name string) (*model.Channel, error) {
	var channel model.Channel
	err := p.DB.NewSelect().Model(&channel).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel by ID: %w", err)
	}
	return &channel, nil
}

func (p *PostgresRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := p.DB.NewDelete().Model((*model.Channel)(nil)).Where("channel_id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	return nil
}

func (p *PostgresRepo) Update(ctx context.Context, channel *model.Channel) error {
	_, err := p.DB.NewUpdate().Model(channel).Where("channel_id = ?", channel.ChannelID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}
	return nil
}

type ChannelPage struct {
	Channels  []model.Channel
	Cursor uint64
}

func (r *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (ChannelPage, error) {
	var channels []model.Channel

	// Query the database for channels
	query := r.DB.NewSelect().
		Model(&channels).
		Order("channel_id ASC").
		Limit(int(page.Size))

	// If a cursor is provided, only retrieve channels with an ID greater than the cursor
	if page.Offset > 0 {
		query.Where("channel_id > ?", page.Offset)
	}

	// Execute the query
	err := query.Scan(ctx)
	if err != nil {
		return ChannelPage{}, fmt.Errorf("failed to retrieve channels: %w", err)
	}

	// If no channels were found, return an empty result
	if len(channels) == 0 {
		return ChannelPage{
			Channels:  []model.Channel{},
			Cursor: 0,
		}, nil
	}

	return ChannelPage{
		Channels:  channels,
		Cursor: page.Size + 50,
	}, nil
}
