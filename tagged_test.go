package tagged_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/splizard/tagged"
)

type Float tagged.Union[[8]byte, struct {
	Bits32 tagged.As[Float, float32]
	Bits64 tagged.As[Float, float64]
}]

var FloatWith = tagged.Fields(Float{})

func TestTagged(t *testing.T) {
	var f = FloatWith.Bits32.New(math.Pi)
	fmt.Println(f)
	f = FloatWith.Bits64.New(math.Pi)
	fmt.Println(f)

	switch tagged.FieldOf(f) {
	case FloatWith.Bits32.Field:
		t.Fatal("expected 64-bit field")
	case FloatWith.Bits64.Field:
	default:
		t.Fatal("expected 64-bit field")
	}

	f = FloatWith.Bits32.New(math.E)

	switch f.Interface().(type) {
	case float32:

	case float64:
		t.Fatal("expected 32-bit field")
	}
}
