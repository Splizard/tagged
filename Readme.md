# Tagged Unions for Go

Always wanted to use tagged unions in Go? Thought it wasn't supported? well, it is!

```go

import "github.com/splizard/tagged"

// Each field in the struct, denotes a possible value for the union.
// A Float can either be a Bits32 (float32) or a Bits64 (float64).
// Up to 65535 fields are supported, of any type, make sure to use
// an appropriate buffer size/value, otherwise [tagged.Fields] will
// panic. If any value contains pointers, then the buffer type must
// be [any].
type Float tagged.Union[[8]byte, struct{
    Bits32 tagged.As[Float, float32]
    Bits64 tagged.As[Float, float64]
}]

// FloatWith is the name of our accessor value. The heavy reflection
// is completed in advance and cached in this value so that we can
// work with [Float] union values efficiently.
var FloatWith = tagged.Fields(Float{})

func main() {
    var pi = FloatWith.Bits32.New(math.Pi)
    var e = FloatWith.Bits64.New(math.E)
    fmt.Println(pi, e)

    switch tagged.FieldOf(pi) {
    case FloatWith.Bits32.Field:
        fmt.Println("pi is a float32")
    case FloatWith.Bits64.Field:
        fmt.Println("pi is a float64")
    }
}
```