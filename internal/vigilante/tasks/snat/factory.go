package snat

import "github.com/conplementag/cops-vigilante/internal/vigilante/tasks"

func NewSnatTask() tasks.Task {
	return &snatTask{}
}
