/*
Package tagged provides type-safe generic high-performance unions for Go.

Try it out https://go.dev/play/p/Pp06ahQrt5-

	package main

	import (
		"fmt"
		"math"

		"github.com/splizard/tagged"
	)

	// Each field in the struct, denotes a possible value for the union.
	// A Float can either be a Bits32 (float32) or a Bits64 (float64).
	// Up to 65535 fields are supported, of any type, make sure to use
	// an appropriate buffer size/value, otherwise [tagged.Fields] will
	// panic. If any value contains pointers, then the buffer type must
	// be [any].
	type Float tagged.Union[[8]byte, struct {
		Bits32 tagged.As[Float, float32]
		Bits64 tagged.As[Float, float64]
	}]

	// FloatWith is the name of our accessor value. The heavy reflection
	// is completed in advance and cached in this value so that we can
	// work with [Float] union values efficiently.
	var FloatWith = tagged.Fields(Float{})

	func check(value Float) {
		switch tagged.FieldOf(value) {
		case FloatWith.Bits32.Field:
			var f32 float32 = FloatWith.Bits32.Get(value)
			fmt.Println("value is a float32", f32)
		case FloatWith.Bits64.Field:
			var f64 float64 = FloatWith.Bits64.Get(value)
			fmt.Println("value is a float64", f64)
		}
	}

	func main() {
		var pi Float
		pi = FloatWith.Bits32.New(math.Pi)
		check(pi)
		pi = FloatWith.Bits64.New(math.Pi)
		check(pi)
	}
*/
package tagged

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

type buffer interface {
	any | [4]byte | [8]byte | [12]byte | [16]byte | [24]byte | [32]byte | [64]byte | [128]byte | [256]byte | [512]byte | [1024]byte
}

// Union is used to create a new tagged union type. Select a buffer size and define
// the fields of the union using a struct. Each field should be of the [As] type.
//
// For example:
//
//	type Float tagged.Union[[8]byte, struct{
//		Bits32 tagged.As[Float, float32]
//		Bits64 tagged.As[Float, float64]
//	}]
type Union[Buf buffer, Values any] struct {
	UnionMethods[Buf, Values]
}

// Field identifies a particular field within a specified
// union type.
type Field[Union any] struct {
	_   *[0]Union
	tag int16
	set uintptr
	any bool
}

// FieldOf returns the currently tagged field in the given union value.
func FieldOf[Buf buffer, Values any, Union isUnion[Buf, Values]](union Union) Field[Union] {
	tag := union.getTag()
	return Field[Union]{
		tag: tag.tag,
		set: tag.set,
		any: tag.any,
	}
}

// Fields returns an accessor value that can be used to create and access
// the fields of a union value.
//
// For example:
//
//	type Float tagged.Union[[8]byte, struct{
//		Bits32 tagged.As[Float, float32]
//		Bits64 tagged.As[Float, float64]
//	}]
//
//	 var FloatUnion = tagged.Fields(Float{})
//
// Up to 65535 fields are supported, of any type, make sure to use
// an appropriate buffer size/value, otherwise [Fields] will
// panic. If any field contains pointers, then the buffer type must
// be [any].
func Fields[Buf buffer, Values any, Union isUnion[Buf, Values]](union Union) Values {
	return union.values()
}

type isUnion[Buf buffer, Values any] interface {
	~struct {
		UnionMethods[Buf, Values]
	}
	getTag() Field[struct{}]
	values() Values
}

// As is used to define a field within a tagged union struct.
type As[Union any, Value any] struct {
	Field[Union]
}

// New returns a new value of the tagged union type, with the specified field set to the given value.
func (f As[Union, Value]) New(value Value) Union {
	if f.set == 0 {
		panic("tagged.Field must be initialized before use")
	}
	if f.any {
		var union UnionMethods[any, struct{}]
		union.tag = f.tag
		union.buf = value
		union.any = true
		return *(*Union)(unsafe.Pointer(&union))
	}
	var union UnionMethods[Value, Value]
	union.tag = f.tag
	*(*Value)(unsafe.Add(unsafe.Pointer(&union), f.set)) = value
	return *(*Union)(unsafe.Pointer(&union))
}

