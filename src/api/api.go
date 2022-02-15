package api

import (
	"time"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/api/middleware"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/global"
	"github.com/fasthttp/router"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func New(gCtx global.Context) <-chan struct{} {
	done := make(chan struct{})
	loader := loaders.New(gCtx)

	gql := GqlHandler(gCtx, loader)

	router := router.New()

	g := router.Group("/{v}")
	mid := func(ctx *fasthttp.RequestCtx) {
		if err := middleware.Auth(gCtx)(ctx); err != nil {
			ctx.Response.Header.Add("X-Auth-Failure", err.Error())
			goto handler
		}

	handler:
		gql(ctx)
	}
	g.GET("", mid)
	g.POST("", mid)

	router.HandleOPTIONS = true
	router.GlobalOPTIONS = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Vary", "Origin")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")

		ctx.SetStatusCode(fasthttp.StatusNoContent)
	}

	server := fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			start := time.Now()
			defer func() {
				var err interface{}
				if err != nil {
					ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
				}
				l := logrus.WithFields(logrus.Fields{
					"status":     ctx.Response.StatusCode(),
					"duration":   int64(time.Since(start) / time.Millisecond),
					"entrypoint": "api",
					"path":       utils.B2S(ctx.Path()),
					"ip":         utils.B2S(ctx.Response.Header.Peek("Cf-Connecting-IP")),
					"origin":     utils.B2S(ctx.Response.Header.Peek("Origin")),
				})
				if err != nil {
					l.Error("panic in handler: ", err)
				} else {
					l.Info("")
				}
			}()

			router.Handler(ctx)
		},
		ReadTimeout:     time.Second * 10,
		WriteTimeout:    time.Second * 10,
		CloseOnShutdown: true,
		Name:            "7TV - GQL",
	}

	go func() {
		if err := server.ListenAndServe(gCtx.Config().Http.URI); err != nil {
			logrus.Fatal("failed to start api server: ", err)
		}
		close(done)
	}()

	go func() {
		<-gCtx.Done()
		_ = server.Shutdown()
	}()

	return done
}
