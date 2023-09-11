package ui

import (
	"context"
)

type UI struct {
	ctx context.Context
}

func New() *UI {
	return &UI{}
}

func (u *UI) Startup(ctx context.Context) {
	u.ctx = ctx
}
