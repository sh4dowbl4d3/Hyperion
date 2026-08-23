package labs

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func validMeta() Meta {
	return Meta{
		Slug:              "test-lab",
		Name:              "Test Lab",
		Description:       "A lab used in tests.",
		Objective:         "Prove the test objective.",
		Category:          "testing",
		Difficulty:        DifficultyMedium,
		XP:                50,
		VulnerabilityType: "test-vulnerability",
		Hints:             []string{"First hint", "Second hint"},
	}
}

type stubLab struct {
	meta Meta
}

func (s stubLab) Meta() Meta { return s.meta }

func (stubLab) RegisterRoutes(*gin.RouterGroup) error { return nil }

func TestRegistryRegisterAndRetrieve(t *testing.T) {
	r := NewRegistry()
	lab := stubLab{meta: validMeta()}

	if err := r.Register(lab); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if r.Count() != 1 {
		t.Fatalf("count = %d, want 1", r.Count())
	}

	got, ok := r.Get("test-lab")
	if !ok {
		t.Fatal("Get could not find registered lab")
	}
	if got.Meta().Name != "Test Lab" {
		t.Errorf("name = %q, want Test Lab", got.Meta().Name)
	}

	if _, ok := r.Get("missing"); ok {
		t.Error("Get returned a lab for unknown slug")
	}
}

func TestRegistryRejectsDuplicateSlug(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(stubLab{meta: validMeta()}); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := r.Register(stubMetaWithSlug(validMeta(), "test-lab"))
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v, want duplicate slug failure", err)
	}
}

func stubMetaWithSlug(m Meta, slug string) stubLab {
	m.Slug = slug
	return stubLab{meta: m}
}

func TestRegistryRejectsInvalidMetadata(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Meta)
		want   string
	}{
		{"bad slug", func(m *Meta) { m.Slug = "Bad Slug!" }, "slug"},
		{"missing name", func(m *Meta) { m.Name = " " }, "name"},
		{"missing description", func(m *Meta) { m.Description = "" }, "description"},
		{"missing objective", func(m *Meta) { m.Objective = "" }, "objective"},
		{"missing category", func(m *Meta) { m.Category = "" }, "category"},
		{"bad difficulty", func(m *Meta) { m.Difficulty = "impossible" }, "difficulty"},
		{"zero xp", func(m *Meta) { m.XP = 0 }, "xp"},
		{"negative xp", func(m *Meta) { m.XP = -5 }, "xp"},
		{"missing vulnerability type", func(m *Meta) { m.VulnerabilityType = "" }, "vulnerability type"},
		{"no hints", func(m *Meta) { m.Hints = nil }, "hint"},
		{"empty hint", func(m *Meta) { m.Hints = []string{""} }, "hint"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			meta := validMeta()
			tc.mutate(&meta)
			r := NewRegistry()
			err := r.Register(stubLab{meta: meta})
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestRegistryAllIsSortedBySlug(t *testing.T) {
	r := NewRegistry()
	first := validMeta()
	first.Slug = "b-lab"
	second := validMeta()
	second.Slug = "a-lab"

	if err := r.Register(stubLab{meta: first}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(stubLab{meta: second}); err != nil {
		t.Fatal(err)
	}

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("len = %d, want 2", len(all))
	}
	if all[0].Meta().Slug != "a-lab" || all[1].Meta().Slug != "b-lab" {
		t.Errorf("order = [%s, %s], want [a-lab b-lab]", all[0].Meta().Slug, all[1].Meta().Slug)
	}
}
