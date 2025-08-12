package plugin

import (
    "context"
    "time"
)

type ctxKey string

const upstreamOverrideKey ctxKey = "plugin.rewrite.upstream_override"
const startTimeKey ctxKey = "plugin.obs.start_time"

func withUpstreamOverride(ctx context.Context, name string) context.Context {
    return context.WithValue(ctx, upstreamOverrideKey, name)
}

func UpstreamOverrideFrom(ctx context.Context) (string, bool) {
    v := ctx.Value(upstreamOverrideKey)
    s, ok := v.(string)
    return s, ok
}

func withStartTime(ctx context.Context, t time.Time) context.Context { return context.WithValue(ctx, startTimeKey, t) }
func startTimeFrom(ctx context.Context) time.Time {
    if v := ctx.Value(startTimeKey); v != nil {
        if t, ok := v.(time.Time); ok { return t }
    }
    return time.Time{}
}

