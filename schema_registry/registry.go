package schema_registry

import (
	_ "embed"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"strconv"
	"strings"
)

var (
	//go:embed resources/auth/user-registered.1.json
	userRegisteredSchema_1 string

	//go:embed resources/auth/user-role-changed.1.json
	userRoleChangedSchema_1 string

	//go:embed resources/tasks/task-assigned.1.json
	taskAssignedSchema_1 string

	//go:embed resources/tasks/task-completed.1.json
	taskCompletedSchema_1 string

	//go:embed resources/tasks/task-created.1.json
	taskCreatedSchema_1 string

	//go:embed resources/tasks/task-created.2.json
	taskCreatedSchema_2 string

	globalRegistry = map[string]string{
		"auth.user-registered.1":   userRegisteredSchema_1,
		"auth.user-role-changed.1": userRoleChangedSchema_1,
		"tasks.task-created.1":     taskCreatedSchema_1,
		"tasks.task-created.2":     taskCreatedSchema_2,
		"tasks.task-completed.1":   taskCompletedSchema_1,
		"tasks.task-assigned.1":    taskAssignedSchema_1,
	}
)

type (
	Schemas struct {
		scopedRegistry map[string][]scopedRegistryEntry
	}

	Scope string

	scopedRegistryEntry struct {
		version int
		schema  string
	}
)

func NewSchemas(scope Scope, name ...string) (Schemas, error) {
	registry := make(map[string][]scopedRegistryEntry)

	for key, schema := range globalRegistry {
		parsed := strings.SplitN(key, ".", 3)
		regScope, regName, regVer := parsed[0], parsed[1], parsed[2]

		regVerInt, err := strconv.Atoi(regVer)
		if err != nil {
			return Schemas{}, fmt.Errorf(`invalid version in key=%s: %w`, key, err)
		}

		if string(scope) == regScope {
			registry[regName] = append(registry[regName], scopedRegistryEntry{
				version: regVerInt, schema: schema,
			})
		}
	}

	return Schemas{scopedRegistry: registry}, nil
}

func (ss *Schemas) Validate(event []byte, name string, version int) (bool, error) {
	schema, err := ss.findSchema(name, version)
	if err != nil {
		return false, err
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(schema),
		gojsonschema.NewBytesLoader(event),
	)
	if err != nil {
		return false, fmt.Errorf(`error during validation: %w`, err)
	}

	// TODO: in case of failed validation, `result` contains details which can be used for diagnostics
	return result.Valid(), nil
}

func (ss *Schemas) findSchema(name string, version int) (string, error) {
	entries, ok := ss.scopedRegistry[name]
	if !ok {
		return "", fmt.Errorf(`unregistered schema for name=%s`, name)
	}

	var schema string
	for _, entry := range entries {
		if entry.version == version {
			schema = entry.schema
			break
		}
	}

	if schema == "" {
		return "", fmt.Errorf(`no schema found for version=%s name=%s`, version, name)
	}

	return schema, nil
}
