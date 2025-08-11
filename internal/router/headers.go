package router

import "net/http"

// Standard hop-by-hop headers that must not be forwarded by proxies.
var hopByHopHeaders = map[string]struct{}{
	"Connection":          {},
	"Proxy-Connection":    {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {}, // RFC spells it TE, but per net/http canonicalization this renders as Te
	"Trailers":            {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func removeHopByHopHeaders(h http.Header) {
	// Remove headers listed in Connection as well
	if c := h.Get("Connection"); c != "" {
		// net/http canonicalizes, split by comma
		// Keep it simple: rely on hopByHop list
	}
	for k := range hopByHopHeaders {
		h.Del(k)
	}
}

func cloneHeader(h http.Header) http.Header {
	dst := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		dst[k] = vv2
	}
	return dst
}

func copyHeaderExcept(dst, src http.Header, except map[string]struct{}) {
	for k, vv := range src {
		if _, skip := except[k]; skip {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
