package ic

import (
	"fmt"
	"reflect"
)

type wrapper string

func (w wrapper) DebugString() string {
	return string(w)
}

func DebugWrap(val any) DebugStringer {
	var s string
	if !isNil(val) {
		if w, ok := val.(wrapper); ok {
			return w
		} else if debugStringer, ok := val.(DebugStringer); ok && debugStringer != nil {
			s = debugStringer.DebugString()
		} else if stringer, ok := val.(fmt.Stringer); ok && stringer != nil {
			s = stringer.String()
		} else if err, ok := val.(error); ok && err != nil {
			s = err.Error()
		} else {
			s = fmt.Sprintf("%#v", val)
		}
	}
	return wrapper(s)
}

func DebugWrapNil(val any) DebugStringer {
	if isNil(val) {
		return wrapper("<nil>")
	} else {
		return DebugWrap(val)
	}
}

func DebugWrapString(val string) DebugStringer {
	return wrapper(val)
}

func isNil(val any) bool {
	if val == nil {
		return true
	}
	switch reflect.TypeOf(val).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice, reflect.Func, reflect.Interface:
		return reflect.ValueOf(val).IsNil()
	}
	return false
}
