package pod

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

func (h *Handler) CreatePod(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}

	var req createPodRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	detail, err := h.service.CreatePod(c.Request().Context(), userPublicID, skillSlug, CreatePodInput{
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	log.Info().Str("pod_slug", detail.Slug).Msg("create pod handled")
	return c.JSON(http.StatusCreated, toPodResponse(detail, 1))
}

func (h *Handler) ListPodsBySkill(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	skillSlug := c.Param("skillSlug")
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}

	rows, err := h.service.ListPodsBySkill(c.Request().Context(), skillSlug)
	if err != nil {
		return err
	}

	resp := make([]podResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toListPodResponse(row))
	}

	log.Info().Str("skill_slug", skillSlug).Int("count", len(resp)).Msg("list pods handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetPod(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	detail, members, pending, count, err := h.service.GetPod(c.Request().Context(), podSlug)
	if err != nil {
		return err
	}

	memberResp := make([]memberResponse, 0, len(members))
	for _, m := range members {
		memberResp = append(memberResp, memberResponse{
			Username: m.Username,
			Name:     m.UserName,
			Role:     string(m.Role),
			Status:   string(m.Status),
			JoinedAt: formatOptionalTime(m.JoinedAt),
		})
	}

	pendingResp := make([]memberResponse, 0, len(pending))
	for _, m := range pending {
		pendingResp = append(pendingResp, memberResponse{
			Username: m.Username,
			Name:     m.UserName,
			Role:     string(m.Role),
			Status:   string(m.Status),
			JoinedAt: formatOptionalTime(m.JoinedAt),
		})
	}

	log.Info().
		Str("pod_slug", podSlug).
		Int("members", len(memberResp)).
		Int("join_requests", len(pendingResp)).
		Msg("get pod handled")
	return c.JSON(http.StatusOK, podDetailResponse{
		Pod:          toPodResponse(detail, count),
		Members:      memberResp,
		JoinRequests: pendingResp,
	})
}

func (h *Handler) ListMyPods(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	rows, err := h.service.ListMyPods(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}

	resp := make([]membershipResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, membershipResponse{
			PodSlug:   row.PodSlug,
			PodName:   row.PodName,
			SkillSlug: row.SkillSlug,
			SkillName: row.SkillName,
			Status:    string(row.Status),
			Role:      string(row.Role),
			JoinedAt:  formatOptionalTime(row.JoinedAt),
		})
	}

	log.Info().Int("count", len(resp)).Msg("list my pods handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) JoinPod(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	membership, err := h.service.RequestJoin(c.Request().Context(), userPublicID, podSlug)
	if err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Msg("join pod handled")
	return c.JSON(http.StatusCreated, map[string]string{
		"pod_slug": podSlug,
		"status":   string(membership.Status),
	})
}

func (h *Handler) AcceptMember(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	ownerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	memberUsername := c.Param("username")
	if podSlug == "" || memberUsername == "" {
		return apperrors.ErrBadRequest
	}

	membership, err := h.service.AcceptMember(c.Request().Context(), ownerPublicID, podSlug, memberUsername)
	if err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Str("member", memberUsername).Msg("accept member handled")
	return c.JSON(http.StatusOK, map[string]string{
		"pod_slug": podSlug,
		"username": memberUsername,
		"status":   string(membership.Status),
	})
}

func (h *Handler) RejectMember(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	ownerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	memberUsername := c.Param("username")
	if podSlug == "" || memberUsername == "" {
		return apperrors.ErrBadRequest
	}

	membership, err := h.service.RejectMember(c.Request().Context(), ownerPublicID, podSlug, memberUsername)
	if err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Str("member", memberUsername).Msg("reject member handled")
	return c.JSON(http.StatusOK, map[string]string{
		"pod_slug": podSlug,
		"username": memberUsername,
		"status":   string(membership.Status),
	})
}

func (h *Handler) LeavePod(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	if err := h.service.LeavePod(c.Request().Context(), userPublicID, podSlug); err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Msg("leave pod handled")
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SetMemberRole(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	ownerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	memberUsername := c.Param("username")
	if podSlug == "" || memberUsername == "" {
		return apperrors.ErrBadRequest
	}

	var req setMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	membership, err := h.service.SetMemberRole(
		c.Request().Context(),
		ownerPublicID,
		podSlug,
		memberUsername,
		req.Role,
	)
	if err != nil {
		return err
	}

	log.Info().
		Str("pod_slug", podSlug).
		Str("member", memberUsername).
		Str("role", string(membership.Role)).
		Msg("set member role handled")

	return c.JSON(http.StatusOK, map[string]string{
		"pod_slug": podSlug,
		"username": memberUsername,
		"role":     string(membership.Role),
		"status":   string(membership.Status),
	})
}

func (h *Handler) DeletePod(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	ownerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	if err := h.service.DeletePod(c.Request().Context(), ownerPublicID, podSlug); err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Msg("delete pod handled")
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RemoveMember(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	ownerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	memberUsername := c.Param("username")
	if podSlug == "" || memberUsername == "" {
		return apperrors.ErrBadRequest
	}

	if err := h.service.RemoveMember(c.Request().Context(), ownerPublicID, podSlug, memberUsername); err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Str("member", memberUsername).Msg("remove member handled")
	return c.NoContent(http.StatusNoContent)
}

func scopedUserPublicID(c echo.Context, service *Service) (uuid.UUID, error) {
	return auth.ResolveScopedUserPublicID(
		c,
		c.Param("username"),
		service.GetUserPublicIDByUsername,
	)
}

func toPodResponse(row sqlc.GetPodBySlugRow, acceptedCount int64) podResponse {
	resp := podResponse{
		ID:            row.ID,
		Slug:          row.Slug,
		Name:          row.Name,
		SkillSlug:     row.SkillSlug,
		SkillName:     row.SkillName,
		OwnerUsername: row.OwnerUsername,
		MaxMembers:    row.MaxMembers,
		AcceptedCount: acceptedCount,
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	return resp
}

func toListPodResponse(row sqlc.ListPodsBySkillSlugRow) podResponse {
	resp := podResponse{
		ID:            row.ID,
		Slug:          row.Slug,
		Name:          row.Name,
		SkillSlug:     row.SkillSlug,
		SkillName:     row.SkillName,
		OwnerUsername: row.OwnerUsername,
		MaxMembers:    row.MaxMembers,
		AcceptedCount: row.AcceptedCount,
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	return resp
}

func formatOptionalTime(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.UTC().Format(time.RFC3339)
	return &s
}
