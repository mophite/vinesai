package langchain

import "context"

// 场景设置
type scene struct{ CallbacksHandler LogHandler }

func (s *scene) Name() string {
	//TODO implement me
	panic("implement me")
}

func (s *scene) Description() string {
	//TODO implement me
	panic("implement me")
}

func (s *scene) Call(ctx context.Context, input string) (string, error) {
	//TODO implement me
	panic("implement me")
}
