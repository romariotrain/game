package models

import "testing"

func TestCalculateQuestEXP(t *testing.T) {
	got := CalculateQuestEXP(25, 3, 2)
	// round(25*0.6 + 3*4 + 2*3) = round(33) = 33
	if got != 33 {
		t.Fatalf("expected 33 exp, got %d", got)
	}
}

func TestRankFromEXPBoundaries(t *testing.T) {
	tests := []struct {
		exp  int
		want QuestRank
	}{
		{10, RankE},
		{11, RankD},
		{18, RankD},
		{19, RankC},
		{28, RankC},
		{29, RankB},
		{40, RankB},
		{41, RankA},
		{55, RankA},
		{56, RankS},
	}

	for _, tc := range tests {
		if got := RankFromEXP(tc.exp); got != tc.want {
			t.Fatalf("exp %d: expected %s, got %s", tc.exp, tc.want, got)
		}
	}
}

func TestAttemptsForQuestEXP(t *testing.T) {
	if got := AttemptsForQuestEXP(14); got != 1 {
		t.Fatalf("expected 1 attempt for exp=14, got %d", got)
	}
	if got := AttemptsForQuestEXP(15); got != 2 {
		t.Fatalf("expected 2 attempts for exp=15, got %d", got)
	}
	if got := AttemptsForQuestEXP(31); got != 3 {
		t.Fatalf("expected 3 attempts for exp=31, got %d", got)
	}
}
