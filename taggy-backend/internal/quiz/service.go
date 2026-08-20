package quiz

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/aigen"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/openrouter"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

const questionSeconds = 60

type AIClient interface {
	Available() bool
	GenerateQuiz(ctx context.Context, topicTitles []string) ([]openrouter.QuizQuestionDraft, error)
}

type Notifier interface {
	NotifyQuizReady(ctx context.Context, userID, podID int64, podSlug string)
	NotifyQuizFailed(ctx context.Context, userID, podID int64, podSlug, note string)
}

type JobPool interface {
	Submit(job aigen.Job) error
}

type Service struct {
	repo     *Repository
	ai       AIClient
	notifier Notifier
	pool     JobPool
	log      zerolog.Logger
}

func NewService(repo *Repository, ai AIClient, notifier Notifier, pool JobPool, log zerolog.Logger) *Service {
	return &Service{repo: repo, ai: ai, notifier: notifier, pool: pool, log: log}
}

func (s *Service) StartQuiz(ctx context.Context, userPublicID uuid.UUID, podSlug string) (quizResponse, error) {
	user, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return quizResponse{}, err
	}

	if existing, err := s.repo.GetInProgressQuiz(ctx, user.ID, pod.ID); err == nil {
		if _, aerr := s.repo.AbandonQuiz(ctx, existing.ID); aerr != nil {
			return quizResponse{}, logging.Unexpected(s.log, aerr, "abandon prior quiz failed")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return quizResponse{}, logging.Unexpected(s.log, err, "get in-progress quiz failed")
	}

	topics, err := s.repo.ListCompletedTopicTitles(ctx, user.ID, pod.SkillID)
	if err != nil {
		return quizResponse{}, logging.Unexpected(s.log, err, "list completed topics failed")
	}
	if len(topics) == 0 {
		return quizResponse{}, logging.Reject(s.log, ErrNoCompletedTopics, "start quiz rejected: no topics")
	}

	if s.ai == nil || !s.ai.Available() {
		return quizResponse{}, logging.Reject(s.log, ErrAIUnavailable, "start quiz rejected: ai unavailable")
	}
	if s.pool == nil {
		return quizResponse{}, logging.Reject(s.log, ErrAIUnavailable, "start quiz rejected: ai pool unavailable")
	}

	quiz, err := s.repo.CreateQuiz(ctx, sqlc.CreatePodQuizParams{
		PodID:                pod.ID,
		UserID:               user.ID,
		SkillID:              pod.SkillID,
		Status:               sqlc.PodQuizStatusGENERATING,
		TopicCount:           int32(len(topics)),
		CompletedTopicTitles: mustJSON(topics),
	})
	if err != nil {
		return quizResponse{}, logging.Unexpected(s.log, err, "create quiz failed")
	}

	if err := s.enqueue(quiz.ID, podSlug); err != nil {
		if _, ferr := s.repo.FailGenerating(ctx, quiz.ID); ferr != nil {
			s.log.Error().Err(ferr).Int64("quiz_id", quiz.ID).Msg("fail quiz after queue full failed")
		}
		return quizResponse{}, logging.Reject(s.log, ErrAIBusy, "start quiz rejected: ai queue full")
	}

	s.log.Info().
		Str("quiz_id", quiz.PublicID.String()).
		Str("pod", podSlug).
		Msg("pod quiz accepted for async generation")

	return s.buildQuizResponse(ctx, quiz, false)
}

// RequeueGenerating re-enqueues GENERATING quizzes after restart.
func (s *Service) RequeueGenerating(ctx context.Context) {
	if s.pool == nil || s.ai == nil || !s.ai.Available() {
		return
	}
	ids, err := s.repo.ListGeneratingIDs(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("list generating quizzes failed")
		return
	}
	for _, id := range ids {
		quiz, err := s.repo.GetQuizByID(ctx, id)
		if err != nil {
			continue
		}
		podSlug := ""
		if pod, err := s.repo.GetPodByID(ctx, quiz.PodID); err == nil {
			podSlug = pod.Slug
		}
		if err := s.enqueue(id, podSlug); err != nil {
			s.log.Warn().Err(err).Int64("id", id).Msg("requeue quiz failed")
		}
	}
	if len(ids) > 0 {
		s.log.Info().Int("count", len(ids)).Msg("requeued generating quizzes")
	}
}

