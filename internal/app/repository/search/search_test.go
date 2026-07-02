package search

import (
	"testing"
)

func TestQueryStringAnalysis(t *testing.T) {
	query := QueryStringAnalysis("test ^test")
	if len(query) != 2 {
		t.Errorf("query length is not 1, got %d", len(query))
	}
	if query[0].Cond != Match || query[1].Cond != Match || query[0].Content != "test" || query[1].Content != "test" {
		t.Errorf("query condition is not Match, got %#+v", query)
	}
}

func TestQueryStringAnalysis_WildCard(t *testing.T) {
	query := QueryStringAnalysis("~test")
	if len(query) != 1 {
		t.Errorf("query length is not 1, got %d", len(query))
	}
	if query[0].Cond != WildCard || query[0].Content != "test" {
		t.Errorf("query condition is not Match, got %#+v", query)
	}
}

func TestQueryStringAnalysis_NotMatch(t *testing.T) {
	query := QueryStringAnalysis("!^test ^!test")
	if len(query) != 2 {
		t.Errorf("query length is not 2, got %d", len(query))
	}
	if query[0].Cond != NotMatch || query[1].Cond != NotMatch || query[0].Content != "test" || query[1].Content != "test" {
		t.Errorf("query condition is not NotMatch, got %v", query)
	}
}

func TestQueryStringAnalysis_NotWildCard(t *testing.T) {
	query := QueryStringAnalysis("~!test !~test")
	if len(query) != 2 {
		t.Errorf("query length is not 2, got %d", len(query))
	}
	if query[0].Cond != NotWildCard || query[1].Cond != NotWildCard || query[0].Content != "test" || query[1].Content != "test" {
		t.Errorf("query condition is not NotWildCard, got %#+v", query)
	}
}

func TestQueryStringAnalysis_Empty(t *testing.T) {
	query := QueryStringAnalysis("  ")
	if len(query) != 0 {
		t.Errorf("query length is not 0, got %d", len(query))
	}
}

func TestQueryStringAnalysis_Complex(t *testing.T) {
	query := QueryStringAnalysis("^debug '~!out of memory' !^error ")
	if len(query) != 3 {
		t.Errorf("query length is not 3, got %d", len(query))
	}
	if query[0].Cond != Match || query[0].Content != "debug" {
		t.Errorf("query condition is not Match, got %s %d", query[0].Content, query[0].Cond)
	}
	if query[1].Cond != NotWildCard || query[1].Content != "out of memory" {
		t.Errorf("query condition is not WildCard, got %s %d", query[1].Content, query[1].Cond)
	}
	if query[2].Cond != NotMatch || query[2].Content != "error" {
		t.Errorf("query condition is not NotMatch, got %s %d", query[2].Content, query[2].Cond)
	}
}
