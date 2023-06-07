package rpc

import (
	"context"
)

//var starter app.Starter

var Call func(ctx context.Context, servicePath string, serviceMethod string, args interface{}, reply interface{}) error
