package technique

import "sync"

// registry 保存 executor 名称到技术实现的映射。
type registry struct {
	mu              sync.RWMutex
	implementations map[string]Technique
}

// global 是进程级默认注册表。
var global = &registry{implementations: make(map[string]Technique)}

// Register 注册 executor 对应的技术实现。
func Register(executor string, implementation Technique) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.implementations[executor] = implementation
}

// Get 按 executor 名称获取技术实现。
func Get(executor string) (Technique, bool) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	implementation, found := global.implementations[executor]
	return implementation, found
}
