package roadmap

import (
	"net/http"
	"strconv"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
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

func (h *Handler) GetRoadmap(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	skillSlug := c.Param("slug")
	if skillSlug == "" {
		skillSlug = c.Param("skillSlug")
	}
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}

	overview, err := h.service.GetRoadmapOverview(c.Request().Context(), skillSlug)
	if err != nil {
		return err
	}

	versions := make([]versionSummaryResponse, 0, len(overview.Versions))
	var current *versionSummaryResponse
	for _, v := range overview.Versions {
		item := toVersionSummary(v)
		versions = append(versions, item)
		if v.IsCurrent {
			copy := item
			current = &copy
		}
	}
	if current == nil {
		for i := range versions {
			if versions[i].Status == string(sqlc.CurrentStatusACTIVE) {
				current = &versions[i]
				break
			}
		}
	}

	log.Info().Str("skill_slug", skillSlug).Int("versions", len(versions)).Msg("roadmap overview handled")
	return c.JSON(http.StatusOK, roadmapSummaryResponse{
		SkillSlug:      overview.SkillSlug,
		SkillName:      overview.SkillName,
		CurrentVersion: current,
		Versions:       versions,
	})
}

func (h *Handler) ListVersions(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	skillSlug := c.Param("slug")
	if skillSlug == "" {
		skillSlug = c.Param("skillSlug")
	}

	versions, err := h.service.ListVersions(c.Request().Context(), skillSlug)
	if err != nil {
		return err
	}

	resp := make([]versionSummaryResponse, 0, len(versions))
	for _, v := range versions {
		resp = append(resp, toVersionSummary(v))
	}

	log.Info().Str("skill_slug", skillSlug).Int("count", len(resp)).Msg("roadmap versions list handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetVersion(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	skillSlug := c.Param("slug")
	if skillSlug == "" {
		skillSlug = c.Param("skillSlug")
	}
	versionRaw := c.Param("versionNumber")
	versionNumber, err := strconv.ParseInt(versionRaw, 10, 32)
	if err != nil || versionNumber < 1 {
		return apperrors.ErrBadRequest
	}

	detail, err := h.service.GetVersionDetail(c.Request().Context(), skillSlug, int32(versionNumber))
	if err != nil {
		return err
	}

	milestones := make([]milestoneResponse, 0, len(detail.Milestones))
	for _, m := range detail.Milestones {
		milestones = append(milestones, toMilestone(m))
	}

	v := detail.Version
	resp := versionDetailResponse{
		SkillSlug:     v.SkillSlug,
		SkillName:     v.SkillName,
		VersionNumber: v.VersionNumber,
		Status:        string(v.Status),
		GeneratedBy:   v.GeneratedBy,
		IsCurrent:     v.IsCurrent,
		PublishedAt:   optionalTime(v.PublishedAt),
		CreatedAt:     formatTime(v.CreatedAt),
		Milestones:    milestones,
	}

	log.Info().
		Str("skill_slug", skillSlug).
		Int32("version", v.VersionNumber).
		Msg("roadmap version detail handled")
	return c.JSON(http.StatusOK, resp)
}

func toVersionSummary(v sqlc.ListRoadmapVersionsBySkillSlugRow) versionSummaryResponse {
	return versionSummaryResponse{
		VersionNumber:  v.VersionNumber,
		Status:         string(v.Status),
		GeneratedBy:    v.GeneratedBy,
		IsCurrent:      v.IsCurrent,
		MilestoneCount: v.MilestoneCount,
		PublishedAt:    optionalTime(v.PublishedAt),
		CreatedAt:      formatTime(v.CreatedAt),
	}
}

func toMilestone(m sqlc.ListMilestonesByRoadmapVersionIDRow) milestoneResponse {
	resp := milestoneResponse{
		Slug:       m.Slug,
		Title:      m.Title,
		OrderIndex: m.OrderIndex,
		Kind:       m.Kind,
	}
	if m.Description.Valid {
		resp.Description = &m.Description.String
	}
	if m.EstimatedHours.Valid {
		resp.EstimatedHours = &m.EstimatedHours.Int32
	}
	if m.Difficulty.Valid {
		resp.Difficulty = &m.Difficulty.String
	}
	if m.Chapter.Valid && m.Chapter.String != "" {
		resp.Chapter = &m.Chapter.String
	}
	if resp.Kind == "" {
		resp.Kind = "TOPIC"
	}
	return resp
}