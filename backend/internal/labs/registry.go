package labs

import (
	"fmt"
	"sort"
)

type Registry struct {
	bySlug map[string]Lab
	slugs  []string
}

func NewRegistry() *Registry {
	return &Registry{bySlug: make(map[string]Lab)}
}

func (r *Registry) Register(lab Lab) error {
	meta := lab.Meta()
	if err := meta.validate(); err != nil {
		return fmt.Errorf("register lab: %w", err)
	}
	if _, exists := r.bySlug[meta.Slug]; exists {
		return fmt.Errorf("register lab: duplicate slug %q", meta.Slug)
	}
	r.bySlug[meta.Slug] = lab
	r.slugs = append(r.slugs, meta.Slug)
	sort.Strings(r.slugs)
	return nil
}

func (r *Registry) MustRegister(lab Lab) {
	if err := r.Register(lab); err != nil {
		panic(err)
	}
}

func (r *Registry) Get(slug string) (Lab, bool) {
	lab, ok := r.bySlug[slug]
	return lab, ok
}

func (r *Registry) All() []Lab {
	all := make([]Lab, 0, len(r.slugs))
	for _, slug := range r.slugs {
		all = append(all, r.bySlug[slug])
	}
	return all
}

func (r *Registry) Count() int {
	return len(r.bySlug)
}
