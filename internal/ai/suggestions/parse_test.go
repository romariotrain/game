package suggestions

import "testing"

func TestParseJSON_Array(t *testing.T) {
	raw := `[
		{"title":"Задача 1","desc":"Описание 1","minutes":20,"effort":2,"friction":1,"stat":"INT","tags":["work"]},
		{"title":"Задача 2","desc":"Описание 2","minutes":15,"effort":1,"friction":1,"stat":"AGI","tags":["health"]}
	]`

	items, err := ParseJSON(raw)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Задача 1" {
		t.Fatalf("unexpected first title: %q", items[0].Title)
	}
}
