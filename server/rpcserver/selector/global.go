package selector

var globalSelector Builder

// GlobalSelector 返回全局选择器构建器。
func GlobalSelector() Builder {
	return globalSelector
}

// SetGlobalSelector 设置全局选择器构建器。
func SetGlobalSelector(builder Builder) {
	globalSelector = builder
}
