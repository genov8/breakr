package breakr

type Breaker struct {
}

func New() *Breaker {
	return &Breaker{}
}

func (b *Breaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return nil, nil
}