func (s *Service) enqueue(id int64, podSlug string) error {
	return s.pool.Submit(func(ctx context.Context) {
		s.generateQuiz(ctx, id, podSlug)
	})
}

func (s *Service) generateQuiz(ctx context.Context, id int64, podSlug string) {
	quiz, err := s.repo.GetQuizByID(ctx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.log.Error().Err(err).Int64("id", id).Msg("quiz generate: get failed")
		}
		return
	}
	if quiz.Status != sqlc.PodQuizStatusGENERATING {
		return
	}

	topics := decodeStringSlice(quiz.CompletedTopicTitles)
	drafts, err := s.ai.GenerateQuiz(ctx, topics)
	if err != nil {
		s.log.Warn().Err(err).Int64("id", id).Msg("quiz generation failed")
		bg := context.WithoutCancel(ctx)
		if _, ferr := s.repo.FailGenerating(bg, id); ferr != nil && !errors.Is(ferr, pgx.ErrNoRows) {
			s.log.Error().Err(ferr).Int64("id", id).Msg("mark quiz failed")
		}
		if s.notifier != nil {
			note := "Quiz generation failed; please try again"
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				note = "Quiz generation timed out; please try again"
			}
			s.notifier.NotifyQuizFailed(bg, quiz.UserID, quiz.PodID, podSlug, note)
		}
		return
	}

	for i, d := range drafts {
		_, err := s.repo.CreateQuestion(ctx, sqlc.CreatePodQuizQuestionParams{
			QuizID:         quiz.ID,
			OrderIndex:     int32(i + 1),
			Difficulty:     int32(d.Difficulty),
			Prompt:         d.Prompt,
			Options:        mustJSON(d.Options),
			CorrectIndices: mustJSON(d.CorrectIndices),
			TopicTitle:     d.Topic,
			Weight:         1,
		})
		if err != nil {
			s.log.Error().Err(err).Int64("id", id).Msg("create quiz question failed")
			bg := context.WithoutCancel(ctx)
			_, _ = s.repo.FailGenerating(bg, id)
			if s.notifier != nil {
				s.notifier.NotifyQuizFailed(bg, quiz.UserID, quiz.PodID, podSlug, "Quiz generation failed; please try again")
			}
			return
		}
	}

	activated, err := s.repo.ActivateQuiz(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		s.log.Error().Err(err).Int64("id", id).Msg("activate quiz failed")
		return
	}

	q1, err := s.repo.GetQuestionByOrder(ctx, activated.ID, 1)
	if err != nil {
		s.log.Error().Err(err).Int64("id", id).Msg("get q1 after generate failed")
		return
	}
	if _, err := s.repo.UpsertAnswerStart(ctx, activated.ID, q1.ID); err != nil {
		s.log.Error().Err(err).Int64("id", id).Msg("start q1 timer after generate failed")
		return
	}

	if s.notifier != nil {
		s.notifier.NotifyQuizReady(context.WithoutCancel(ctx), activated.UserID, activated.PodID, podSlug)
	}
	s.log.Info().
		Str("quiz_id", activated.PublicID.String()).
		Str("pod", podSlug).
		Msg("pod quiz ready")
}

