package quiz

import "errors"

var (
	ErrPodNotFound          = errors.New("pod not found")
	ErrSkillNotFound        = errors.New("skill not found")
	ErrNotAcceptedMember    = errors.New("must be an accepted pod member")
	ErrNoCompletedTopics    = errors.New("complete at least one topic before taking a quiz")
	ErrAIUnavailable        = errors.New("quiz generation is unavailable")
	ErrAIFailed             = errors.New("failed to generate quiz")
	ErrQuizNotFound         = errors.New("quiz not found")
	ErrQuizNotOwned         = errors.New("quiz does not belong to this user")
	ErrQuizNotInProgress    = errors.New("quiz is not in progress")
	ErrQuestionNotFound     = errors.New("question not found")
	ErrAlreadyAnswered      = errors.New("question already answered")
	ErrAnswerNotStarted     = errors.New("start the question timer first")
	ErrInProgressExists     = errors.New("finish or abandon your in-progress quiz first")
	ErrIncompleteAnswers    = errors.New("answer all questions before completing")
)
