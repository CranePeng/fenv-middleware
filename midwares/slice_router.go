package midwares

import (
	"context"
	"log"
	"math"
	"net/http"
	"strings"
)

// 路由中间件，目标定位是 tcp、http通用的中间件

// 可用来定制trace数据等

const abortIndex int8 = math.MaxInt8 / 2 //最多 63 个中间件
type (
	HandlerFunc func(*SliceRouterContext)

	// router 结构体
	SliceRouter struct {
		groups []*SliceGroup
	}

	// group 结构体
	SliceGroup struct {
		*SliceRouter
		path     string
		handlers []HandlerFunc
	}

	// router上下文
	SliceRouterContext struct {
		Rw  http.ResponseWriter
		Req *http.Request
		Ctx context.Context
		*SliceGroup
		index int8
	}
)

func newSliceRouterContext(rw http.ResponseWriter, req *http.Request, r *SliceRouter) *SliceRouterContext {
	newSliceGroup := &SliceGroup{}
	//最长url前缀匹配
	matchUrlLen := 0
	for _, group := range r.groups {
		//fmt.Println("req.RequestURI")
		//fmt.Println(req.RequestURI)
		if strings.HasPrefix(req.RequestURI, group.path) {
			pathLen := len(group.path)
			// 如果匹配上
			if pathLen > matchUrlLen {
				matchUrlLen = pathLen
				//浅拷贝数组指针
				*newSliceGroup = *group
			}
		}
	}
	c := &SliceRouterContext{Rw: rw, Req: req, SliceGroup: newSliceGroup, Ctx: req.Context()}
	c.Reset()
	return c
}

func (c *SliceRouterContext) Get(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

func (c *SliceRouterContext) Set(key, val interface{}) {
	c.Ctx = context.WithValue(c.Ctx, key, val)
}

// 从最先加入中间件开始回调
func (c *SliceRouterContext) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		log.Println("c.index")
		log.Println(c.index)
		c.handlers[c.index](c)
		c.index++
	}
}

// 跳出中间件方法
func (c *SliceRouterContext) Abort() {
	c.index = abortIndex
}

// 是否跳过了回调
func (c *SliceRouterContext) IsAborted() bool {
	return c.index >= abortIndex
}

// 重置回调
func (c *SliceRouterContext) Reset() {
	c.index = -1
}

/*
	路由处理器
*/
type SliceRouterHandler struct {
	coreFunc func(*SliceRouterContext) http.Handler
	router   *SliceRouter
}

func (w *SliceRouterHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c := newSliceRouterContext(rw, req, w.router)
	if w.coreFunc != nil {
		c.handlers = append(c.handlers, func(c *SliceRouterContext) {
			w.coreFunc(c).ServeHTTP(rw, req)
		})
	}
	// 每次执行的时候重置
	c.Reset()
	c.Next()
}

/*
	构造方法：路由处理器
*/
func NewSliceRouterHandler(coreFunc func(*SliceRouterContext) http.Handler, router *SliceRouter) *SliceRouterHandler {
	return &SliceRouterHandler{
		coreFunc: coreFunc,
		router:   router,
	}
}

/*
	构造方法：router
*/
func NewSliceRouter() *SliceRouter {
	return &SliceRouter{}
}

// 创建 Group
func (g *SliceRouter) Group(path string) *SliceGroup {
	return &SliceGroup{
		SliceRouter: g,
		path:        path,
	}
}

// 构造 路由回调方法
func (g *SliceGroup) Use(middlewares ...HandlerFunc) *SliceGroup {
	g.handlers = append(g.handlers, middlewares...)
	existsFlag := false
	for _, oldGroup := range g.SliceRouter.groups {
		if oldGroup == g {
			existsFlag = true
		}
	}
	if !existsFlag {
		g.SliceRouter.groups = append(g.SliceRouter.groups, g)
	}
	return g
}