func (s *Service) StartQuestion(ctx context.Context, userPublicID uuid.UUID, podSlug, quizPublicID string, order int32) (startResponse, error) {
	user, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return startResponse{}, err
	}
	quiz, err := s.loadOwnedInProgress(ctx, user.ID, pod.ID, quizPublicID)
	if err != nil {
		return startResponse{}, err
	}
	q, err := s.repo.GetQuestionByOrder(ctx, quiz.ID, order)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return startResponse{}, logging.Reject(s.log, ErrQuestionNotFound, "start question rejected")
		}
		return startResponse{}, logging.Unexpected(s.log, err, "get question failed")
	}
	ans, err := s.repo.UpsertAnswerStart(ctx, quiz.ID, q.ID)
	if err != nil {
		return startResponse{}, logging.Unexpected(s.log, err, "upsert answer start failed")
	}
	left := questionSeconds
	if ans.StartedAt.Valid {
		elapsed := int(time.Since(ans.StartedAt.Time).Seconds())
		left = questionSeconds - elapsed
		if left < 0 {
			left = 0
		}
	}
	return startResponse{OrderIndex: int(order), SecondsLeft: left}, nil
}

func (s *Service) AnswerQuestion(
	ctx context.Context,
	userPublicID uuid.UUID,
	podSlug, quizPublicID string,
	order int32,
	selected []int,
) (answerResponse, error) {
	user, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return answerResponse{}, err
	}
	quiz, err := s.loadOwnedInProgress(ctx, user.ID, pod.ID, quizPublicID)
	if err != nil {
		return answerResponse{}, err
	}
	q, err := s.repo.GetQuestionByOrder(ctx, quiz.ID, order)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return answerResponse{}, logging.Reject(s.log, ErrQuestionNotFound, "answer rejected: question")
		}
		return answerResponse{}, logging.Unexpected(s.log, err, "get question failed")
	}

	ans, err := s.repo.GetAnswer(ctx, quiz.ID, q.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return answerResponse{}, logging.Reject(s.log, ErrAnswerNotStarted, "answer rejected: not started")
		}
		return answerResponse{}, logging.Unexpected(s.log, err, "get answer failed")
	}
	if ans.AnsweredAt.Valid {
		return answerResponse{}, logging.Reject(s.log, ErrAlreadyAnswered, "answer rejected: already answered")
	}
	if !ans.StartedAt.Valid {
		return answerResponse{}, logging.Reject(s.log, ErrAnswerNotStarted, "answer rejected: missing start")
	}

	correct := decodeIntSlice(q.CorrectIndices)
	timedOut := time.Since(ans.StartedAt.Time) > time.Duration(questionSeconds)*time.Second
	isCorrect := false
	if !timedOut {
		isCorrect = sameIntSet(selected, correct)
	}

	selectedJSON := mustJSON(normalizeSelected(selected))
	if _, err := s.repo.SaveAnswer(ctx, quiz.ID, q.ID, selectedJSON, isCorrect, timedOut); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return answerResponse{}, logging.Reject(s.log, ErrAlreadyAnswered, "answer rejected: race")
		}
		return answerResponse{}, logging.Unexpected(s.log, err, "save answer failed")
	}

	return answerResponse{
		IsCorrect:      isCorrect,
		TimedOut:       timedOut,
		CorrectIndices: correct,
	}, nil
}

func (s *Service) CompleteQuiz(ctx context.Context, userPublicID uuid.UUID, podSlug, quizPublicID string) (quizResponse, error) {
	user, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return quizResponse{}, err
	}
	quiz, err := s.loadOwnedInProgress(ctx, user.ID, pod.ID, quizPublicID)
	if err != nil {
		return quizResponse{}, err
	}

	questions, err := s.repo.ListQuestions(ctx, quiz.ID)
	if err != nil {
		return quizResponse{}, logging.Unexpected(s.log, err, "list questions failed")
	}
	for _, q := range questions {
		ans, err := s.repo.GetAnswer(ctx, quiz.ID, q.ID)
		if err != nil || !ans.AnsweredAt.Valid {
			return quizResponse{}, logging.Reject(s.log, ErrIncompleteAnswers, "complete rejected: incomplete")
		}
	}

	correctCount, err := s.repo.CountCorrect(ctx, quiz.ID)
	if err != nil {
		return quizResponse{}, logging.Unexpected(s.log, err, "count correct failed")
	}
	score := quiz.TopicCount * int32(correctCount)
	quiz, err = s.repo.CompleteQuiz(ctx, quiz.ID, int32(correctCount), score)
	if err != nil {
		return quizResponse{}, logging.Unexpected(s.log, err, "complete quiz failed")
	}

	s.log.Info().
		Str("quiz_id", quiz.PublicID.String()).
		Int32("score", quiz.Score).
		Msg("pod quiz completed")

	return s.buildQuizResponse(ctx, quiz, true)
}

