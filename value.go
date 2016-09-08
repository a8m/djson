package djson

// ValueType identifies the type of a parsed value.
type ValueType int

// Type returns itself and provides an easy default implementation
// for embedding in a Value.
func (t ValueType) Type() ValueType {
	return t
}

const (
	Null ValueType = iota
	Bool
	String
	Number
	Object
	Array
	Unknown
)

type Value struct {
	data interface{}
}

func (v *Value) Data() interface{} {
	return v.data
}
