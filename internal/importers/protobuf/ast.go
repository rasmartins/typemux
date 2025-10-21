package protobuf

// ProtoSchema represents a parsed protobuf schema
type ProtoSchema struct {
	Syntax   string
	Package  string
	Imports  []string
	Options  map[string]string
	Messages []*ProtoMessage
	Enums    []*ProtoEnum
	Services []*ProtoService
}

// ProtoMessage represents a protobuf message
type ProtoMessage struct {
	Name     string
	Fields   []*ProtoField
	Enums    []*ProtoEnum
	Messages []*ProtoMessage // nested messages
	Options  map[string]string
	Reserved []string
	OneOfs   []*ProtoOneOf
}

// ProtoField represents a field in a message
type ProtoField struct {
	Name       string
	Type       string
	Number     int
	Repeated   bool
	Optional   bool
	Deprecated bool
	Comment    string
}

// ProtoEnum represents an enum
type ProtoEnum struct {
	Name   string
	Values []*ProtoEnumValue
}

// ProtoEnumValue represents an enum value
type ProtoEnumValue struct {
	Name   string
	Number int
}

// ProtoService represents a gRPC service
type ProtoService struct {
	Name    string
	Methods []*ProtoMethod
}

// ProtoMethod represents a service method
type ProtoMethod struct {
	Name         string
	InputType    string
	OutputType   string
	ClientStream bool
	ServerStream bool
}

// ProtoOneOf represents a oneof field
type ProtoOneOf struct {
	Name   string
	Fields []*ProtoField
}
