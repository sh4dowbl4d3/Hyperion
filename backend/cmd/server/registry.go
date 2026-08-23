package main

import (
	"moderndvwa/backend/internal/labs"
)

func registryMetas(registry *labs.Registry) []labs.Meta {
	all := registry.All()
	metas := make([]labs.Meta, 0, len(all))
	for _, lab := range all {
		metas = append(metas, lab.Meta())
	}
	return metas
}