// Get returns the value of the specified field in the given union value. If the
// field is not set, Get will panic.
func (f As[Union, Value]) Get(union Union) Value {
	value, ok := f.Lookup(union)
	if !ok {
		panic("tagged.Field.Get called with wrong tag")
	}
	return value
}

// Lookup returns the value of the specified field in the given union value. If the
// field is not set, Lookup will return the zero value for the field type and false.
func (f As[Union, Value]) Lookup(union Union) (Value, bool) {
	if f.set == 0 {
		panic("tagged.Field must be initialized before use")
	}
	var zero Value
	if f.any {
		var typed = *(*UnionMethods[any, struct{}])(reflect.ValueOf(&union).UnsafePointer())
		if typed.tag != f.tag {
			return zero, false
		}
		return typed.buf.(Value), true
	}
	var typed = *(*UnionMethods[Value, struct{}])(reflect.ValueOf(&union).UnsafePointer())
	if typed.tag != f.tag {
		return zero, false
	}
	return *(*Value)(unsafe.Add(unsafe.Pointer(&union), f.set)), true
}

func (field *As[Union, Value]) load(tag int16, direct bool, buf, offset uintptr) {
	var union Union
	var value Value
	if hasPointers(reflect.TypeOf(value)) && direct {
		panic(fmt.Sprintf("cannot load %T into [%d]byte %T, contains pointers", value, buf, union))
	}
	if direct && unsafe.Sizeof(value) > buf {
		panic(fmt.Sprintf("cannot load %T into [%d]byte %T, buffer too small", value, buf, union))
	}
	field.tag = tag
	field.set = offset
	field.any = !direct
}

func (field *As[Union, Value]) get(ptr unsafe.Pointer) any {
	return field.Get(*(*Union)(ptr))
}

func hasPointers(value reflect.Type) bool {
	switch value.Kind() {
	case reflect.Ptr, reflect.Chan, reflect.Map, reflect.Interface, reflect.Slice, reflect.Func, reflect.UnsafePointer:
		return true
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if hasPointers(value.Field(i).Type) {
				return true
			}
		}
	case reflect.Array:
		return hasPointers(value.Elem())

	}
	return false
}

// UnionMethods are the exported methods for a tagged union type.
// included and exported for documentation purposes only.
type UnionMethods[Buf any, Values any] struct {
	tag int16
	any bool
	buf Buf
}

func (union UnionMethods[Buf, Values]) getTag() Field[struct{}] {
	return Field[struct{}]{
		tag: union.tag,
		set: unsafe.Offsetof(union.buf),
		any: union.any,
	}
}

type loadable interface {
	load(tag int16, direct bool, buf, offset uintptr)
}

type gettable interface {
	loadable
	get(unsafe.Pointer) any
}

func (union UnionMethods[Buf, Values]) values() Values {
	var buffer Buf
	var values Values
	var rvalue = reflect.ValueOf(&values).Elem()
	var direct = reflect.TypeOf(buffer).Kind() == reflect.Array
	for i := 0; i < rvalue.NumField(); i++ {
		if i > math.MaxInt32 {
			panic(fmt.Sprintf("too many fields in %T", values))
		}
		rvalue.Field(i).Addr().Interface().(loadable).load(int16(i), direct, unsafe.Sizeof(buffer), unsafe.Offsetof(union.buf))
	}
	return values
}

// Interface returns the value of the tagged union as an any value.
func (union UnionMethods[Buf, Values]) Interface() any {
	var buffer Buf
	var values Values
	var rvalue = reflect.ValueOf(&values).Elem()
	var direct = reflect.TypeOf(buffer).Kind() == reflect.Array
	i := union.tag
	getter := rvalue.Field(int(i)).Addr().Interface().(gettable)
	getter.load(int16(i), direct, unsafe.Sizeof(buffer), unsafe.Offsetof(union.buf))
	return getter.get(unsafe.Pointer(&union))
}

// String implements the fmt.Stringer interface.
func (union UnionMethods[Buf, Values]) String() string {
	return fmt.Sprint(union.Interface())
}
