package logs

import (
	"context"
	"strings"
)

// FreeformEntry is a freeform log entry.
type FreeformEntry map[string]any

func newFreeformEntry() *FreeformEntry { return &FreeformEntry{} }

type keyValue struct {
	Key   string
	Value any
}

func (kv keyValue) adjust(e FreeformEntry, adj func(map[string]any, string)) {
	split := strings.Split(kv.Key, ".")

	current := e
	for i, sub := range split {
		if i == len(split)-1 {
			adj(current, sub)
			break
		}

		if _, ok := current[sub]; !ok {
			current[sub] = FreeformEntry{}
		}

		if nested, ok := current[sub].(map[string]any); ok {
			current = nested
		} else {
			m := make(map[string]any)
			current[sub] = m
			current = m
		}
	}
}

func (kv keyValue) overwrite(e FreeformEntry) {
	kv.adjust(e, func(m map[string]any, k string) {
		m[k] = kv.Value
	})
}

type keyValues []keyValue

func (k keyValues) overwrite(e FreeformEntry) {
	for _, kv := range k {
		kv.overwrite(e)
	}
}

func toKeyValues(args ...any) keyValues {
	kvs := make([]keyValue, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			kvs[i/2] = keyValue{Key: key, Value: args[i+1]}
		}
	}
	return kvs
}

// GetFreeformEntry retrieves the freeform log entry from the context. The
// function will return nil if no freeform log entry is found in the context.
func GetFreeformEntry(ctx context.Context) *FreeformEntry {
	return GetEntry[FreeformEntry](ctx)
}

// AdjustFreeform adjusts the freeform log entry in the context. The
// function will return false if no freeform log entry is found in the context.
func AdjustFreeform(ctx context.Context, fns ...func(*FreeformEntry)) bool {
	adj := make([]Adjuster[FreeformEntry], len(fns))
	for i, fn := range fns {
		adj[i] = func(e *FreeformEntry) { fn(e) }
	}
	return Adjust(ctx, adj...)
}

// Add adds key-value pairs to a freeform log entry. The function will return
// false if no freeform log entry is found in the context.
func Add(ctx context.Context, args ...any) bool {
	if e := GetFreeformEntry(ctx); e != nil {
		toKeyValues(args...).overwrite(*e)
		return true
	}

	return false
}

// Append adds values to an existing key of the freeform log entry in the
// context. If the key does not exist, it will be created. The function will
// return false if no freeform log entry is found in the context, or if the key
// exists but its value is not []T.
func Append[T any](ctx context.Context, key string, values ...T) bool {
	adjusted := false

	Adjust(ctx, func(e *FreeformEntry) {
		kv := keyValue{Key: key, Value: values}

		kv.adjust(*e, func(m map[string]any, k string) {
			if _, exists := m[k]; !exists {
				m[k] = values
				adjusted = true
			} else if existing, ok := m[k].([]T); ok {
				m[k] = append(existing, values...)
				adjusted = true
			}
		})
	})

	return adjusted
}
