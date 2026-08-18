package roadmap

import "errors"

var (
	ErrRoadmapNotFound        = errors.New("roadmap not found for skill")
	ErrVersionNotFound        = errors.New("roadmap version not found")
	ErrVersionNotSelectable   = errors.New("draft roadmap versions cannot be selected")
	ErrAlreadyOnVersion       = errors.New("already on this roadmap version")
)
