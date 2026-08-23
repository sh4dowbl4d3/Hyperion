package labs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type Meta struct {
	Slug              string
	Name              string
	Description       string
	Objective         string
	Category          string
	Difficulty        Difficulty
	XP                int
	VulnerabilityType string
	Hints             []string
}

func (m Meta) validate() error {
	switch {
	case !slugPattern.MatchString(m.Slug) || len(m.Slug) < 2 || len(m.Slug) > 40:
		return fmt.Errorf("lab %q: slug must be kebab-case, 2-40 chars", m.Slug)
	case strings.TrimSpace(m.Name) == "":
		return fmt.Errorf("lab %q: name is required", m.Slug)
	case strings.TrimSpace(m.Description) == "":
		return fmt.Errorf("lab %q: description is required", m.Slug)
	case strings.TrimSpace(m.Objective) == "":
		return fmt.Errorf("lab %q: objective is required", m.Slug)
	case strings.TrimSpace(m.Category) == "":
		return fmt.Errorf("lab %q: category is required", m.Slug)
	case m.Difficulty != DifficultyEasy && m.Difficulty != DifficultyMedium && m.Difficulty != DifficultyHard:
		return fmt.Errorf("lab %q: difficulty must be easy, medium or hard", m.Slug)
	case m.XP <= 0 || m.XP > 1000:
		return fmt.Errorf("lab %q: xp must be between 1 and 1000", m.Slug)
	case strings.TrimSpace(m.VulnerabilityType) == "":
		return fmt.Errorf("lab %q: vulnerability type is required", m.Slug)
	case len(m.Hints) == 0:
		return fmt.Errorf("lab %q: at least one hint is required", m.Slug)
	}
	for i, hint := range m.Hints {
		if strings.TrimSpace(hint) == "" {
			return fmt.Errorf("lab %q: hint %d is empty", m.Slug, i+1)
		}
	}
	return nil
}

type Lab interface {
	Meta() Meta
	RegisterRoutes(group *gin.RouterGroup) error
}
