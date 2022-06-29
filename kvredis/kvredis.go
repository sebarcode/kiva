package kvredis

type RedisProvider struct {
}

func New() (*RedisProvider, error) {
	p := new(RedisProvider)
	return p, nil
}
