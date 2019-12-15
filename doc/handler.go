package doc

import (
	"bytes"
	"github.com/kataras/iris/v12/context"
)

func New() context.Handler {
	return func(ctx context.Context) {
		if !IsOn() {
			ctx.Next()
			return
		}

		call := &Call{}
		Before(call, ctx.Request())

		ctx.Record()
		ctx.Next()

		r := NewResponseRecorder(ctx.Recorder().Naive())
		r.Body = bytes.NewBuffer(ctx.Recorder().Body())
		r.Status = ctx.Recorder().StatusCode()

		After(call, r, ctx.Request())
	}
}
