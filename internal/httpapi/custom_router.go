package httpapi

import (
	"net/http"
	"regexp"
	"strings"
)

// 简单的正则路由，支持路径参数
type route struct {
	method  string
	pattern *regexp.Regexp
	handler http.HandlerFunc
}

type Router struct {
	routes []route
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) HandleFunc(method, pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, route{
		method:  method,
		pattern: regexp.MustCompile("^" + convertPattern(pattern) + "$"),
		handler: handler,
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	for _, rt := range r.routes {
		if rt.method != req.Method {
			continue
		}
		if !rt.pattern.MatchString(path) {
			continue
		}
		params := make(map[string]string)
		matches := rt.pattern.FindStringSubmatch(path)
		names := rt.pattern.SubexpNames()
		for i, name := range names {
			if name != "" {
				params[name] = matches[i]
			}
		}
		// 将参数放入context或全局？这里简单通过query传递，实际用PathValue? 我们使用自定义，直接修改请求URL的查询参数
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
		rt.handler(w, req)
		return
	}
	http.NotFound(w, req)
}

// 将路径参数如{id}改为正则捕获组
func convertPattern(p string) string {
	p = strings.ReplaceAll(p, "{", "(?P<")
	p = strings.ReplaceAll(p, "}", ">[^/]+)")
	return p
}
