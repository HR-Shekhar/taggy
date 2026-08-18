package http

import (
	"errors"
	"net/http"

	"github.com/HR-Shekhar/taggy-backend/internal/audio"
	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	"github.com/HR-Shekhar/taggy-backend/internal/billing"
	"github.com/HR-Shekhar/taggy-backend/internal/community"
	"github.com/HR-Shekhar/taggy-backend/internal/notification"
	"github.com/HR-Shekhar/taggy-backend/internal/pod"
	"github.com/HR-Shekhar/taggy-backend/internal/progress"
	"github.com/HR-Shekhar/taggy-backend/internal/quiz"
	"github.com/HR-Shekhar/taggy-backend/internal/report"
	"github.com/HR-Shekhar/taggy-backend/internal/roadmap"
	"github.com/HR-Shekhar/taggy-backend/internal/roadmaprequest"
	"github.com/HR-Shekhar/taggy-backend/internal/search"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/skill"
	"github.com/HR-Shekhar/taggy-backend/internal/skillrequest"
	"github.com/HR-Shekhar/taggy-backend/internal/user"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// ErrorResponse is the JSON structure returned to API clients whenever
// an error occurs.
type ErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func ErrorHandler(log zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var (
			status  int
			message string
			fields  map[string]string
		)

		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			status = httpErr.Code

			if msg, ok := httpErr.Message.(string); ok {
				message = msg
			} else {
				message = http.StatusText(status)
			}

		} else {
			var validationErr apperrors.ValidationError
			switch {
			case errors.As(err, &validationErr):
				status = http.StatusUnprocessableEntity
				message = validationErr.Error()
				fields = validationErr.Fields

			case errors.Is(err, apperrors.ErrBadRequest):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, apperrors.ErrTooManyRequests):
				status = http.StatusTooManyRequests
				message = err.Error()

			case errors.Is(err, auth.ErrInvalidCredentials),
				errors.Is(err, auth.ErrInvalidRefreshToken),
				errors.Is(err, auth.ErrSessionExpired),
				errors.Is(err, auth.ErrInvalidOTP),
				errors.Is(err, auth.ErrInvalidOAuthState),
				errors.Is(err, apperrors.ErrUnauthorized):
				status = http.StatusUnauthorized
				message = err.Error()

			case errors.Is(err, auth.ErrEmailNotVerified):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, auth.ErrOTPExpired):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, auth.ErrEmailAlreadyVerified),
				errors.Is(err, auth.ErrEmailInUse),
				errors.Is(err, auth.ErrUsernameInUse):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, auth.ErrOAuthNotConfigured),
				errors.Is(err, auth.ErrOAuthAccountInvalid),
				errors.Is(err, auth.ErrOAuthUsernameRequired),
				errors.Is(err, auth.ErrInvalidUsername),
				errors.Is(err, auth.ErrInvalidRegistrationToken):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, auth.ErrEmailDeliveryFailed):
				status = http.StatusBadGateway
				message = err.Error()

			case errors.Is(err, user.ErrUsernameTaken),
				errors.Is(err, user.ErrInvalidUsername):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, user.ErrInvalidAvatar):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, user.ErrAvatarStorageUnavailable):
				status = http.StatusServiceUnavailable
				message = err.Error()

			case errors.Is(err, user.ErrAvatarUploadFailed):
				status = http.StatusBadGateway
				message = err.Error()

			case errors.Is(err, skill.ErrSkillNotFound),
				errors.Is(err, skill.ErrCommunityNotFound),
				errors.Is(err, skill.ErrRoadmapNotFound),
				errors.Is(err, skill.ErrUserSkillNotFound),
				errors.Is(err, skill.ErrMilestoneNotFound),
				errors.Is(err, skill.ErrVersionNotFound),
				errors.Is(err, roadmap.ErrRoadmapNotFound),
				errors.Is(err, roadmap.ErrVersionNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, skill.ErrAlreadyEnrolled),
				errors.Is(err, skill.ErrActiveSkillLimit),
				errors.Is(err, skill.ErrMilestoneOutOfOrder),
				errors.Is(err, skill.ErrSubtopicsIncomplete),
				errors.Is(err, skill.ErrMilestoneAlreadyComplete),
				errors.Is(err, skill.ErrAlreadyOnVersion):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, skill.ErrInvalidMilestoneAction),
				errors.Is(err, skill.ErrVersionNotSelectable),
				errors.Is(err, roadmap.ErrVersionNotSelectable):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, progress.ErrNotEnrolledInSkill):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, pod.ErrPodNotFound),
				errors.Is(err, pod.ErrMembershipNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, pod.ErrNotEnrolledInSkill),
				errors.Is(err, pod.ErrNotPodOwner),
				errors.Is(err, pod.ErrCannotRemoveOwner),
				errors.Is(err, pod.ErrCannotChangeOwnRole),
				errors.Is(err, pod.ErrNotAcceptedMember):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, pod.ErrAlreadyInActivePod),
				errors.Is(err, pod.ErrAlreadyMember),
				errors.Is(err, pod.ErrAlreadyPending),
				errors.Is(err, pod.ErrPodFull),
				errors.Is(err, pod.ErrMembershipNotPending),
				errors.Is(err, pod.ErrPodSlugTaken),
				errors.Is(err, pod.ErrPodNotEmpty):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, pod.ErrInvalidPodName),
				errors.Is(err, pod.ErrInvalidPodSlug),
				errors.Is(err, pod.ErrInvalidMemberRole):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, community.ErrCommunityNotFound),
				errors.Is(err, community.ErrChannelNotFound),
				errors.Is(err, community.ErrMessageNotFound),
				errors.Is(err, community.ErrPodNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, community.ErrNotEnrolledInSkill),
				errors.Is(err, community.ErrNotAcceptedPodMember),
				errors.Is(err, community.ErrNotMessageAuthor):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, community.ErrInvalidMessageContent),
				errors.Is(err, community.ErrInvalidReplyTarget),
				errors.Is(err, community.ErrInvalidChatRoom):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, audio.ErrRoomNotFound),
				errors.Is(err, audio.ErrPodNotFound),
				errors.Is(err, audio.ErrChannelNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, audio.ErrNotEnrolledInSkill),
				errors.Is(err, audio.ErrNotAcceptedPodMember),
				errors.Is(err, audio.ErrNotRoomHost),
				errors.Is(err, audio.ErrNotParticipant):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, audio.ErrActiveRoomExists),
				errors.Is(err, audio.ErrRoomFull),
				errors.Is(err, audio.ErrRoomNotActive):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, audio.ErrInvalidRoomTitle):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, audio.ErrLiveKitNotConfigured):
				status = http.StatusServiceUnavailable
				message = err.Error()

			case errors.Is(err, notification.ErrNotificationNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, notification.ErrAlreadyRead):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, report.ErrReportNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, report.ErrDuplicateOpen):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, report.ErrInvalidTarget),
				errors.Is(err, report.ErrInvalidReason),
				errors.Is(err, report.ErrCannotReportSelf):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, search.ErrInvalidQuery),
				errors.Is(err, search.ErrInvalidType):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, skillrequest.ErrInvalidName),
				errors.Is(err, skillrequest.ErrInvalidDescription):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, skillrequest.ErrSimilarFound):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, skillrequest.ErrDuplicatePending),
				errors.Is(err, roadmaprequest.ErrDuplicatePending),
				errors.Is(err, skillrequest.ErrSlugTaken):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, skillrequest.ErrRequestNotFound),
				errors.Is(err, roadmaprequest.ErrRequestNotFound),
				errors.Is(err, roadmaprequest.ErrSkillNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, skillrequest.ErrNotPending),
				errors.Is(err, roadmaprequest.ErrNotPending),
				errors.Is(err, roadmaprequest.ErrNoActiveVersion):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, skillrequest.ErrAIUnavailable),
				errors.Is(err, roadmaprequest.ErrAIUnavailable),
				errors.Is(err, apperrors.ErrServiceUnavailable):
				status = http.StatusServiceUnavailable
				message = err.Error()

			case errors.Is(err, skillrequest.ErrAIFailed),
				errors.Is(err, roadmaprequest.ErrAIFailed):
				status = http.StatusBadGateway
				message = err.Error()

			case errors.Is(err, skillrequest.ErrNotAdmin),
				errors.Is(err, roadmaprequest.ErrNotAdmin),
				errors.Is(err, roadmaprequest.ErrNotEnrolled):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, billing.ErrBillingNotConfigured):
				status = http.StatusServiceUnavailable
				message = err.Error()

			case errors.Is(err, billing.ErrAlreadyPremium):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, billing.ErrInvalidSignature),
				errors.Is(err, billing.ErrPaymentNotPayable):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, billing.ErrPaymentNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, billing.ErrPaymentForbidden):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, billing.ErrOrderCreateFailed):
				status = http.StatusBadGateway
				message = err.Error()

			case errors.Is(err, quiz.ErrPodNotFound),
				errors.Is(err, quiz.ErrSkillNotFound),
				errors.Is(err, quiz.ErrQuizNotFound),
				errors.Is(err, quiz.ErrQuestionNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, quiz.ErrNotAcceptedMember),
				errors.Is(err, quiz.ErrQuizNotOwned):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, quiz.ErrNoCompletedTopics),
				errors.Is(err, quiz.ErrQuizNotInProgress),
				errors.Is(err, quiz.ErrAlreadyAnswered),
				errors.Is(err, quiz.ErrAnswerNotStarted),
				errors.Is(err, quiz.ErrInProgressExists),
				errors.Is(err, quiz.ErrIncompleteAnswers):
				status = http.StatusConflict
				message = err.Error()

			case errors.Is(err, quiz.ErrAIUnavailable):
				status = http.StatusServiceUnavailable
				message = err.Error()

			case errors.Is(err, quiz.ErrAIFailed):
				status = http.StatusBadGateway
				message = err.Error()

			case errors.Is(err, apperrors.ErrForbidden):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, apperrors.ErrNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, apperrors.ErrConflict):
				status = http.StatusConflict
				message = err.Error()

			default:
				status = http.StatusInternalServerError
				message = "internal server error"
			}
		}

		event := log.Warn()
		if status >= 500 {
			event = log.Error()
		}

		if status >= 400 {
			uri := c.Request().URL.RequestURI()
			routePath := c.Path()
			if routePath == "" {
				routePath = c.Request().URL.Path
			}
			event.
				Err(err).
				Int("status", status).
				Str("method", c.Request().Method).
				Str("path", routePath).
				Str("uri", uri).
				Msg("request failed")
		}

		_ = c.JSON(status, ErrorResponse{
			Message: message,
			Fields:  fields,
		})
	}
}
