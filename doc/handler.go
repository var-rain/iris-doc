package doc

import (
	"github.com/kataras/iris/v12"
	"strings"
)

func New() iris.Handler {
	return func(ctx iris.Context) {
		if !IsOn() {
			ctx.Next()
			return
		}

		call := &Call{}
		Before(call, ctx.Request())

		w := ctx.Recorder()
		ctx.Next()

		if code := ctx.GetStatusCode(); IsStatusCodeValid(code) {
			call.MethodType = ctx.Method()
			call.CurrentPath = strings.Split(ctx.Request().RequestURI, "?")[0]
			call.ResponseBody = string(w.Body()[0:])
			call.ResponseCode = code

			m := make(map[string]string, len(w.Header()))
			for k, v := range w.Header() {
				m[k] = strings.Join(v, " ")
			}
			call.RequestHeader = m

			go GenerateHtml(call)
		}
	}
}
