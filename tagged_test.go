package tagged_test

import (
	"fmt"
	"math"
	"testing"

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

func TestTagged(t *testing.T) {
	var pi Float
	pi = FloatWith.Bits32.New(math.Pi)
	check(pi)
	pi = FloatWith.Bits64.New(math.Pi)
	check(pi)
}
