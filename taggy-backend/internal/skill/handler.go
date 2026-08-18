package skill

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

func (h *Handler) ListSkills(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	skills, err := h.service.ListSkills(c.Request().Context())
	if err != nil {
		return err
	}

	resp := make([]skillResponse, 0, len(skills))
	for _, s := range skills {
		resp = append(resp, toSkillResponse(s))
	}

	log.Info().Int("count", len(resp)).Msg("skills list handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetSkillBySlug(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	slug := c.Param("slug")

	skill, community, err := h.service.GetSkillBySlug(c.Request().Context(), slug)
	if err != nil {
		return err
	}

	log.Info().Str("slug", slug).Msg("skill detail handled")
	return c.JSON(http.StatusOK, skillDetailResponse{
		Skill:     toSkillResponse(skill),
		Community: toCommunityResponse(community),
	})
}

func (h *Handler) JoinSkill(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	skillSlug := c.Param("slug")
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}

	userSkill, community, skill, err := h.service.JoinSkill(c.Request().Context(), userPublicID, skillSlug)
	if err != nil {
		return err
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Str("skill_slug", skillSlug).
		Msg("join skill handled")

	return c.JSON(http.StatusCreated, joinSkillResponse{
		UserSkill: toUserSkillResponse(userSkill, skill.Name, skill.Slug),
		Community: toCommunityResponse(community),
	})
}

func (h *Handler) ListMySkills(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	rows, err := h.service.ListMySkills(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}

	resp := make([]userSkillResponse, 0, len(rows))
	for _, row := range rows {
		percent := 0.0
		if row.MilestoneCount > 0 {
			percent = float64(row.CompletedCount) / float64(row.MilestoneCount) * 100
		}
		resp = append(resp, userSkillResponse{
			SkillSlug:            row.SkillSlug,
			SkillName:            row.SkillName,
			Status:               string(row.Status),
			StartedAt:            formatTime(row.StartedAt),
			RoadmapVersionNumber: row.RoadmapVersionNumber,
			RoadmapVersionStatus: string(row.RoadmapVersionStatus),
			MilestoneCount:       row.MilestoneCount,
			CompletedCount:       row.CompletedCount,
			CompletionPercent:    percent,
		})
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Int("count", len(resp)).
		Msg("my skills handled")

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) SwitchRoadmapVersion(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	var req switchRoadmapVersionRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userSkill, version, err := h.service.SwitchRoadmapVersion(
		c.Request().Context(),
		userPublicID,
		skillSlug,
		req.VersionNumber,
	)
	if err != nil {
		return err
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Str("skill_slug", skillSlug).
		Int32("version", req.VersionNumber).
		Msg("switch roadmap version handled")

	return c.JSON(http.StatusOK, userSkillResponse{
		SkillSlug:            version.SkillSlug,
		SkillName:            version.SkillName,
		Status:               string(userSkill.Status),
		StartedAt:            formatTime(userSkill.StartedAt),
		RoadmapVersionNumber: version.VersionNumber,
		RoadmapVersionStatus: string(version.Status),
	})
}

func (h *Handler) ListMilestones(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	rows, err := h.service.ListMilestones(c.Request().Context(), userPublicID, skillSlug)
	if err != nil {
		return err
	}

	resp := make([]milestoneResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toMilestoneProgressResponse(row))
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Str("skill_slug", skillSlug).
		Msg("milestones list handled")

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateMilestone(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	milestoneSlug := c.Param("milestoneSlug")

	var req updateMilestoneRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid milestone update payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	input := UpdateMilestoneInput{
		Action: MilestoneAction(req.Action),
	}

	if req.Action == string(MilestoneActionPostpone) && req.PostponedUntil != nil {
		t, err := time.Parse(time.RFC3339, *req.PostponedUntil)
		if err != nil {
			return apperrors.ErrBadRequest
		}
		input.PostponedUntil = &t
	}

	progress, err := h.service.UpdateMilestone(
		c.Request().Context(),
		userPublicID,
		skillSlug,
		milestoneSlug,
		input,
	)
	if err != nil {
		return err
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Str("milestone_slug", milestoneSlug).
		Msg("milestone update handled")

	return c.JSON(http.StatusOK, milestoneProgressToResponse(progress))
}

func scopedUserPublicID(c echo.Context, service *Service) (uuid.UUID, error) {
	return auth.ResolveScopedUserPublicID(
		c,
		c.Param("username"),
		service.GetUserPublicIDByUsername,
	)
}

func toSkillResponse(s sqlc.Skill) skillResponse {
	resp := skillResponse{
		Name: s.Name,
		Slug: s.Slug,
	}
	if s.Description.Valid {
		resp.Description = &s.Description.String
	}
	return resp
}

func toCommunityResponse(c sqlc.GetCommunityBySkillIDRow) communityResponse {
	resp := communityResponse{
		Slug: c.SkillSlug,
		Name: c.Name,
	}
	if c.Description.Valid {
		resp.Description = &c.Description.String
	}
	return resp
}

func toUserSkillResponse(us sqlc.Userskill, name, slug string) userSkillResponse {
	return userSkillResponse{
		SkillSlug: slug,
		SkillName: name,
		Status:    string(us.Status),
		StartedAt: formatTime(us.StartedAt),
	}
}

func toMilestoneProgressResponse(row sqlc.ListMilestoneProgressByUserSkillIDRow) milestoneResponse {
	resp := milestoneResponse{
		Slug:       row.Slug,
		Title:      row.Title,
		OrderIndex: row.OrderIndex,
		Kind:       row.Kind,
		Status:     string(row.Status),
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	if row.EstimatedHours.Valid {
		resp.EstimatedHours = &row.EstimatedHours.Int32
	}
	if row.Difficulty.Valid {
		resp.Difficulty = &row.Difficulty.String
	}
	if row.Chapter.Valid && row.Chapter.String != "" {
		resp.Chapter = &row.Chapter.String
	}
	if resp.Kind == "" {
		resp.Kind = "TOPIC"
	}
	if row.CompletedAt.Valid {
		t := formatTime(row.CompletedAt)
		resp.CompletedAt = &t
	}
	if row.PostponedUntil.Valid {
		t := formatTime(row.PostponedUntil)
		resp.PostponedUntil = &t
	}
	return resp
}

func milestoneProgressToResponse(p sqlc.GetMilestoneProgressBySlugRow) milestoneResponse {
	resp := milestoneResponse{
		Slug:       p.Slug,
		Title:      p.Title,
		OrderIndex: p.OrderIndex,
		Status:     string(p.Status),
	}
	if p.Description.Valid {
		resp.Description = &p.Description.String
	}
	if p.EstimatedHours.Valid {
		resp.EstimatedHours = &p.EstimatedHours.Int32
	}
	if p.Difficulty.Valid {
		resp.Difficulty = &p.Difficulty.String
	}
	if p.CompletedAt.Valid {
		t := formatTime(p.CompletedAt)
		resp.CompletedAt = &t
	}
	if p.PostponedUntil.Valid {
		t := formatTime(p.PostponedUntil)
		resp.PostponedUntil = &t
	}
	return resp
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}
