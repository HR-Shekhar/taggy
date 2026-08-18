package openrouter

import "testing"

func TestRepairJSONTrailingCommas(t *testing.T) {
	in := `{"topics":[{"name":"HTML","subtopics":["A","B",]},{"name":"CSS","subtopics":["C",],},]}`
	out := repairJSON(in)
	outline, err := parseCompactRoadmap(out)
	if err != nil {
		t.Fatalf("parse after repair: %v\nraw=%s", err, out)
	}
	if len(outline.Topics) != 2 {
		t.Fatalf("topics=%d want 2", len(outline.Topics))
	}
	if len(outline.Topics[0].Subtopics) != 2 {
		t.Fatalf("subtopics=%v", outline.Topics[0].Subtopics)
	}
}

func TestFlattenCompactRoadmap(t *testing.T) {
	drafts := flattenCompactRoadmap(compactRoadmap{
		Topics: []compactTopic{
			{Name: "Supervised Learning", Subtopics: []string{"Linear Regression", "KNN"}},
		},
	})
	if len(drafts) != 3 {
		t.Fatalf("got %d drafts want 3", len(drafts))
	}
	if drafts[0].Kind != "CHAPTER" || drafts[1].Kind != "TOPIC" {
		t.Fatalf("unexpected kinds: %+v", drafts)
	}
}
