package router

import (
    "net/http"
)

// isGRPC determines whether the request is a gRPC call.
// We check HTTP/2 and the canonical content-type prefix.
func isGRPC(r *http.Request) bool {
    if r.ProtoMajor < 2 { return false }
    ct := r.Header.Get("Content-Type")
    if ct == "" { return false }
    // gRPC content types are like: application/grpc, application/grpc+proto, application/grpc+json
    return len(ct) >= 16 && ct[:16] == "application/grpc"
}

