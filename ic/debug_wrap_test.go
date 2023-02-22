package ic

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDebugWrap(t *testing.T) {
	var nilStringer *isStringer
	var nilPtrStringer *isPtrStringer
	var nilDebugStringer *isDebugStringer
	var nilPtrDebugStringer *isPtrDebugStringer
	realStringer := &isStringer{"test"}
	realPtrStringer := &isPtrStringer{"test"}
	realDebugStringer := &isDebugStringer{"test"}
	realPtrDebugStringer := &isPtrDebugStringer{"test"}
	var nilMap map[string]int
	var nilSlice []int
	var nilChan chan int
	tests := []struct {
		name string
		val  any
		want wrapper
	}{
		{"String", "test", wrapper(`"test"`)},
		{"Int", 1, wrapper(`1`)},
		{"Stringer", isStringer{"foo"}, wrapper(`isStringer("foo")`)},
		{"Wrapper", wrapper("foo"), wrapper("foo")},
		{"Real Stringer", realStringer, wrapper(`isStringer("test")`)},
		{"Real Ptr Stringer", realPtrStringer, wrapper(`isStringer("test")`)},
		{"Real DebugStringer", realDebugStringer, wrapper(`isDebugStringer("test")`)},
		{"Real Ptr DebugStringer", realPtrDebugStringer, wrapper(`isPtrDebugStringer("test")`)},
		{"Nil", nil, wrapper("")},
		{"Nil Stringer", nilStringer, wrapper("")},
		{"Nil Ptr Stringer", nilPtrStringer, wrapper("")},
		{"Nil DebugStringer", nilDebugStringer, wrapper("")},
		{"Nil Ptr DebugStringer", nilPtrDebugStringer, wrapper("")},
		{"Nil Map", nilMap, wrapper("")},
		{"Nil Chan", nilChan, wrapper("")},
		{"Nil Slice", nilSlice, wrapper("")}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DebugWrap(tt.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DebugWrap() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDebugNil(t *testing.T) {
	var nilStringer *isStringer
	var nilPtrStringer *isPtrStringer
	var nilDebugStringer *isDebugStringer
	var nilPtrDebugStringer *isPtrDebugStringer
	realStringer := &isStringer{"test"}
	realPtrStringer := &isPtrStringer{"test"}
	realDebugStringer := &isDebugStringer{"test"}
	realPtrDebugStringer := &isPtrDebugStringer{"test"}

	var nilMap map[string]int
	var nilSlice []int
	var nilChan chan int
	tests := []struct {
		name string
		val  any
		want wrapper
	}{
		{"String", "test", wrapper(`"test"`)},
		{"Int", 1, wrapper(`1`)},
		{"Stringer", isStringer{"foo"}, wrapper(`isStringer("foo")`)},
		{"Wrapper", wrapper("foo"), wrapper("foo")},
		{"Real Stringer", realStringer, wrapper(`isStringer("test")`)},
		{"Real Ptr Stringer", realPtrStringer, wrapper(`isStringer("test")`)},
		{"Real DebugStringer", realDebugStringer, wrapper(`isDebugStringer("test")`)},
		{"Real Ptr DebugStringer", realPtrDebugStringer, wrapper(`isPtrDebugStringer("test")`)},
		{"Nil", nil, wrapper("<nil>")},
		{"Nil Stringer", nilStringer, wrapper("<nil>")},
		{"Nil Ptr Stringer", nilPtrStringer, wrapper("<nil>")},
		{"Nil DebugStringer", nilDebugStringer, wrapper("<nil>")},
		{"Nil Ptr DebugStringer", nilPtrDebugStringer, wrapper("<nil>")},
		{"Nil Map", nilMap, wrapper("<nil>")},
		{"Nil Chan", nilChan, wrapper("<nil>")},
		{"Nil Slice", nilSlice, wrapper("<nil>")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DebugWrapNil(tt.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DebugWrapNil() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

type isStringer struct {
	s string
}

func (str isStringer) String() string {
	return fmt.Sprintf("isStringer(%q)", str.s)
}

type isPtrStringer struct {
	s string
}

func (str *isPtrStringer) String() string {
	return fmt.Sprintf("isStringer(%q)", str.s)
}

type isDebugStringer struct {
	s string
}

func (i isDebugStringer) DebugString() string {
	return fmt.Sprintf("isDebugStringer(%q)", i.s)
}

type isPtrDebugStringer struct {
	s string
}

func (i *isPtrDebugStringer) DebugString() string {
	return fmt.Sprintf("isPtrDebugStringer(%q)", i.s)
}
