package game

import (
	"testing"

	"solo-leveling/internal/models"
)

func TestRankFromSuggestion_Boundaries(t *testing.T) {
	tests := []struct {
		name string
		s    models.AISuggestion
		want models.QuestRank
	}{
		{
			name: "E max 10 exp",
			s:    models.AISuggestion{Minutes: 5, Effort: 1, Friction: 1, Stat: "INT"}, // 3 + 4 + 3 = 10
			want: models.RankE,
		},
		{
			name: "D min 11 exp",
			s:    models.AISuggestion{Minutes: 6, Effort: 1, Friction: 1, Stat: "INT"}, // 4 + 4 + 3 = 11
			want: models.RankD,
		},
		{
			name: "B boundary 29 exp",
			s:    models.AISuggestion{Minutes: 20, Effort: 4, Friction: 1, Stat: "INT"}, // 12 + 16 + 3 = 31
			want: models.RankB,
		},
		{
			name: "S high exp",
			s:    models.AISuggestion{Minutes: 60, Effort: 5, Friction: 3, Stat: "INT"}, // 36 + 20 + 9 = 65
			want: models.RankS,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RankFromSuggestion(tc.s)
			if got != tc.want {
				t.Fatalf("got rank %s, want %s", got, tc.want)
			}
		})
	}
}

func TestEXPFromSuggestion(t *testing.T) {
	got := EXPFromSuggestion(models.AISuggestion{
		Minutes:  25,
		Effort:   3,
		Friction: 2,
	})
	// round(25*0.6 + 3*4 + 2*3) = round(15 + 12 + 6) = 33
	if got != 33 {
		t.Fatalf("expected exp 33, got %d", got)
	}
}
