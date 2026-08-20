package quiz

import (
	"context"
	"encoding/json"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

func (r *Repository) GetUserByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, publicID)
}

func (r *Repository) GetPodBySlug(ctx context.Context, slug string) (sqlc.GetPodBySlugRow, error) {
	return r.queries.GetPodBySlug(ctx, slug)
}

func (r *Repository) GetPodByID(ctx context.Context, id int64) (sqlc.Pod, error) {
	return r.queries.GetPodByID(ctx, id)
}

func (r *Repository) GetSkillBySlug(ctx context.Context, slug string) (sqlc.Skill, error) {
	return r.queries.GetSkillBySlug(ctx, slug)
}

func (r *Repository) GetPodMembership(ctx context.Context, podID, userID int64) (sqlc.PodMembership, error) {
	return r.queries.GetPodMembershipByPodAndUser(ctx, sqlc.GetPodMembershipByPodAndUserParams{
		PodID:  podID,
		UserID: userID,
	})
}

func (r *Repository) ListCompletedTopicTitles(ctx context.Context, userID, skillID int64) ([]string, error) {
	return r.queries.ListCompletedTopicTitlesForUserSkill(ctx, sqlc.ListCompletedTopicTitlesForUserSkillParams{
		UserID:  userID,
		SkillID: skillID,
	})
}

func (r *Repository) GetInProgressQuiz(ctx context.Context, userID, podID int64) (sqlc.PodQuiz, error) {
	return r.queries.GetInProgressPodQuiz(ctx, sqlc.GetInProgressPodQuizParams{
		UserID: userID,
		PodID:  podID,
	})
}

func (r *Repository) CreateQuiz(ctx context.Context, arg sqlc.CreatePodQuizParams) (sqlc.PodQuiz, error) {
	return r.queries.CreatePodQuiz(ctx, arg)
}

func (r *Repository) CreateQuestion(ctx context.Context, arg sqlc.CreatePodQuizQuestionParams) (sqlc.PodQuizQuestion, error) {
	return r.queries.CreatePodQuizQuestion(ctx, arg)
}

func (r *Repository) GetQuizByID(ctx context.Context, id int64) (sqlc.PodQuiz, error) {
	return r.queries.GetPodQuizByID(ctx, id)
}

func (r *Repository) GetQuizByPublicID(ctx context.Context, id uuid.UUID) (sqlc.PodQuiz, error) {
	return r.queries.GetPodQuizByPublicID(ctx, id)
}

func (r *Repository) ListGeneratingIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListGeneratingPodQuizzes(ctx)
}

func (r *Repository) ActivateQuiz(ctx context.Context, id int64) (sqlc.PodQuiz, error) {
	return r.queries.ActivatePodQuiz(ctx, id)
}

func (r *Repository) FailGenerating(ctx context.Context, id int64) (sqlc.PodQuiz, error) {
	return r.queries.FailPodQuiz(ctx, id)
}

func (r *Repository) ListQuestions(ctx context.Context, quizID int64) ([]sqlc.PodQuizQuestion, error) {
	return r.queries.ListPodQuizQuestions(ctx, quizID)
}

func (r *Repository) GetQuestionByOrder(ctx context.Context, quizID int64, order int32) (sqlc.PodQuizQuestion, error) {
	return r.queries.GetPodQuizQuestionByOrder(ctx, sqlc.GetPodQuizQuestionByOrderParams{
		QuizID:     quizID,
		OrderIndex: order,
	})
}

func (r *Repository) UpsertAnswerStart(ctx context.Context, quizID, questionID int64) (sqlc.PodQuizAnswer, error) {
	return r.queries.UpsertPodQuizAnswerStart(ctx, sqlc.UpsertPodQuizAnswerStartParams{
		QuizID:     quizID,
		QuestionID: questionID,
	})
}

func (r *Repository) GetAnswer(ctx context.Context, quizID, questionID int64) (sqlc.PodQuizAnswer, error) {
	return r.queries.GetPodQuizAnswer(ctx, sqlc.GetPodQuizAnswerParams{
		QuizID:     quizID,
		QuestionID: questionID,
	})
}

func (r *Repository) SaveAnswer(ctx context.Context, quizID, questionID int64, selected []byte, isCorrect, timedOut bool) (sqlc.PodQuizAnswer, error) {
	return r.queries.SavePodQuizAnswer(ctx, sqlc.SavePodQuizAnswerParams{
		QuizID:          quizID,
		QuestionID:      questionID,
		SelectedIndices: selected,
		IsCorrect:       isCorrect,
		TimedOut:        timedOut,
	})
}

func (r *Repository) CountCorrect(ctx context.Context, quizID int64) (int64, error) {
	return r.queries.CountCorrectPodQuizAnswers(ctx, quizID)
}

func (r *Repository) CompleteQuiz(ctx context.Context, quizID int64, correct, score int32) (sqlc.PodQuiz, error) {
	return r.queries.CompletePodQuiz(ctx, sqlc.CompletePodQuizParams{
		ID:           quizID,
		CorrectCount: correct,
		Score:        score,
	})
}

func (r *Repository) AbandonQuiz(ctx context.Context, quizID int64) (sqlc.PodQuiz, error) {
	return r.queries.AbandonPodQuiz(ctx, quizID)
}

func (r *Repository) ListMyQuizzes(ctx context.Context, userID, podID int64, limit int32) ([]sqlc.PodQuiz, error) {
	return r.queries.ListMyPodQuizzes(ctx, sqlc.ListMyPodQuizzesParams{
		UserID: userID,
		PodID:  podID,
		Limit:  limit,
	})
}

func (r *Repository) ListPodLeaderboard(ctx context.Context, podID int64) ([]sqlc.ListPodLeaderboardRow, error) {
	return r.queries.ListPodLeaderboard(ctx, podID)
}

func (r *Repository) ListCommunityLeaderboard(ctx context.Context, skillID int64) ([]sqlc.ListCommunityPodLeaderboardRow, error) {
	return r.queries.ListCommunityPodLeaderboard(ctx, skillID)
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func decodeStringSlice(b []byte) []string {
	var out []string
	if len(b) == 0 {
		return out
	}
	_ = json.Unmarshal(b, &out)
	return out
}

func decodeIntSlice(b []byte) []int {
	var out []int
	if len(b) == 0 {
		return out
	}
	_ = json.Unmarshal(b, &out)
	return out
}
