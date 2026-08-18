package progress

import (
	"net/http"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (h *Handler) LogStudySession(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	var req logStudySessionRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid study session payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	studiedAt := time.Now()
	if req.StudiedAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.StudiedAt)
		if err != nil {
			return apperrors.ErrBadRequest
		}
		studiedAt = parsed
	}

	session, streak, err := h.service.LogStudySession(c.Request().Context(), userPublicID, LogStudySessionInput{
		SkillSlug:       req.SkillSlug,
		DurationMinutes: req.DurationMinutes,
		Notes:           req.Notes,
		StudiedAt:       studiedAt,
	})
	if err != nil {
		return err
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Str("skill_slug", req.SkillSlug).
		Msg("study session handled")

	return c.JSON(http.StatusCreated, logStudySessionResponse{
		Session: toStudySessionResponse(session, req.SkillSlug),
		Streak:  toStreakResponse(streak),
	})
}

func (h *Handler) ListStudySessions(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	var skillSlug *string
	if raw := c.QueryParam("skill_slug"); raw != "" {
		skillSlug = &raw
	}

	sessions, err := h.service.ListStudySessions(c.Request().Context(), userPublicID, skillSlug)
	if err != nil {
		return err
	}

	resp := make([]studySessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, toStudySessionRowResponse(s))
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Int("count", len(resp)).
		Msg("study sessions listed")

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetStreak(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	streak, err := h.service.GetStreak(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}

	log.Debug().Str("user_id", userPublicID.String()).Msg("streak returned")
	return c.JSON(http.StatusOK, toStreakResponse(streak))
}

func (h *Handler) GetProgressSummary(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	summary, err := h.service.GetProgressSummary(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}

	log.Info().Str("user_id", userPublicID.String()).Msg("progress summary returned")

	return c.JSON(http.StatusOK, progressSummaryResponse{
		TotalMinutes:   summary.TotalMinutes,
		WeeklyMinutes:  summary.WeeklyMinutes,
		MonthlyMinutes: summary.MonthlyMinutes,
		CurrentStreak:  summary.CurrentStreak,
		LongestStreak:  summary.LongestStreak,
	})
}

func scopedUserPublicID(c echo.Context, service *Service) (uuid.UUID, error) {
	return auth.ResolveScopedUserPublicID(
		c,
		c.Param("username"),
		service.GetUserPublicIDByUsername,
	)
}

func toStudySessionResponse(s sqlc.StudySession, skillSlug string) studySessionResponse {
	resp := studySessionResponse{
		SkillSlug:       skillSlug,
		DurationMinutes: s.DurationMinutes,
		StudiedAt:       formatTime(s.StudiedAt),
		CreatedAt:       formatTime(s.CreatedAt),
	}
	if s.Notes.Valid {
		resp.Notes = &s.Notes.String
	}
	return resp
}

func toStudySessionRowResponse(s sqlc.ListStudySessionsByUserIDRow) studySessionResponse {
	resp := studySessionResponse{
		SkillSlug:       s.SkillSlug,
		DurationMinutes: s.DurationMinutes,
		StudiedAt:       formatTime(s.StudiedAt),
		CreatedAt:       formatTime(s.CreatedAt),
	}
	if s.Notes.Valid {
		resp.Notes = &s.Notes.String
	}
	return resp
}

func toStreakResponse(s sqlc.Streak) streakResponse {
	resp := streakResponse{
		CurrentStreak: s.CurrentStreak,
		LongestStreak: s.LongestStreak,
		FreezeCount:   s.FreezeCount,
	}
	if s.LastActivityDate.Valid {
		d := s.LastActivityDate.Time.Format("2006-01-02")
		resp.LastActivityDate = &d
	}
	return resp
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}
