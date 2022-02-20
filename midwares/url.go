package midwares

import (
	"log"
)

func Url() func(c *SliceRouterContext) {
	return func(c *SliceRouterContext) {
		if c.Req.RequestURI == "/favicon.ico" {
			c.Abort()
		} else {
			log.Println("url 中间件拦截,请求url：", c.Req.RequestURI)
			c.Next()
			log.Println("url 中间件拦截回调")
		}

	}
}