func (s *Service) GetQuiz(ctx context.Context, userPublicID uuid.UUID, podSlug, quizPublicID string) (quizResponse, error) {
	user, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return quizResponse{}, err
	}
	quiz, err := s.repo.GetQuizByPublicID(ctx, mustParseUUID(quizPublicID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return quizResponse{}, logging.Reject(s.log, ErrQuizNotFound, "get quiz rejected")
		}
		return quizResponse{}, logging.Unexpected(s.log, err, "get quiz failed")
	}
	if quiz.UserID != user.ID || quiz.PodID != pod.ID {
		return quizResponse{}, logging.Reject(s.log, ErrQuizNotOwned, "get quiz forbidden")
	}
	return s.buildQuizResponse(ctx, quiz, quiz.Status == sqlc.PodQuizStatusCOMPLETED)
}

func (s *Service) ListMine(ctx context.Context, userPublicID uuid.UUID, podSlug string) ([]quizResponse, error) {
	user, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListMyQuizzes(ctx, user.ID, pod.ID, 20)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list my quizzes failed")
	}
	out := make([]quizResponse, 0, len(rows))
	for _, row := range rows {
		resp, err := s.buildQuizResponse(ctx, row, true)
		if err != nil {
			return nil, err
		}
		out = append(out, resp)
	}
	return out, nil
}

func (s *Service) PodLeaderboard(ctx context.Context, userPublicID uuid.UUID, podSlug string) ([]podLeaderboardEntry, error) {
	_, pod, err := s.requireAcceptedMember(ctx, userPublicID, podSlug)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListPodLeaderboard(ctx, pod.ID)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "pod leaderboard failed")
	}
	out := make([]podLeaderboardEntry, 0, len(rows))
	for i, row := range rows {
		out = append(out, podLeaderboardEntry{
			Username:   row.Username,
			PublicID:   row.PublicID.String(),
			Name:       row.Name,
			BestScore:  int(row.BestScore),
			TopicCount: int(row.TopicCount),
			Rank:       i + 1,
		})
	}
	return out, nil
}

func (s *Service) CommunityLeaderboard(ctx context.Context, userPublicID uuid.UUID, skillSlug string) ([]communityLeaderboardEntry, error) {
	if _, err := s.repo.GetUserByPublicID(ctx, userPublicID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, logging.Reject(s.log, ErrNotAcceptedMember, "community lb: user missing")
		}
		return nil, logging.Unexpected(s.log, err, "community lb get user failed")
	}
	skill, err := s.repo.GetSkillBySlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, logging.Reject(s.log, ErrSkillNotFound, "community lb: skill missing")
		}
		return nil, logging.Unexpected(s.log, err, "community lb get skill failed")
	}
	rows, err := s.repo.ListCommunityLeaderboard(ctx, skill.ID)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "community leaderboard failed")
	}
	out := make([]communityLeaderboardEntry, 0, len(rows))
	for i, row := range rows {
		out = append(out, communityLeaderboardEntry{
			PodSlug:     row.PodSlug,
			PodName:     row.PodName,
			TotalScore:  int(row.TotalScore),
			MemberCount: int(row.MemberCount),
			Rank:        i + 1,
		})
	}
	return out, nil
}

func (s *Service) requireAcceptedMember(ctx context.Context, userPublicID uuid.UUID, podSlug string) (sqlc.User, sqlc.GetPodBySlugRow, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotAcceptedMember, "member check: user")
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "get user failed")
	}
	pod, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrPodNotFound, "member check: pod")
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "get pod failed")
	}
	mem, err := s.repo.GetPodMembership(ctx, pod.ID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotAcceptedMember, "member check: membership")
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "get membership failed")
	}
	if mem.Status != sqlc.UserPodMembershipStatusACCEPTED {
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotAcceptedMember, "member check: not accepted")
	}
	return user, pod, nil
}

