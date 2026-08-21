package openrouter

import "testing"

func TestCloseIncompleteJSON_TruncatedRoadmap(t *testing.T) {
	raw := `{
  "topics": [
    {
      "name": "Foundations of Linear Algebra for Quantum Machine Learning",
      "outcome": "Learner can manipulate vectors, matrices, and complex numbers to describe quantum states.",
      "difficulty": "BEGINNER",
      "subtopics": [
        {
          "name": "Scalars, Vectors, and Vector Spaces",
          "estimated_hours": 2,
          "outcome": "Learner can define scalars, vectors, and perform basic vector addition and scalar multiplication.",
          "difficu`
	outline, err := parseCompactRoadmap(raw)
	if err != nil {
		t.Fatalf("parse repaired json: %v\n%s", err, repairJSON(raw))
	}
	if len(outline.Topics) == 0 {
		t.Fatal("expected at least one salvaged topic")
	}
	if outline.Topics[0].Name == "" {
		t.Fatal("expected salvaged topic name")
	}
}
