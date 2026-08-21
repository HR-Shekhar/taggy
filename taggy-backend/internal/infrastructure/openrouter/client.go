package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	MaxMilestones        = 128
	minTopics            = 5
	maxTopics            = 16
	minSubtopics         = 3
	maxSubtopicsPerTopic = 8
)

// MilestoneDraft is a curriculum unit stored in draft_milestones JSON and later
// inserted into milestones (chapter + topic names; descriptions optional/empty).
type MilestoneDraft struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	EstimatedHours int    `json:"estimated_hours"`
	OrderIndex     int    `json:"order_index"`
	Difficulty     string `json:"difficulty"`
	Slug           string `json:"slug"`
	Chapter        string `json:"chapter"`
	Kind           string `json:"kind"` // CHAPTER | TOPIC
}

// Client talks to an OpenAI-compatible chat completions API.
type Client struct {
	apiKey    string
	model     string
	baseURL   string
	maxTokens int
	jsonMode  bool
	http      *http.Client
	log       zerolog.Logger
}

type Options struct {
	MaxTokens int
	JSONMode  bool
	Timeout   time.Duration
}

func New(apiKey, model, baseURL string, log zerolog.Logger) *Client {
	return NewWithOptions(apiKey, model, baseURL, Options{}, log)
}

func NewWithOptions(apiKey, model, baseURL string, opts Options, log zerolog.Logger) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://integrate.api.nvidia.com/v1"
	}
	if strings.TrimSpace(model) == "" {
		model = "nvidia/nemotron-3-super-120b-a12b"
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 8192
	}
	return &Client{
		apiKey:    strings.TrimSpace(apiKey),
		model:     model,
		baseURL:   strings.TrimRight(baseURL, "/"),
		maxTokens: maxTokens,
		jsonMode:  opts.JSONMode,
		http:      &http.Client{Timeout: timeout},
		log:       log,
	}
}

