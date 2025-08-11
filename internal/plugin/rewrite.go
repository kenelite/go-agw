package plugin

import (
	"fmt"
	"strings"
)

// RewritePlugin performs routing and request rewrite actions.
// Config example:
// name: rewrite
// config:
//
//	set_path: "/v2${path}"
//	strip_prefix: "/api"
//	add_prefix: "/edge"
//	add_headers:
//	  X-From: "go-agw"
//	set_upstream: "echo"
type RewritePlugin struct {
	setPath     string
	stripPrefix string
	addPrefix   string
	addHeaders  map[string]string
	setUpstream string
}

func (p *RewritePlugin) Name() string { return "rewrite" }

func (p *RewritePlugin) Init(cfg map[string]any) error {
	if v, ok := cfg["set_path"].(string); ok {
		p.setPath = v
	}
	if v, ok := cfg["strip_prefix"].(string); ok {
		p.stripPrefix = v
	}
	if v, ok := cfg["add_prefix"].(string); ok {
		p.addPrefix = v
	}
	if v, ok := cfg["set_upstream"].(string); ok {
		p.setUpstream = v
	}
	if ah, ok := cfg["add_headers"].(map[string]any); ok {
		p.addHeaders = map[string]string{}
		for k, vi := range ah {
			if vs, ok := vi.(string); ok {
				p.addHeaders[k] = vs
			}
		}
	}
	return nil
}

func (p *RewritePlugin) BeforeDispatch(ctx *RequestContext) (bool, error) {
	r := ctx.Request
	// path rewrites
	path := r.URL.Path
	if p.stripPrefix != "" && strings.HasPrefix(path, p.stripPrefix) {
		path = strings.TrimPrefix(path, p.stripPrefix)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
	}
	if p.addPrefix != "" {
		ap := p.addPrefix
		if !strings.HasPrefix(ap, "/") {
			ap = "/" + ap
		}
		if strings.HasSuffix(ap, "/") && strings.HasPrefix(path, "/") {
			path = ap + strings.TrimPrefix(path, "/")
		} else if !strings.HasSuffix(ap, "/") && !strings.HasPrefix(path, "/") {
			path = ap + "/" + path
		} else {
			path = ap + path
		}
	}
	if p.setPath != "" {
		path = strings.ReplaceAll(p.setPath, "${path}", path)
	}
	ctx.Request.URL.Path = path

	// header injection
	for k, v := range p.addHeaders {
		ctx.Request.Header.Set(k, v)
	}

	// stash chosen upstream in context for router to consult
	if p.setUpstream != "" {
		ctx.Request = ctx.Request.WithContext(withUpstreamOverride(ctx.Request.Context(), p.setUpstream))
	}

	return false, nil
}

func (p *RewritePlugin) AfterDispatch(ctx *RequestContext) { /* no-op */ _ = fmt.Sprintf }

func init() { Register("rewrite", func() Plugin { return &RewritePlugin{} }) }
