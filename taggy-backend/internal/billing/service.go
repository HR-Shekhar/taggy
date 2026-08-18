package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	razorpay "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/utils"
	"github.com/rs/zerolog"
)

type Config struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
	AmountPaise   int32
	Currency      string
}

type Service struct {
	repo   *Repository
	cfg    Config
	client *razorpay.Client
	log    zerolog.Logger
}

func NewService(repo *Repository, cfg Config, log zerolog.Logger) *Service {
	if cfg.AmountPaise <= 0 {
		cfg.AmountPaise = 49900
	}
	if strings.TrimSpace(cfg.Currency) == "" {
		cfg.Currency = "INR"
	}
	cfg.Currency = strings.ToUpper(strings.TrimSpace(cfg.Currency))

	var client *razorpay.Client
	if strings.TrimSpace(cfg.KeyID) != "" && strings.TrimSpace(cfg.KeySecret) != "" {
		client = razorpay.NewClient(cfg.KeyID, cfg.KeySecret)
	}
	return &Service{repo: repo, cfg: cfg, client: client, log: log}
}

func (s *Service) Available() bool {
	return s != nil && s.client != nil
}

func (s *Service) Status(ctx context.Context, userPublicID uuid.UUID) (statusResponse, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statusResponse{}, logging.Reject(s.log, ErrPaymentForbidden, "billing status: user not found")
		}
		return statusResponse{}, logging.Unexpected(s.log, err, "billing status get user failed")
	}
	return statusResponse{
		Subscription:       string(user.Subscription),
		PremiumAmountPaise: s.cfg.AmountPaise,
		Currency:           s.cfg.Currency,
		BillingConfigured:  s.Available(),
	}, nil
}

