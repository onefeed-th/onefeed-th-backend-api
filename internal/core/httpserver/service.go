package httpserver

import "context"

type Service[TReq any, TResp any] func(ctx context.Context, req TReq) (TResp, error)
