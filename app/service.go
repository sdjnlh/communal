package app

import (
	"code.letsit.cn/go/common"
)

type Service struct {
	BaseApp
}

func NewService(name string) *Service {
	service := &Service{
		BaseApp: BaseApp{
			name: name,
		},
	}
	service.SetPriority(PriorityHigh)
	return service
}

func (app *Service) Start(ctx *common.Context) error {
	app.Subscribe(app.name, app)

	err := (&app.BaseApp).Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
