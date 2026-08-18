package notification

import (
	"context"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

func (r *Repository) GetUserByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, publicID)
}

func (r *Repository) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return uuid.UUID{}, err
	}
	return user.PublicID, nil
}

func (r *Repository) CreateNotification(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
	return r.queries.CreateNotification(ctx, arg)
}

func (r *Repository) ListNotificationsByUserID(
	ctx context.Context,
	userID int64,
	unreadOnly bool,
	limit int32,
) ([]sqlc.Notification, error) {
	return r.queries.ListNotificationsByUserID(ctx, sqlc.ListNotificationsByUserIDParams{
		UserID:     userID,
		UnreadOnly: unreadOnly,
		NotifLimit: limit,
	})
}

func (r *Repository) GetNotificationByIDAndUserID(ctx context.Context, id, userID int64) (sqlc.Notification, error) {
	return r.queries.GetNotificationByIDAndUserID(ctx, sqlc.GetNotificationByIDAndUserIDParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, userID int64) (sqlc.Notification, error) {
	return r.queries.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *Repository) MarkAllNotificationsRead(ctx context.Context, userID int64) (int64, error) {
	return r.queries.MarkAllNotificationsRead(ctx, userID)
}

func (r *Repository) CountUnreadNotifications(ctx context.Context, userID int64) (int64, error) {
	return r.queries.CountUnreadNotifications(ctx, userID)
}

func (r *Repository) DeleteReadNotificationsByUserID(ctx context.Context, userID int64) (int64, error) {
	return r.queries.DeleteReadNotificationsByUserID(ctx, userID)
}
