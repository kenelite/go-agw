package plugin

func getStringOr(m map[string]any, key, def string) string {
    if v, ok := m[key].(string); ok && v != "" { return v }
    return def
}