func (s *Service) CreateOrder(ctx context.Context, userPublicID uuid.UUID) (createOrderResponse, error) {
	if !s.Available() {
		return createOrderResponse{}, logging.Reject(s.log, ErrBillingNotConfigured, "create order rejected: billing not configured")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return createOrderResponse{}, logging.Reject(s.log, ErrPaymentForbidden, "create order: user not found")
		}
		return createOrderResponse{}, logging.Unexpected(s.log, err, "create order get user failed")
	}
	if user.Subscription == sqlc.SubscriptionTierPREMIUM {
		return createOrderResponse{}, logging.Reject(s.log, ErrAlreadyPremium, "create order rejected: already premium")
	}

	receipt := fmt.Sprintf("tg_%s_%d", strings.ReplaceAll(user.PublicID.String(), "-", "")[:12], time.Now().Unix()%1_000_000)
	if len(receipt) > 40 {
		receipt = receipt[:40]
	}

	data := map[string]interface{}{
		"amount":   s.cfg.AmountPaise,
		"currency": s.cfg.Currency,
		"receipt":  receipt,
		"notes": map[string]interface{}{
			"user_public_id": user.PublicID.String(),
			"username":       user.Username,
			"product":        "taggy_premium",
		},
	}

	body, err := s.client.Order.Create(data, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("razorpay order create failed")
		return createOrderResponse{}, ErrOrderCreateFailed
	}

	orderID, _ := body["id"].(string)
	if orderID == "" {
		s.log.Error().Interface("body", body).Msg("razorpay order missing id")
		return createOrderResponse{}, ErrOrderCreateFailed
	}

	_, err = s.repo.CreatePayment(ctx, sqlc.CreatePaymentParams{
		UserID:          user.ID,
		RazorpayOrderID: orderID,
		Amount:          s.cfg.AmountPaise,
		Currency:        s.cfg.Currency,
		Receipt:         receipt,
	})
	if err != nil {
		return createOrderResponse{}, logging.Unexpected(s.log, err, "persist payment row failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("order_id", orderID).
		Int32("amount", s.cfg.AmountPaise).
		Msg("razorpay order created")

	return createOrderResponse{
		KeyID:        s.cfg.KeyID,
		OrderID:      orderID,
		Amount:       s.cfg.AmountPaise,
		Currency:     s.cfg.Currency,
		Name:         "Taggy",
		Description:  "Taggy Premium — unlimited skills",
		PrefillName:  user.Name,
		PrefillEmail: user.Email,
	}, nil
}

func (s *Service) VerifyCheckout(ctx context.Context, userPublicID uuid.UUID, orderID, paymentID, signature string) (verifyResponse, error) {
	if !s.Available() {
		return verifyResponse{}, logging.Reject(s.log, ErrBillingNotConfigured, "verify rejected: billing not configured")
	}

	orderID = strings.TrimSpace(orderID)
	paymentID = strings.TrimSpace(paymentID)
	signature = strings.TrimSpace(signature)
	if orderID == "" || paymentID == "" || signature == "" {
		return verifyResponse{}, ErrInvalidSignature
	}

	params := map[string]interface{}{
		"razorpay_order_id":   orderID,
		"razorpay_payment_id": paymentID,
	}
	if !utils.VerifyPaymentSignature(params, signature, s.cfg.KeySecret) {
		return verifyResponse{}, logging.Reject(s.log, ErrInvalidSignature, "verify rejected: bad signature")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return verifyResponse{}, logging.Reject(s.log, ErrPaymentForbidden, "verify: user not found")
		}
		return verifyResponse{}, logging.Unexpected(s.log, err, "verify get user failed")
	}

	payment, err := s.repo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return verifyResponse{}, logging.Reject(s.log, ErrPaymentNotFound, "verify: payment not found")
		}
		return verifyResponse{}, logging.Unexpected(s.log, err, "verify get payment failed")
	}
	if payment.UserID != user.ID {
		return verifyResponse{}, logging.Reject(s.log, ErrPaymentForbidden, "verify: payment user mismatch")
	}

	if payment.Status == sqlc.PaymentStatusPAID {
		if user.Subscription != sqlc.SubscriptionTierPREMIUM {
			user, err = s.repo.UpdateSubscription(ctx, user.ID, sqlc.SubscriptionTierPREMIUM)
			if err != nil {
				return verifyResponse{}, logging.Unexpected(s.log, err, "verify promote already-paid failed")
			}
		}
		return verifyResponse{
			Subscription: string(user.Subscription),
			Message:      "Premium already unlocked",
		}, nil
	}

	if _, err := s.repo.MarkPaymentPaidIdempotent(ctx, orderID, paymentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return verifyResponse{}, logging.Reject(s.log, ErrPaymentNotPayable, "verify: payment not in payable state")
		}
		return verifyResponse{}, logging.Unexpected(s.log, err, "verify mark paid failed")
	}

	user, err = s.repo.UpdateSubscription(ctx, user.ID, sqlc.SubscriptionTierPREMIUM)
	if err != nil {
		return verifyResponse{}, logging.Unexpected(s.log, err, "verify update subscription failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("order_id", orderID).
		Str("payment_id", paymentID).
		Msg("premium unlocked via checkout verify")

	return verifyResponse{
		Subscription: string(user.Subscription),
		Message:      "Welcome to Taggy Premium",
	}, nil
}

// HandleWebhook verifies Razorpay webhook signature and promotes the user on paid events.
func (s *Service) HandleWebhook(ctx context.Context, rawBody []byte, signature string) error {
	secret := strings.TrimSpace(s.cfg.WebhookSecret)
	if secret == "" {
		secret = s.cfg.KeySecret
	}
	if secret == "" {
		return logging.Reject(s.log, ErrBillingNotConfigured, "webhook rejected: no secret")
	}
	if !utils.VerifyWebhookSignature(string(rawBody), signature, secret) {
		return logging.Reject(s.log, ErrInvalidSignature, "webhook rejected: bad signature")
	}

	var envelope struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID      string `json:"id"`
					OrderID string `json:"order_id"`
					Status  string `json:"status"`
				} `json:"entity"`
			} `json:"payment"`
			Order struct {
				Entity struct {
					ID string `json:"id"`
				} `json:"entity"`
			} `json:"order"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return logging.Reject(s.log, ErrInvalidSignature, "webhook rejected: bad json")
	}

	event := envelope.Event
	orderID := envelope.Payload.Payment.Entity.OrderID
	paymentID := envelope.Payload.Payment.Entity.ID
	if orderID == "" {
		orderID = envelope.Payload.Order.Entity.ID
	}
	if orderID == "" {
		s.log.Info().Str("event", event).Msg("webhook ignored: no order id")
		return nil
	}

	switch event {
	case "payment.captured", "order.paid", "payment.authorized":
		// continue
	default:
		s.log.Info().Str("event", event).Msg("webhook ignored: unhandled event")
		return nil
	}

	payment, err := s.repo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn().Str("order_id", orderID).Msg("webhook: unknown order")
			return nil
		}
		return logging.Unexpected(s.log, err, "webhook get payment failed")
	}

	if payment.Status != sqlc.PaymentStatusPAID {
		if _, err := s.repo.MarkPaymentPaidIdempotent(ctx, orderID, paymentID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return logging.Unexpected(s.log, err, "webhook mark paid failed")
		}
	}

	if _, err := s.repo.UpdateSubscription(ctx, payment.UserID, sqlc.SubscriptionTierPREMIUM); err != nil {
		return logging.Unexpected(s.log, err, "webhook update subscription failed")
	}

	s.log.Info().
		Str("event", event).
		Str("order_id", orderID).
		Str("payment_id", paymentID).
		Int64("user_id", payment.UserID).
		Msg("premium unlocked via webhook")
	return nil
}