func (s *Service) loadOwnedInProgress(ctx context.Context, userID, podID int64, quizPublicID string) (sqlc.PodQuiz, error) {
	id, err := uuid.Parse(quizPublicID)
	if err != nil {
		return sqlc.PodQuiz{}, logging.Reject(s.log, ErrQuizNotFound, "bad quiz id")
	}
	quiz, err := s.repo.GetQuizByPublicID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodQuiz{}, logging.Reject(s.log, ErrQuizNotFound, "quiz missing")
		}
		return sqlc.PodQuiz{}, logging.Unexpected(s.log, err, "get quiz failed")
	}
	if quiz.UserID != userID || quiz.PodID != podID {
		return sqlc.PodQuiz{}, logging.Reject(s.log, ErrQuizNotOwned, "quiz not owned")
	}
	if quiz.Status == sqlc.PodQuizStatusGENERATING {
		return sqlc.PodQuiz{}, logging.Reject(s.log, ErrQuizGenerating, "quiz still generating")
	}
	if quiz.Status == sqlc.PodQuizStatusFAILED {
		return sqlc.PodQuiz{}, logging.Reject(s.log, ErrQuizFailed, "quiz generation failed")
	}
	if quiz.Status != sqlc.PodQuizStatusINPROGRESS {
		return sqlc.PodQuiz{}, logging.Reject(s.log, ErrQuizNotInProgress, "quiz not in progress")
	}
	return quiz, nil
}

func (s *Service) buildQuizResponse(ctx context.Context, quiz sqlc.PodQuiz, includeAnswers bool) (quizResponse, error) {
	questions, err := s.repo.ListQuestions(ctx, quiz.ID)
	if err != nil {
		return quizResponse{}, logging.Unexpected(s.log, err, "list questions for response failed")
	}
	qResp := make([]questionResponse, 0, len(questions))
	for _, q := range questions {
		item := questionResponse{
			OrderIndex: int(q.OrderIndex),
			Difficulty: int(q.Difficulty),
			Prompt:     q.Prompt,
			Options:    decodeStringSlice(q.Options),
			TopicTitle: q.TopicTitle,
		}
		ans, err := s.repo.GetAnswer(ctx, quiz.ID, q.ID)
		if err == nil {
			if ans.AnsweredAt.Valid {
				item.Answered = true
				c := ans.IsCorrect
				t := ans.TimedOut
				item.IsCorrect = &c
				item.TimedOut = &t
				if includeAnswers || ans.AnsweredAt.Valid {
					item.CorrectIndices = decodeIntSlice(q.CorrectIndices)
				}
			} else if ans.StartedAt.Valid {
				elapsed := int(time.Since(ans.StartedAt.Time).Seconds())
				left := questionSeconds - elapsed
				if left < 0 {
					left = 0
				}
				item.SecondsLeft = &left
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return quizResponse{}, logging.Unexpected(s.log, err, "get answer for response failed")
		}
		qResp = append(qResp, item)
	}

	resp := quizResponse{
		ID:           quiz.PublicID.String(),
		Status:       string(quiz.Status),
		TopicCount:   int(quiz.TopicCount),
		CorrectCount: int(quiz.CorrectCount),
		Score:        int(quiz.Score),
		Topics:       decodeStringSlice(quiz.CompletedTopicTitles),
		Questions:    qResp,
		CreatedAt:    quiz.CreatedAt.Time.UTC().Format(time.RFC3339),
	}
	if quiz.CompletedAt.Valid {
		s := quiz.CompletedAt.Time.UTC().Format(time.RFC3339)
		resp.CompletedAt = &s
	}
	return resp, nil
}

func sameIntSet(a, b []int) bool {
	a = normalizeSelected(a)
	b = normalizeSelected(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func normalizeSelected(in []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(in))
	for _, v := range in {
		if v < 0 || v > 4 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
