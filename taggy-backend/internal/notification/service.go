package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

const (
	defaultLimit int32 = 50
	maxLimit     int32 = 100
)

type Service struct {
	repo *Repository
	log  zerolog.Logger
}

func NewService(repo *Repository, log zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	id, err := s.repo.GetUserPublicIDByUsername(ctx, username)
	if err != nil {
		return uuid.Nil, logging.Unexpected(s.log, err, "get user public id by username failed")
	}
	return id, nil
}

func (s *Service) List(
	ctx context.Context,
	userPublicID uuid.UUID,
	unreadOnly bool,
	limit int32,
) ([]sqlc.Notification, int64, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, logging.Reject(s.log, apperrors.ErrNotFound, "list notifications rejected: user not found")
		}
		return nil, 0, logging.Unexpected(s.log, err, "list notifications user lookup failed")
	}

	rows, err := s.repo.ListNotificationsByUserID(ctx, user.ID, unreadOnly, normalizeLimit(limit))
	if err != nil {
		return nil, 0, logging.Unexpected(s.log, err, "list notifications failed")
	}
	unread, err := s.repo.CountUnreadNotifications(ctx, user.ID)
	if err != nil {
		return nil, 0, logging.Unexpected(s.log, err, "count unread notifications failed")
	}
	return rows, unread, nil
}

func (s *Service) MarkRead(ctx context.Context, userPublicID uuid.UUID, notificationID int64) (sqlc.Notification, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Notification{}, logging.Reject(s.log, apperrors.ErrNotFound, "mark notification read rejected: user not found")
		}
		return sqlc.Notification{}, logging.Unexpected(s.log, err, "mark notification read user lookup failed")
	}

	existing, err := s.repo.GetNotificationByIDAndUserID(ctx, notificationID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Notification{}, logging.Reject(s.log, ErrNotificationNotFound, "mark notification read rejected: not found")
		}
		return sqlc.Notification{}, logging.Unexpected(s.log, err, "mark notification read lookup failed")
	}
	if existing.IsRead {
		return existing, logging.Reject(s.log, ErrAlreadyRead, "mark notification read rejected: already read")
	}

	updated, err := s.repo.MarkNotificationRead(ctx, notificationID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Notification{}, logging.Reject(s.log, ErrAlreadyRead, "mark notification read rejected: already read")
		}
		return sqlc.Notification{}, logging.Unexpected(s.log, err, "mark notification read failed")
	}

	s.log.Info().
		Int64("notification_id", notificationID).
		Str("user", user.Username).
		Msg("notification marked read")

	return updated, nil
}

func (s *Service) MarkAllRead(ctx context.Context, userPublicID uuid.UUID) (int64, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, logging.Reject(s.log, apperrors.ErrNotFound, "mark all notifications read rejected: user not found")
		}
		return 0, logging.Unexpected(s.log, err, "mark all notifications read user lookup failed")
	}
	n, err := s.repo.MarkAllNotificationsRead(ctx, user.ID)
	if err != nil {
		return 0, logging.Unexpected(s.log, err, "mark all notifications read failed")
	}

	s.log.Info().
		Int64("count", n).
		Str("user", user.Username).
		Msg("all notifications marked read")

	return n, nil
}

func (s *Service) ClearRead(ctx context.Context, userPublicID uuid.UUID) (int64, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, logging.Reject(s.log, apperrors.ErrNotFound, "clear read notifications rejected: user not found")
		}
		return 0, logging.Unexpected(s.log, err, "clear read notifications user lookup failed")
	}

	n, err := s.repo.DeleteReadNotificationsByUserID(ctx, user.ID)
	if err != nil {
		return 0, logging.Unexpected(s.log, err, "clear read notifications failed")
	}

	s.log.Info().
		Int64("deleted", n).
		Str("user", user.Username).
		Msg("read notifications cleared")

	return n, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) error {
	typ, err := parseType(input.Type)
	if err != nil {
		return logging.Reject(s.log, err, "create notification rejected: invalid type")
	}
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	if title == "" || body == "" {
		return logging.Reject(s.log, apperrors.ErrBadRequest, "create notification rejected: empty title or body")
	}

	params := sqlc.CreateNotificationParams{
		UserID: input.UserID,
		Type:   typ,
		Title:  title,
		Body:   body,
	}
	if input.EntityType != nil && strings.TrimSpace(*input.EntityType) != "" {
		params.EntityType = pgtype.Text{String: strings.TrimSpace(*input.EntityType), Valid: true}
	}
	if input.EntityID != nil {
		params.EntityID = pgtype.Int8{Int64: *input.EntityID, Valid: true}
	}

	created, err := s.repo.CreateNotification(ctx, params)
	if err != nil {
		return logging.Unexpected(s.log, err, "create notification failed")
	}

	s.log.Info().
		Int64("notification_id", created.ID).
		Int64("user_id", input.UserID).
		Str("type", string(typ)).
		Msg("notification created")

	return nil
}

