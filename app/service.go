package app

import (
	"github.com/sdjnlh/communal"
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

func (app *Service) Start(ctx *communal.Context) error {
	app.Subscribe(app.name, app)

	err := (&app.BaseApp).Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
