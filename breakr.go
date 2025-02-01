package breakr

type Breaker struct {
	config Config
}

func New(config Config) *Breaker {
	return &Breaker{
		config: config,
	}
}

func (b *Breaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return nil, nil
}
