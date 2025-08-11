package plugin

import "context"

type ctxKey string

const upstreamOverrideKey ctxKey = "plugin.rewrite.upstream_override"

func withUpstreamOverride(ctx context.Context, name string) context.Context {
    return context.WithValue(ctx, upstreamOverrideKey, name)
}

func UpstreamOverrideFrom(ctx context.Context) (string, bool) {
    v := ctx.Value(upstreamOverrideKey)
    s, ok := v.(string)
    return s, ok
}

