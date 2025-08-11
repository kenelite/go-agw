package plugin

// Simple compile-time registry for built-in plugins

type Constructor func() Plugin

var registry = map[string]Constructor{}

func Register(name string, ctor Constructor) { registry[name] = ctor }

func getConstructor(name string) Constructor { return registry[name] }

