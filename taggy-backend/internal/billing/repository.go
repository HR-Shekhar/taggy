package billing

import (
	"context"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *Repository) CreatePayment(ctx context.Context, arg sqlc.CreatePaymentParams) (sqlc.Payment, error) {
	return r.queries.CreatePayment(ctx, arg)
}

func (r *Repository) GetPaymentByOrderID(ctx context.Context, orderID string) (sqlc.Payment, error) {
	return r.queries.GetPaymentByOrderID(ctx, orderID)
}

func (r *Repository) MarkPaymentPaidIdempotent(ctx context.Context, orderID, paymentID string) (sqlc.Payment, error) {
	return r.queries.MarkPaymentPaidIdempotent(ctx, sqlc.MarkPaymentPaidIdempotentParams{
		RazorpayOrderID: orderID,
		RazorpayPaymentID: pgtype.Text{
			String: paymentID,
			Valid:  paymentID != "",
		},
	})
}

func (r *Repository) UpdateSubscription(ctx context.Context, userID int64, tier sqlc.SubscriptionTier) (sqlc.User, error) {
	return r.queries.UpdateSubscription(ctx, sqlc.UpdateSubscriptionParams{
		ID:           userID,
		Subscription: tier,
	})
}
