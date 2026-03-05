package bsl

import "testing"

func TestSearchByRussianName(t *testing.T) {
	results := Search("СтрНайти")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "СтрНайти" {
		t.Fatalf("expected СтрНайти, got %s", results[0].Name)
	}
}

func TestSearchByEnglishName(t *testing.T) {
	results := Search("StrFind")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	results := Search("strfind")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for case-insensitive search, got %d", len(results))
	}
}

func TestSearchNotFound(t *testing.T) {
	results := Search("НесуществующаяФункция")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchPartialMatch(t *testing.T) {
	results := Search("Формат")
	if len(results) < 1 {
		t.Fatal("expected at least 1 result for partial match")
	}
}
