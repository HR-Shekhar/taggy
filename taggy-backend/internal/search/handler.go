package search

import (
	"net/http"
	"strconv"
	"strings"

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

func (h *Handler) Search(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	q := c.QueryParam("q")
	if q == "" {
		q = c.QueryParam("query")
	}

	var types []string
	if raw := c.QueryParam("types"); raw != "" {
		types = strings.Split(raw, ",")
	} else if raw := c.QueryParam("type"); raw != "" {
		types = []string{raw}
	}

	var limit int32
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 0 {
			return apperrors.ErrBadRequest
		}
		limit = int32(n)
	}

	result, err := h.service.Search(c.Request().Context(), Input{
		Query: q,
		Types: types,
		Limit: limit,
	})
	if err != nil {
		return err
	}

	resp := searchResponse{Query: strings.TrimSpace(q)}
	if result.Skills != nil {
		resp.Skills = make([]skillHitResponse, 0, len(result.Skills))
		for _, hit := range result.Skills {
			resp.Skills = append(resp.Skills, skillHitResponse{
				ID:          hit.ID,
				Name:        hit.Name,
				Slug:        hit.Slug,
				Description: hit.Description,
			})
		}
	}
	if result.Users != nil {
		resp.Users = make([]userHitResponse, 0, len(result.Users))
		for _, hit := range result.Users {
			resp.Users = append(resp.Users, userHitResponse{
				PublicID:          hit.PublicID,
				Username:          hit.Username,
				Name:              hit.Name,
				ProfilePictureURL: hit.ProfilePictureURL,
				Bio:               hit.Bio,
			})
		}
	}
	if result.Communities != nil {
		resp.Communities = make([]communityHitResponse, 0, len(result.Communities))
		for _, hit := range result.Communities {
			resp.Communities = append(resp.Communities, communityHitResponse{
				ID:          hit.ID,
				Name:        hit.Name,
				Description: hit.Description,
				SkillSlug:   hit.SkillSlug,
				SkillName:   hit.SkillName,
			})
		}
	}

	log.Info().Str("query", resp.Query).Msg("search handled")
	return c.JSON(http.StatusOK, resp)
}