func (c *Client) Available() bool {
	return c != nil && c.apiKey != ""
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *respFormat   `json:"response_format,omitempty"`
	Temperature    float64       `json:"temperature"`
	TopP           float64       `json:"top_p,omitempty"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
	Stream         bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type respFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// compactRoadmap is the AI contract: chapters with outcomes + timed subtopics.
//
//	{"topics":[{"name":"...","outcome":"...","difficulty":"BEGINNER","subtopics":[{"name":"...","estimated_hours":2,"outcome":"...","difficulty":"BEGINNER"}]}]}
type compactRoadmap struct {
	Topics []compactTopic `json:"topics"`
}

type compactTopic struct {
	Name       string            `json:"name"`
	Outcome    string            `json:"outcome"`
	Difficulty string            `json:"difficulty"`
	Subtopics  []compactSubtopic `json:"subtopics"`
}

type compactSubtopic struct {
	Name           string `json:"name"`
	EstimatedHours int    `json:"estimated_hours"`
	Outcome        string `json:"outcome"`
	Difficulty     string `json:"difficulty"`
}

// SkillRequestEvaluation is the AI gate for whether a skill request is worth building.
type SkillRequestEvaluation struct {
	WorthConsidering bool   `json:"worth_considering"`
	Reason           string `json:"reason"`
}

// RoadmapEditEvaluation judges whether an edit rationale is relevant to the skill.
type RoadmapEditEvaluation struct {
	WorthConsidering bool   `json:"worth_considering"`
	Reason           string `json:"reason"`
}

// EvaluateSkillRequest judges whether a user skill request should be generated.
func (c *Client) EvaluateSkillRequest(ctx context.Context, name, description string) (SkillRequestEvaluation, error) {
	if !c.Available() {
		c.log.Error().Str("op", "evaluate_skill").Str("name", strings.TrimSpace(name)).Msg("ai skill evaluation skipped: api key not configured")
		return SkillRequestEvaluation{}, fmt.Errorf("ai api key not configured")
	}
	system := `You are Taggy's skill-request reviewer. Decide if a user request is worth turning into a full learning roadmap.
Return ONLY one valid JSON object:
{"worth_considering":true,"reason":"short plain reason"}

Accept (worth_considering=true) when the ask is a real learnable skill or subject with clear curriculum potential (e.g. web development, Spanish, piano, data analysis, public speaking).

Reject (worth_considering=false) when it is spam, nonsense, jokes, adult/illegal content, get-rich-quick scams, empty/vague one-word gibberish with no learning intent, or clearly not a skill people study on Taggy.

Reason must be one short sentence the requester can read.`

	user := fmt.Sprintf("Skill name: %s\n", strings.TrimSpace(name))
	if d := strings.TrimSpace(description); d != "" {
		user += "Description: " + d + "\n"
	}
	user += "Return the JSON decision only."

	raw, err := c.chatJSON(ctx, system, user)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("op", "evaluate_skill").
			Str("name", strings.TrimSpace(name)).
			Str("model", c.model).
			Msg("ai skill evaluation failed")
		return SkillRequestEvaluation{}, err
	}
	cleaned := repairJSON(raw)
	var ev SkillRequestEvaluation
	if err := json.Unmarshal([]byte(cleaned), &ev); err != nil {
		c.log.Error().
			Err(err).
			Str("op", "evaluate_skill").
			Str("name", strings.TrimSpace(name)).
			Str("snippet", truncate(raw, 500)).
			Msg("ai skill evaluation parse failed")
		return SkillRequestEvaluation{}, fmt.Errorf("parse evaluation json: %w", err)
	}
	ev.Reason = strings.TrimSpace(ev.Reason)
	if ev.Reason == "" {
		if ev.WorthConsidering {
			ev.Reason = "Looks like a learnable skill with a clear curriculum path."
		} else {
			ev.Reason = "This request does not look like a skill worth building a Taggy roadmap for."
		}
	}
	return ev, nil
}

// EvaluateRoadmapEdit judges whether a user's edit rationale is relevant and worth applying.
func (c *Client) EvaluateRoadmapEdit(ctx context.Context, skillName, description, rationale string) (RoadmapEditEvaluation, error) {
	if !c.Available() {
		c.log.Error().Str("op", "evaluate_roadmap_edit").Str("skill", strings.TrimSpace(skillName)).Msg("ai roadmap edit evaluation skipped: api key not configured")
		return RoadmapEditEvaluation{}, fmt.Errorf("ai api key not configured")
	}
	system := `You are Taggy's roadmap-edit reviewer. Decide if a user's suggested change is worth generating a new syllabus draft for admin review.
Return ONLY one valid JSON object:
{"worth_considering":true,"reason":"short plain reason"}

Accept (worth_considering=true) when the rationale is a clear, relevant curriculum improvement for this skill (add missing topics, reorder, update for modern tools, fix gaps, change depth).

Reject (worth_considering=false) when the rationale is spam, nonsense, empty/vague, unrelated to the skill, joke requests, illegal/adult content, or would not meaningfully improve the learning path.

Reason must be one short sentence the requester (and admin) can read.`

	user := fmt.Sprintf("Skill: %s\n", strings.TrimSpace(skillName))
	if d := strings.TrimSpace(description); d != "" {
		user += "Skill description: " + d + "\n"
	}
	user += "Edit rationale: " + strings.TrimSpace(rationale) + "\n"
	user += "Return the JSON decision only."

	raw, err := c.chatJSON(ctx, system, user)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("op", "evaluate_roadmap_edit").
			Str("skill", strings.TrimSpace(skillName)).
			Str("model", c.model).
			Msg("ai roadmap edit evaluation failed")
		return RoadmapEditEvaluation{}, err
	}
	cleaned := repairJSON(raw)
	var ev RoadmapEditEvaluation
	if err := json.Unmarshal([]byte(cleaned), &ev); err != nil {
		c.log.Error().
			Err(err).
			Str("op", "evaluate_roadmap_edit").
			Str("skill", strings.TrimSpace(skillName)).
			Str("snippet", truncate(raw, 500)).
			Msg("ai roadmap edit evaluation parse failed")
		return RoadmapEditEvaluation{}, fmt.Errorf("parse edit evaluation json: %w", err)
	}
	ev.Reason = strings.TrimSpace(ev.Reason)
	if ev.Reason == "" {
		if ev.WorthConsidering {
			ev.Reason = "Relevant curriculum improvement for this skill."
		} else {
			ev.Reason = "This edit suggestion does not look relevant or useful for this skill."
		}
	}
	return ev, nil
}

// GenerateRoadmap asks for a followable syllabus (topics → timed subtopics + outcomes),
// then flattens into CHAPTER + TOPIC milestones. currentOutline is optional context for edits.
func (c *Client) GenerateRoadmap(
	ctx context.Context,
	skillName, description, rationale, currentOutline string,
) ([]MilestoneDraft, error) {
	if !c.Available() {
		c.log.Error().Str("op", "generate_roadmap").Str("skill", skillName).Msg("ai roadmap generation skipped: api key not configured")
		return nil, fmt.Errorf("ai api key not configured")
	}

	start := time.Now()
	system := `You are Taggy's senior curriculum architect. Design a realistic, followable roadmap a motivated beginner can complete week by week — the quality bar of a great online course outline, not a keyword dump.
Return ONLY one valid JSON object (strict JSON: no trailing commas, no comments, no markdown) matching:
{"topics":[{"name":"Chapter name","outcome":"One sentence: what the learner can do after this chapter","difficulty":"BEGINNER","subtopics":[{"name":"Lesson name","estimated_hours":2,"outcome":"One sentence lesson goal","difficulty":"BEGINNER"}]}]}

Audience & tone:
- Absolute beginner → competent / job-ready or practically skilled
- Plain human language; pair jargon with everyday words when needed
- Order easiest → harder; each chapter assumes only prior chapters
- Ban vague titles alone: "Basics", "Advanced", "Overview", "Introduction", "Misc", "Deep dive"

Difficulty (required on every chapter and lesson — must vary across the course):
- Exactly one of: BEGINNER, INTERMEDIATE, ADVANCED
- Early foundation chapters/lessons: mostly BEGINNER
- Mid course core skills: mostly INTERMEDIATE
- Later projects, polish, and advanced tooling: INTERMEDIATE and ADVANCED
- Do NOT label everything INTERMEDIATE — spread levels honestly along the path

Structure (critical):
- 10–14 chapters covering foundations → core skills → practice → tooling → polish
- Each chapter: 5–8 concrete lessons (real study units someone can schedule)
- Where natural, end a chapter with a small practice/project lesson
- estimated_hours = realistic focused study hours per lesson, typically 1–6 (integer, never 0)
- outcome strings are one short sentence each; keep names concise
- Aim for a complete path (~70–110 lessons including chapter headers); do not stop early
- Output must parse in a strict JSON parser`

	user := fmt.Sprintf("Skill: %s\n", skillName)
	if d := strings.TrimSpace(description); d != "" {
		user += "Description: " + d + "\n"
	}
	if o := strings.TrimSpace(currentOutline); o != "" {
		user += "Current roadmap outline to improve (preserve what still works; revise using the rationale):\n" + o + "\n"
	}
	if r := strings.TrimSpace(rationale); r != "" {
		user += "Update rationale: " + r + "\n"
	}
	user += "Write a complete, followable beginner-friendly syllabus from zero to a competent level with realistic hours on every lesson.\n"
	user += "Return the JSON outline only."

	isEdit := strings.TrimSpace(rationale) != "" || strings.TrimSpace(currentOutline) != ""
	op := "generate_roadmap"
	if isEdit {
		op = "generate_roadmap_edit"
	}

	raw, err := c.chatJSON(ctx, system, user)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("op", op).
			Str("skill", skillName).
			Str("model", c.model).
			Str("base_url", c.baseURL).
			Dur("latency", time.Since(start)).
			Msg("ai roadmap generation request failed")
		return nil, err
	}
	outline, err := parseCompactRoadmap(raw)
	if err != nil {
		c.log.Warn().
			Err(err).
			Str("op", op).
			Str("skill", skillName).
			Str("snippet", truncate(raw, 500)).
			Msg("ai compact roadmap parse failed; retrying")
		raw, err = c.chatJSON(ctx, system, user+"\nImportant: strict JSON only — no trailing commas. Subtopics must be objects with name, estimated_hours, outcome, difficulty (BEGINNER|INTERMEDIATE|ADVANCED).")
		if err != nil {
			c.log.Error().
				Err(err).
				Str("op", op).
				Str("skill", skillName).
				Str("model", c.model).
				Dur("latency", time.Since(start)).
				Msg("ai roadmap generation retry request failed")
			return nil, err
		}
		outline, err = parseCompactRoadmap(raw)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("op", op).
				Str("skill", skillName).
				Str("snippet", truncate(raw, 800)).
				Msg("ai compact roadmap parse failed after retry")
			return nil, fmt.Errorf("parse roadmap json: %w", err)
		}
	}

	drafts := flattenCompactRoadmap(outline)
	if len(drafts) < minTopics {
		err := fmt.Errorf("too few milestones generated (%d)", len(drafts))
		c.log.Error().
			Err(err).
			Str("op", op).
			Str("skill", skillName).
			Int("milestones", len(drafts)).
			Int("topics", len(outline.Topics)).
			Msg("ai roadmap generation produced too few milestones")
		return nil, err
	}

	topicCount := len(outline.Topics)
	if topicCount > maxTopics {
		topicCount = maxTopics
	}
	c.log.Info().
		Str("model", c.model).
		Str("base_url", c.baseURL).
		Dur("latency", time.Since(start)).
		Int("topics_raw", len(outline.Topics)).
		Int("topics_used", topicCount).
		Int("milestones", len(drafts)).
		Int("max_milestones", MaxMilestones).
		Msg("ai compact roadmap generated")

	return drafts, nil
}

// QuizQuestionDraft is one AI-generated multiple-choice item (answers included for storage).
type QuizQuestionDraft struct {
	Topic          string   `json:"topic"`
	Difficulty     int      `json:"difficulty"`
	Prompt         string   `json:"prompt"`
	Options        []string `json:"options"`
	CorrectIndices []int    `json:"correct"`
}

type compactQuiz struct {
	Questions []QuizQuestionDraft `json:"questions"`
}

const quizQuestionCount = 10

// GenerateQuiz builds 10 timed MCQs from completed topic titles.
func (c *Client) GenerateQuiz(ctx context.Context, topicTitles []string) ([]QuizQuestionDraft, error) {
	if !c.Available() {
		return nil, fmt.Errorf("ai api key not configured")
	}
	titles := make([]string, 0, len(topicTitles))
	for _, t := range topicTitles {
		t = strings.TrimSpace(t)
		if t != "" {
			titles = append(titles, t)
		}
	}
	if len(titles) == 0 {
		return nil, fmt.Errorf("no topics provided")
	}

	start := time.Now()
	system := `You are Taggy's quiz writer for learners who finished specific roadmap topics.
Return ONLY one valid JSON object (strict JSON: no trailing commas, no comments, no markdown):
{"questions":[{"topic":"Topic title","difficulty":1,"prompt":"Question?","options":["A","B","C","D","E"],"correct":[0,2]}]}

Rules:
- Exactly 10 questions
- Each question: exactly 5 short options (indexes 0–4)
- correct = array of one or more correct option indexes (multi-select allowed)
- difficulty must be 1 through 10 in increasing order across the 10 questions
- Each question must be solvable in under 60 seconds by someone who studied that topic
- Draw randomly across the provided completed topics (reuse topics if fewer than 10)
- Plain beginner-friendly wording; no trick questions; no empty options
- Output must parse in a strict JSON parser`

	user := "Completed topics:\n"
	for i, t := range titles {
		user += fmt.Sprintf("%d. %s\n", i+1, t)
	}
	user += "Return the JSON quiz only."

	raw, err := c.chatJSON(ctx, system, user)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("op", "generate_quiz").
			Str("model", c.model).
			Int("topics", len(titles)).
			Dur("latency", time.Since(start)).
			Msg("ai quiz generation request failed")
		return nil, err
	}
	quiz, err := parseCompactQuiz(raw)
	if err != nil {
		c.log.Warn().Err(err).Str("snippet", truncate(raw, 500)).Msg("ai quiz parse failed; retrying")
		raw, err = c.chatJSON(ctx, system, user+"\nImportant: strict JSON only — no trailing commas. Exactly 10 questions.")
		if err != nil {
			c.log.Error().
				Err(err).
				Str("op", "generate_quiz").
				Str("model", c.model).
				Dur("latency", time.Since(start)).
				Msg("ai quiz generation retry request failed")
			return nil, err
		}
		quiz, err = parseCompactQuiz(raw)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("op", "generate_quiz").
				Str("snippet", truncate(raw, 800)).
				Msg("ai quiz parse failed after retry")
			return nil, fmt.Errorf("parse quiz json: %w", err)
		}
	}

	out := normalizeQuiz(quiz.Questions, titles)
	if len(out) != quizQuestionCount {
		err := fmt.Errorf("expected %d valid questions, got %d", quizQuestionCount, len(out))
		c.log.Error().
			Err(err).
			Str("op", "generate_quiz").
			Int("got", len(out)).
			Int("raw_questions", len(quiz.Questions)).
			Msg("ai quiz generation produced invalid questions")
		return nil, err
	}

	c.log.Info().
		Str("model", c.model).
		Dur("latency", time.Since(start)).
		Int("topics", len(titles)).
		Int("questions", len(out)).
		Msg("ai quiz generated")
	return out, nil
}

func parseCompactQuiz(raw string) (compactQuiz, error) {
	cleaned := repairJSON(raw)
	var quiz compactQuiz
	if err := json.Unmarshal([]byte(cleaned), &quiz); err != nil {
		return compactQuiz{}, err
	}
	if len(quiz.Questions) == 0 {
		return compactQuiz{}, fmt.Errorf("empty questions array")
	}
	return quiz, nil
}

func normalizeQuiz(in []QuizQuestionDraft, fallbackTopics []string) []QuizQuestionDraft {
	out := make([]QuizQuestionDraft, 0, quizQuestionCount)
	for i, q := range in {
		if len(out) >= quizQuestionCount {
			break
		}
		prompt := strings.TrimSpace(q.Prompt)
		if prompt == "" {
			continue
		}
		opts := make([]string, 0, 5)
		for _, o := range q.Options {
			o = strings.TrimSpace(o)
			if o != "" {
				opts = append(opts, o)
			}
		}
		if len(opts) < 5 {
			continue
		}
		if len(opts) > 5 {
			opts = opts[:5]
		}
		seen := map[int]struct{}{}
		correct := make([]int, 0, len(q.CorrectIndices))
		for _, idx := range q.CorrectIndices {
			if idx < 0 || idx > 4 {
				continue
			}
			if _, ok := seen[idx]; ok {
				continue
			}
			seen[idx] = struct{}{}
			correct = append(correct, idx)
		}
		if len(correct) == 0 {
			continue
		}
		topic := strings.TrimSpace(q.Topic)
		if topic == "" && len(fallbackTopics) > 0 {
			topic = fallbackTopics[i%len(fallbackTopics)]
		}
		diff := q.Difficulty
		if diff < 1 || diff > 10 {
			diff = len(out) + 1
		}
		out = append(out, QuizQuestionDraft{
			Topic:          topic,
			Difficulty:     diff,
			Prompt:         prompt,
			Options:        opts,
			CorrectIndices: correct,
		})
	}
	// Force increasing difficulty labels 1..n for stored order.
	for i := range out {
		out[i].Difficulty = i + 1
	}
	return out
}

var trailingCommaRE = regexp.MustCompile(`,\s*([}\]])`)

func parseCompactRoadmap(raw string) (compactRoadmap, error) {
	cleaned := repairJSON(raw)
	var outline compactRoadmap
	if err := json.Unmarshal([]byte(cleaned), &outline); err != nil {
		// Legacy shape: subtopics as string arrays — normalize via raw map.
		var legacy struct {
			Topics []struct {
				Name      string          `json:"name"`
				Outcome   string          `json:"outcome"`
				Subtopics json.RawMessage `json:"subtopics"`
			} `json:"topics"`
		}
		if err2 := json.Unmarshal([]byte(cleaned), &legacy); err2 != nil {
			return compactRoadmap{}, err
		}
		outline.Topics = make([]compactTopic, 0, len(legacy.Topics))
		for _, t := range legacy.Topics {
			topic := compactTopic{Name: t.Name, Outcome: t.Outcome}
			var objs []compactSubtopic
			if json.Unmarshal(t.Subtopics, &objs) == nil && len(objs) > 0 {
				topic.Subtopics = objs
			} else {
				var names []string
				if json.Unmarshal(t.Subtopics, &names) == nil {
					for _, n := range names {
						topic.Subtopics = append(topic.Subtopics, compactSubtopic{Name: n, EstimatedHours: 2})
					}
				}
			}
			outline.Topics = append(outline.Topics, topic)
		}
	}
	if len(outline.Topics) == 0 {
		return compactRoadmap{}, fmt.Errorf("empty topics array")
	}
	return outline, nil
}

func repairJSON(s string) string {
	s = extractJSONObject(s)
	prev := ""
	for s != prev {
		prev = s
		s = trailingCommaRE.ReplaceAllString(s, "$1")
	}
	return strings.TrimSpace(s)
}

func clampHours(h int) int {
	if h < 1 {
		return 1
	}
	if h > 40 {
		return 40
	}
	return h
}

// normalizeDifficulty maps AI output to BEGINNER | INTERMEDIATE | ADVANCED.
func normalizeDifficulty(raw string, fallback string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "_")
	switch s {
	case "BEGINNER", "EASY", "INTRO", "INTRODUCTORY", "BASIC", "FUNDAMENTALS":
		return "BEGINNER"
	case "ADVANCED", "HARD", "EXPERT", "ADVANCED_LEVEL":
		return "ADVANCED"
	case "INTERMEDIATE", "MEDIUM", "MID":
		return "INTERMEDIATE"
	}
	fb := strings.ToUpper(strings.TrimSpace(fallback))
	switch fb {
	case "BEGINNER", "INTERMEDIATE", "ADVANCED":
		return fb
	}
	return "INTERMEDIATE"
}

// chapterDifficultyFromSubs picks the chapter label from its lessons (mode, then hardest).
func chapterDifficultyFromSubs(subs []compactSubtopic, topicFallback string) string {
	counts := map[string]int{"BEGINNER": 0, "INTERMEDIATE": 0, "ADVANCED": 0}
	for _, s := range subs {
		d := normalizeDifficulty(s.Difficulty, topicFallback)
		counts[d]++
	}
	best := normalizeDifficulty(topicFallback, "INTERMEDIATE")
	bestN := -1
	for _, d := range []string{"BEGINNER", "INTERMEDIATE", "ADVANCED"} {
		if counts[d] > bestN {
			bestN = counts[d]
			best = d
		}
	}
	if bestN <= 0 {
		return normalizeDifficulty(topicFallback, "INTERMEDIATE")
	}
	return best
}

func flattenCompactRoadmap(outline compactRoadmap) []MilestoneDraft {
	out := make([]MilestoneDraft, 0, MaxMilestones)
	seenSlug := map[string]struct{}{}

	topics := outline.Topics
	if len(topics) > maxTopics {
		topics = topics[:maxTopics]
	}

	for topicPos, t := range topics {
		if len(out) >= MaxMilestones {
			break
		}
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}

		// Progressive default if AI omits difficulty: early BEGINNER → mid INTERMEDIATE → late ADVANCED
		progressFallback := "INTERMEDIATE"
		fracDenom := len(topics) - 1
		if fracDenom < 1 {
			fracDenom = 1
		}
		frac := float64(topicPos) / float64(fracDenom)
		if frac < 0.34 {
			progressFallback = "BEGINNER"
		} else if frac > 0.72 {
			progressFallback = "ADVANCED"
		}

		subs := t.Subtopics
		if len(subs) > maxSubtopicsPerTopic {
			subs = subs[:maxSubtopicsPerTopic]
		}

		chapterDiff := normalizeDifficulty(t.Difficulty, progressFallback)
		if strings.TrimSpace(t.Difficulty) == "" {
			chapterDiff = chapterDifficultyFromSubs(subs, progressFallback)
		}

		chapterIdx := len(out)
		chapterSlug := uniqueSlug(slugify(name), "chapter", seenSlug)
		chapterOutcome := strings.TrimSpace(t.Outcome)
		out = append(out, MilestoneDraft{
			Title:          name,
			Description:    chapterOutcome,
			EstimatedHours: 1, // replaced after summing subtopics
			OrderIndex:     chapterIdx + 1,
			Difficulty:     chapterDiff,
			Slug:           chapterSlug,
			Chapter:        name,
			Kind:           "CHAPTER",
		})

		chapterHours := 0
		addedSubs := 0
		for subPos, rawSub := range subs {
			if len(out) >= MaxMilestones {
				break
			}
			sub := strings.TrimSpace(rawSub.Name)
			if sub == "" || strings.EqualFold(sub, name) {
				continue
			}
			hours := clampHours(rawSub.EstimatedHours)
			if rawSub.EstimatedHours <= 0 {
				hours = 2
			}
			subFallback := chapterDiff
			if strings.TrimSpace(rawSub.Difficulty) == "" && len(subs) > 1 {
				denom := len(subs) - 1
				if denom < 1 {
					denom = 1
				}
				subFrac := float64(subPos) / float64(denom)
				if subFrac < 0.33 {
					subFallback = normalizeDifficulty(chapterDiff, "BEGINNER")
					if chapterDiff == "ADVANCED" {
						subFallback = "INTERMEDIATE"
					}
				} else if subFrac > 0.66 && chapterDiff != "BEGINNER" {
					subFallback = chapterDiff
				}
			}
			subDiff := normalizeDifficulty(rawSub.Difficulty, subFallback)
			subSlug := uniqueSlug(slugify(sub), "topic", seenSlug)
			out = append(out, MilestoneDraft{
				Title:          sub,
				Description:    strings.TrimSpace(rawSub.Outcome),
				EstimatedHours: hours,
				OrderIndex:     len(out) + 1,
				Difficulty:     subDiff,
				Slug:           subSlug,
				Chapter:        name,
				Kind:           "TOPIC",
			})
			chapterHours += hours
			addedSubs++
			if addedSubs >= maxSubtopicsPerTopic {
				break
			}
		}
		if chapterHours < 1 {
			chapterHours = 1
		}
		out[chapterIdx].EstimatedHours = chapterHours
	}
	return out
}

func uniqueSlug(base, prefix string, seen map[string]struct{}) string {
	slug := base
	if slug == "" {
		slug = prefix
	}
	if prefix == "chapter" && !strings.HasPrefix(slug, "chapter-") {
		slug = "chapter-" + slug
	}
	if _, ok := seen[slug]; !ok {
		seen[slug] = struct{}{}
		return slug
	}
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		if _, ok := seen[candidate]; !ok {
			seen[candidate] = struct{}{}
			return candidate
		}
	}
	fallback := fmt.Sprintf("%s-%d", slug, len(seen)+1)
	seen[fallback] = struct{}{}
	return fallback
}

func (c *Client) chatJSON(ctx context.Context, system, user string) (string, error) {
	const maxAttempts = 5
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		content, err := c.chatJSONOnce(ctx, system, user)
		if err == nil {
			return content, nil
		}
		lastErr = err
		if !isRetryableAIError(err) || attempt == maxAttempts {
			return "", err
		}
		backoff := time.Duration(attempt*attempt) * 2 * time.Second
		c.log.Warn().
			Err(err).
			Int("attempt", attempt).
			Dur("backoff", backoff).
			Msg("ai request retrying after rate limit/transient error")
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
	}
	return "", lastErr
}

func isRetryableAIError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate") ||
		strings.Contains(msg, "status 503") ||
		strings.Contains(msg, "status 502")
}

func (c *Client) chatJSONOnce(ctx context.Context, system, user string) (string, error) {
	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.3,
		TopP:        1,
		MaxTokens:   c.maxTokens,
		Stream:      false,
	}
	if c.jsonMode {
		payload.ResponseFormat = &respFormat{Type: "json_object"}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("model", c.model).
			Str("base_url", c.baseURL).
			Dur("latency", time.Since(start)).
			Msg("ai chat request transport failed")
		return "", err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		c.log.Error().
			Err(err).
			Str("model", c.model).
			Dur("latency", time.Since(start)).
			Msg("ai chat response body read failed")
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(raw))
		if len(snippet) > 400 {
			snippet = snippet[:400] + "…"
		}
		evt := c.log.Error()
		if res.StatusCode == http.StatusTooManyRequests ||
			res.StatusCode == http.StatusBadGateway ||
			res.StatusCode == http.StatusServiceUnavailable {
			evt = c.log.Warn()
		}
		evt.
			Int("status", res.StatusCode).
			Dur("latency", time.Since(start)).
			Str("model", c.model).
			Str("base_url", c.baseURL).
			Str("body", snippet).
			Msg("ai chat completion failed")
		switch res.StatusCode {
		case http.StatusPaymentRequired:
			return "", fmt.Errorf("ai provider payment required (402)")
		case http.StatusTooManyRequests:
			return "", fmt.Errorf("ai provider rate limited (429)")
		default:
			return "", fmt.Errorf("ai provider status %d", res.StatusCode)
		}
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		c.log.Error().
			Err(err).
			Str("model", c.model).
			Str("snippet", truncate(string(raw), 400)).
			Msg("ai chat response unmarshal failed")
		return "", err
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		c.log.Error().
			Str("model", c.model).
			Int("choices", len(parsed.Choices)).
			Str("snippet", truncate(string(raw), 400)).
			Msg("ai chat empty response")
		return "", fmt.Errorf("ai provider empty response")
	}
	return extractJSONObject(parsed.Choices[0].Message.Content), nil
}

func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevHyphen = false
		default:
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "milestone"
	}
	if len(out) > 60 {
		out = strings.Trim(out[:60], "-")
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
