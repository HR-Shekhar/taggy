package quiz

import (
	"net/http"
	"strconv"

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

func (h *Handler) Start(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.StartQuiz(c.Request().Context(), userPublicID, c.Param("podSlug"))
	if err != nil {
		return err
	}
	log.Info().Str("quiz_id", resp.ID).Str("status", resp.Status).Msg("pod quiz accepted for generation")
	return c.JSON(http.StatusAccepted, resp)
}

func (h *Handler) Get(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetQuiz(c.Request().Context(), userPublicID, c.Param("podSlug"), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListMine(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.ListMine(c.Request().Context(), userPublicID, c.Param("podSlug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) StartQuestion(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	order, err := strconv.ParseInt(c.Param("order"), 10, 32)
	if err != nil || order < 1 || order > 10 {
		return apperrors.ErrBadRequest
	}
	resp, err := h.service.StartQuestion(
		c.Request().Context(),
		userPublicID,
		c.Param("podSlug"),
		c.Param("id"),
		int32(order),
	)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) AnswerQuestion(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	order, err := strconv.ParseInt(c.Param("order"), 10, 32)
	if err != nil || order < 1 || order > 10 {
		return apperrors.ErrBadRequest
	}
	var req answerRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	resp, err := h.service.AnswerQuestion(
		c.Request().Context(),
		userPublicID,
		c.Param("podSlug"),
		c.Param("id"),
		int32(order),
		req.SelectedIndices,
	)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Complete(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.CompleteQuiz(c.Request().Context(), userPublicID, c.Param("podSlug"), c.Param("id"))
	if err != nil {
		return err
	}
	log.Info().Str("quiz_id", resp.ID).Int("score", resp.Score).Msg("pod quiz completed")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) PodLeaderboard(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.PodLeaderboard(c.Request().Context(), userPublicID, c.Param("podSlug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CommunityLeaderboard(c echo.Context) error {
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.service.CommunityLeaderboard(c.Request().Context(), userPublicID, c.Param("skillSlug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