// NotifyPodJoinRequest notifies the pod owner (best-effort).
func (s *Service) NotifyPodJoinRequest(ctx context.Context, ownerUserID int64, podID int64, podSlug, requesterUsername string) {
	s.emit(ctx, CreateInput{
		UserID:     ownerUserID,
		Type:       string(sqlc.NotificationTypePODJOINREQUEST),
		EntityType: strPtr("pod"),
		EntityID:   &podID,
		Title:      "Pod join request",
		Body:       fmt.Sprintf("%s requested to join %s", requesterUsername, podSlug),
	})
}

func (s *Service) NotifyPodJoinAccepted(ctx context.Context, memberUserID int64, podID int64, podSlug string) {
	s.emit(ctx, CreateInput{
		UserID:     memberUserID,
		Type:       string(sqlc.NotificationTypePODJOINACCEPTED),
		EntityType: strPtr("pod"),
		EntityID:   &podID,
		Title:      "Pod join accepted",
		Body:       fmt.Sprintf("Your request to join %s was accepted", podSlug),
	})
}

func (s *Service) NotifyPodJoinRejected(ctx context.Context, memberUserID int64, podID int64, podSlug string) {
	s.emit(ctx, CreateInput{
		UserID:     memberUserID,
		Type:       string(sqlc.NotificationTypePODJOINREJECTED),
		EntityType: strPtr("pod"),
		EntityID:   &podID,
		Title:      "Pod join rejected",
		Body:       fmt.Sprintf("Your request to join %s was rejected", podSlug),
	})
}

func (s *Service) NotifyPodMemberRemoved(ctx context.Context, memberUserID int64, podID int64, podSlug string) {
	s.emit(ctx, CreateInput{
		UserID:     memberUserID,
		Type:       string(sqlc.NotificationTypePODMEMBERREMOVED),
		EntityType: strPtr("pod"),
		EntityID:   &podID,
		Title:      "Removed from pod",
		Body:       fmt.Sprintf("You were removed from %s", podSlug),
	})
}

func (s *Service) NotifyMilestoneCompleted(ctx context.Context, userID, milestoneID int64, skillSlug, milestoneTitle string) {
	s.emit(ctx, CreateInput{
		UserID:     userID,
		Type:       string(sqlc.NotificationTypeMILESTONECOMPLETED),
		EntityType: strPtr("skill"),
		EntityID:   &milestoneID,
		Title:      "Milestone completed",
		Body:       fmt.Sprintf("You completed \"%s\" on %s", milestoneTitle, skillSlug),
	})
}

func (s *Service) NotifyMilestoneDue(ctx context.Context, userID, milestoneID int64, skillSlug, milestoneTitle, dueAtRFC3339 string) {
	s.emit(ctx, CreateInput{
		UserID:     userID,
		Type:       string(sqlc.NotificationTypeMILESTONEDUE),
		EntityType: strPtr("skill"),
		EntityID:   &milestoneID,
		Title:      "Milestone deadline",
		Body:       fmt.Sprintf("\"%s\" on %s is due by %s", milestoneTitle, skillSlug, dueAtRFC3339),
	})
}

func (s *Service) NotifyRoadmapUpdated(ctx context.Context, userID, skillID int64, skillSlug string, versionNumber int32) {
	s.emit(ctx, CreateInput{
		UserID:     userID,
		Type:       string(sqlc.NotificationTypeROADMAPUPDATED),
		EntityType: strPtr("skill"),
		EntityID:   &skillID,
		Title:      "Roadmap updated",
		Body:       fmt.Sprintf("Your %s roadmap is now version %d", skillSlug, versionNumber),
	})
}

