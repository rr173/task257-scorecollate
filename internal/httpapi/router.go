package httpapi

import (
	"net/http"
	"regexp"

	"task257-scorecollate/internal/service"
)

// Handler 接收已匹配路径参数的路由处理函数。
type Handler func(w http.ResponseWriter, r *http.Request, params map[string]string)

type route struct {
	method  string
	re      *regexp.Regexp
	handler Handler
}

// API 是 HTTP 层，注册全部 /api 路由与根页面。
type API struct {
	svc    *service.Service
	routes []route
}

// New 构造 HTTP 层并注册路由。
func New(svc *service.Service) *API {
	a := &API{svc: svc}
	a.register()
	return a
}

func (a *API) handle(method, pattern string, h Handler) {
	a.routes = append(a.routes, route{method: method, re: regexp.MustCompile(pattern), handler: h})
}

func (a *API) register() {
	// 根页面（并排校勘视图）
	a.handle("GET", `^/$`, a.handleView)
	// 健康与自检
	a.handle("GET", `^/api/health$`, a.handleHealth)
	a.handle("POST", `^/api/projects/(?P<id>[^/]+)/selfcheck$`, a.handleSelfCheck)
	// 项目
	a.handle("POST", `^/api/projects$`, a.handleCreateProject)
	a.handle("GET", `^/api/projects$`, a.handleListProjects)
	a.handle("GET", `^/api/projects/(?P<id>[^/]+)$`, a.handleGetProject)
	a.handle("POST", `^/api/projects/(?P<id>[^/]+)/state$`, a.handleTransitionProject)
	// 来源
	a.handle("POST", `^/api/projects/(?P<id>[^/]+)/sources$`, a.handleCreateSource)
	a.handle("GET", `^/api/projects/(?P<id>[^/]+)/sources$`, a.handleListSources)
	a.handle("POST", `^/api/sources/(?P<id>[^/]+)/parent$`, a.handleReparentSource)
	// 片段
	a.handle("POST", `^/api/projects/(?P<id>[^/]+)/fragments$`, a.handleCreateFragment)
	a.handle("GET", `^/api/projects/(?P<id>[^/]+)/fragments$`, a.handleListFragments)
	a.handle("GET", `^/api/fragments/(?P<id>[^/]+)$`, a.handleGetFragment)
	a.handle("POST", `^/api/fragments/(?P<id>[^/]+)/state$`, a.handleSetFragmentState)
	a.handle("POST", `^/api/fragments/(?P<id>[^/]+)/parse$`, a.handleParseFragment)
	a.handle("GET", `^/api/fragments/(?P<id>[^/]+)/measures$`, a.handleListMeasures)
	// 对齐
	a.handle("POST", `^/api/projects/(?P<id>[^/]+)/align$`, a.handleAlignProject)
	// 异文
	a.handle("GET", `^/api/projects/(?P<id>[^/]+)/variants$`, a.handleListVariants)
	a.handle("GET", `^/api/variants/(?P<id>[^/]+)$`, a.handleGetVariant)
	a.handle("POST", `^/api/variants/(?P<id>[^/]+)/adjudicate$`, a.handleAdjudicateVariant)
	// 校勘版本
	a.handle("POST", `^/api/projects/(?P<id>[^/]+)/editions$`, a.handleCreateEdition)
	a.handle("GET", `^/api/projects/(?P<id>[^/]+)/editions$`, a.handleListEditions)
	a.handle("GET", `^/api/editions/(?P<id>[^/]+)$`, a.handleGetEdition)
	a.handle("POST", `^/api/editions/(?P<id>[^/]+)/publish$`, a.handlePublishEdition)
	a.handle("POST", `^/api/editions/(?P<id>[^/]+)/supersede$`, a.handleSupersedeEdition)
}

// ServeHTTP 按 方法 + 正则路径 匹配路由。
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	for _, rt := range a.routes {
		if rt.method != r.Method {
			continue
		}
		m := rt.re.FindStringSubmatch(path)
		if m == nil {
			continue
		}
		params := map[string]string{}
		for i, name := range rt.re.SubexpNames() {
			if i > 0 && name != "" {
				params[name] = m[i]
			}
		}
		rt.handler(w, r, params)
		return
	}
	writeError(w, http.StatusNotFound, "route not found")
}
