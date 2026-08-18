package billing

import (
	"io"
	"net/http"

	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Handler struct {
	service *Service
	log     zerolog.Logger
}

func NewHandler(service *Service, log zerolog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) Status(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.Status(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateOrder(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.CreateOrder(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}
	log.Info().Str("order_id", resp.OrderID).Msg("checkout order handled")
	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Verify(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	var req verifyRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.service.VerifyCheckout(
		c.Request().Context(),
		userPublicID,
		req.RazorpayOrderID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
	)
	if err != nil {
		return err
	}
	log.Info().Str("subscription", resp.Subscription).Msg("checkout verify handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Webhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return apperrors.ErrBadRequest
	}
	sig := c.Request().Header.Get("X-Razorpay-Signature")
	if err := h.service.HandleWebhook(c.Request().Context(), body, sig); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
