package api

import (
	"context"

	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"hyperion/backend/internal/labs"
)

type fakeCatalog struct {
	entries []labs.CatalogEntry
	getErr  error
}

func (f *fakeCatalog) List(context.Context) ([]labs.CatalogEntry, error) {
	return f.entries, nil
}

func (f *fakeCatalog) Get(_ context.Context, slug string) (*labs.CatalogEntry, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for i := range f.entries {
		if f.entries[i].Slug == slug {
			return &f.entries[i], nil
		}
	}
	return nil, labs.ErrUnknownLab
}

type fakeProgress struct {
	records []labs.Record
}

func (f *fakeProgress) ForUser(context.Context, string) ([]labs.Record, error) {
	return f.records, nil
}

func (f *fakeProgress) SummaryForUser(ctx context.Context, userID string) (*labs.Summary, error) {
	summary := &labs.Summary{TotalLabs: len(f.records)}
	for _, rec := range f.records {
		if rec.Status == labs.StatusCompleted {
			summary.CompletedLabs++
			summary.XPEarned += rec.XPAwarded
		}
	}
	return summary, nil
}

func newLabsFixture(catalog *fakeCatalog, progress *fakeProgress) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	tokens := newTestTokens()
	router := NewRouter(Deps{
		Log:      discardAPILogger(),
		Tokens:   tokens,
		Catalog:  catalog,
		Progress: progress,
	})
	return router
}

func completedAt(timeUTC time.Time) *time.Time { return &timeUTC }

func TestLabsListRequiresAuthentication(t *testing.T) {
	router := newLabsFixture(&fakeCatalog{}, &fakeProgress{})
	rec := doJSON(router, http.MethodGet, "/api/v1/labs", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestLabsListMergesCatalogWithProgress(t *testing.T) {
	completed := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	catalog := &fakeCatalog{entries: []labs.CatalogEntry{
		{Slug: "lab-a", Name: "Lab A", Description: "First.", Category: "injection", Difficulty: labs.DifficultyEasy, XP: 50},
		{Slug: "lab-b", Name: "Lab B", Description: "Second.", Category: "auth", Difficulty: labs.DifficultyHard, XP: 120},
	}}
	progress := &fakeProgress{records: []labs.Record{
		{LabSlug: "lab-a", Status: labs.StatusCompleted, XPAwarded: 50, CompletedAt: completedAt(completed)},
	}}

	router := newLabsFixture(catalog, progress)
	token := issueTestToken(t, router)
	rec := doJSON(router, http.MethodGet, "/api/v1/labs", token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	items, _ := body["labs"].([]any)
	if len(items) != 2 {
		t.Fatalf("labs = %v, want 2 entries", body["labs"])
	}
	first := items[0].(map[string]any)
	if first["status"] != "completed" || first["xp_awarded"] != float64(50) {
		t.Errorf("lab-a merge wrong: %v", first)
	}
	second := items[1].(map[string]any)
	if second["status"] != "not_started" || second["xp_awarded"] != float64(0) {
		t.Errorf("lab-b default status wrong: %v", second)
	}
	if _, leaks := first["vulnerability_type"]; leaks {
		t.Error("listing must not expose vulnerability_type")
	}
	if _, leaks := second["objective"]; leaks {
		t.Error("listing must not expose objective")
	}
}

func TestLabsDetailHidesVulnerabilityTypeUntilCompleted(t *testing.T) {
	catalog := &fakeCatalog{entries: []labs.CatalogEntry{
		{
			Slug: "secret-lab", Name: "Secret Lab", Description: "Hidden vuln.",
			Category: "auth", Difficulty: labs.DifficultyHard, XP: 100,
			Objective:         "Do the thing.",
			VulnerabilityType: "super-secret-type",
			Hints:             []string{"hint one", "hint two"},
		},
	}}

	t.Run("not started", func(t *testing.T) {
		router := newLabsFixture(catalog, &fakeProgress{})
		rec := doJSON(router, http.MethodGet, "/api/v1/labs/secret-lab", issueTestToken(t, router), nil)
		body := decode(t, rec)
		lab := body["lab"].(map[string]any)
		if lab["vulnerability_type"] != nil {
			t.Errorf("vulnerability_type leaked before completion: %v", lab["vulnerability_type"])
		}
		hints, _ := lab["hints"].([]any)
		if len(hints) != 2 {
			t.Errorf("hints = %v, want both hints present for progressive reveal", hints)
		}
	})

	t.Run("completed", func(t *testing.T) {
		progress := &fakeProgress{records: []labs.Record{
			{LabSlug: "secret-lab", Status: labs.StatusCompleted, XPAwarded: 100},
		}}
		router := newLabsFixture(catalog, progress)
		rec := doJSON(router, http.MethodGet, "/api/v1/labs/secret-lab", issueTestToken(t, router), nil)
		body := decode(t, rec)
		lab := body["lab"].(map[string]any)
		if lab["vulnerability_type"] != "super-secret-type" {
			t.Errorf("vulnerability_type = %v, want revealed after completion", lab["vulnerability_type"])
		}
	})
}

func TestLabsDetailUnknownSlug(t *testing.T) {
	router := newLabsFixture(&fakeCatalog{}, &fakeProgress{})
	rec := doJSON(router, http.MethodGet, "/api/v1/labs/nope", issueTestToken(t, router), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	body := decode(t, rec)
	errObj, _ := body["error"].(map[string]any)
	if errObj == nil || errObj["code"] != "not_found" {
		t.Errorf("expected not_found envelope, got %v", body)
	}
}

func TestProgressOverviewReturnsSummaryAndRecords(t *testing.T) {
	progress := &fakeProgress{records: []labs.Record{
		{LabSlug: "done-lab", Status: labs.StatusCompleted, XPAwarded: 75},
		{LabSlug: "wip-lab", Status: labs.StatusInProgress, XPAwarded: 0},
	}}
	router := newLabsFixture(&fakeCatalog{}, progress)

	rec := doJSON(router, http.MethodGet, "/api/v1/progress", issueTestToken(t, router), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := decode(t, rec)
	summary := body["summary"].(map[string]any)
	if summary["xp_earned"] != float64(75) || summary["completed_labs"] != float64(1) {
		t.Errorf("summary = %v, want xp_earned 75 and completed 1", summary)
	}
	records, _ := body["records"].([]any)
	if len(records) != 2 {
		t.Errorf("records = %v, want 2", records)
	}
}
