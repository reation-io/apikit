package tags

import "reflect"

// getTypeName returns the type name of an interface value
func getTypeName(v any) string {
	if v == nil {
		return "nil"
	}
	return reflect.TypeOf(v).String()
}