func (s *Service) NotifyCommunityAnnouncement(ctx context.Context, userID, communityID int64, skillSlug, body string) {
	s.emit(ctx, CreateInput{
		UserID:     userID,
		Type:       string(sqlc.NotificationTypeCOMMUNITYANNOUNCEMENT),
		EntityType: strPtr("community"),
		EntityID:   &communityID,
		Title:      "Community announcement",
		Body:       body,
	})
}

func (s *Service) NotifySkillRequestApproved(ctx context.Context, userID, skillID int64, skillSlug string) {
	s.emit(ctx, CreateInput{
		UserID:     userID,
		Type:       string(sqlc.NotificationTypeSKILLREQUESTAPPROVED),
		EntityType: strPtr("skill"),
		EntityID:   &skillID,
		Title:      "Skill request approved",
		Body:       fmt.Sprintf("Your skill request was approved. You can join %s.", skillSlug),
	})
}

func (s *Service) NotifySkillRequestRejected(ctx context.Context, userID int64, skillName, note string) {
	body := fmt.Sprintf("Your request to add \"%s\" was rejected.", skillName)
	if strings.TrimSpace(note) != "" {
		body += " Note: " + strings.TrimSpace(note)
	}
	s.emit(ctx, CreateInput{
		UserID: userID,
		Type:   string(sqlc.NotificationTypeSKILLREQUESTREJECTED),
		Title:  "Skill request rejected",
		Body:   body,
	})
}

func (s *Service) NotifyRoadmapRequestApproved(ctx context.Context, userID, skillID int64, skillSlug string, versionNumber int32) {
	s.emit(ctx, CreateInput{
		UserID:     userID,
		Type:       string(sqlc.NotificationTypeROADMAPREQUESTAPPROVED),
		EntityType: strPtr("skill"),
		EntityID:   &skillID,
		Title:      "Roadmap update approved",
		Body:       fmt.Sprintf("Your roadmap edit for %s was published as version %d.", skillSlug, versionNumber),
	})
}

func (s *Service) NotifyRoadmapRequestRejected(ctx context.Context, userID int64, skillSlug, note string) {
	body := fmt.Sprintf("Your roadmap edit request for %s was rejected.", skillSlug)
	if strings.TrimSpace(note) != "" {
		body += " Note: " + strings.TrimSpace(note)
	}
	s.emit(ctx, CreateInput{
		UserID: userID,
		Type:   string(sqlc.NotificationTypeROADMAPREQUESTREJECTED),
		Title:  "Roadmap update rejected",
		Body:   body,
	})
}

func (s *Service) emit(ctx context.Context, input CreateInput) {
	_ = s.Create(ctx, input)
}

func parseType(raw string) (sqlc.NotificationType, error) {
	t := sqlc.NotificationType(strings.ToUpper(strings.TrimSpace(raw)))
	switch t {
	case sqlc.NotificationTypePODJOINREQUEST,
		sqlc.NotificationTypePODJOINACCEPTED,
		sqlc.NotificationTypePODJOINREJECTED,
		sqlc.NotificationTypePODMEMBERREMOVED,
		sqlc.NotificationTypeMILESTONEDUE,
		sqlc.NotificationTypeMILESTONECOMPLETED,
		sqlc.NotificationTypeROADMAPUPDATED,
		sqlc.NotificationTypePROPOSALAPPROVED,
		sqlc.NotificationTypePROPOSALREJECTED,
		sqlc.NotificationTypeCOMMUNITYANNOUNCEMENT,
		sqlc.NotificationTypeSYSTEM,
		sqlc.NotificationTypeSKILLREQUESTAPPROVED,
		sqlc.NotificationTypeSKILLREQUESTREJECTED,
		sqlc.NotificationTypeROADMAPREQUESTAPPROVED,
		sqlc.NotificationTypeROADMAPREQUESTREJECTED:
		return t, nil
	default:
		return "", apperrors.ErrBadRequest
	}
}

func normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func strPtr(s string) *string { return &s }
