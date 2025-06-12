package breakr

import (
	"context"
	"github.com/genov8/breakr/config"
	"github.com/genov8/breakr/internal/breakr"
)

type Breaker struct {
	internal *breakr.Breaker
}

func New(cfg config.Config) *Breaker {
	return &Breaker{internal: breakr.New(cfg)}
}

func (b *Breaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return b.internal.Execute(fn)
}

func (b *Breaker) ExecuteCtx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	return b.internal.ExecuteCtx(ctx, fn)
}

func (b *Breaker) State() string {
	return b.internal.State().String()
}
