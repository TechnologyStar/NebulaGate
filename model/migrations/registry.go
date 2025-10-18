package migrations

import "sync"

type SchemaProvider func() []interface{}

var (
	schemaProviderMu sync.RWMutex
	schemaProviders  = map[string]SchemaProvider{}
)

func RegisterSchemaProvider(version string, provider SchemaProvider) {
	schemaProviderMu.Lock()
	defer schemaProviderMu.Unlock()
	schemaProviders[version] = provider
}

func schemaTables(version string) ([]interface{}, bool) {
	schemaProviderMu.RLock()
	defer schemaProviderMu.RUnlock()
	provider, ok := schemaProviders[version]
	if !ok || provider == nil {
		return nil, false
	}
	tables := provider()
	copied := make([]interface{}, len(tables))
	copy(copied, tables)
	return copied, true
}
