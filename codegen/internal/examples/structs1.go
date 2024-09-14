package codegen

type Struct1 struct {
	Struct2
}

type Struct2 struct {
	Text string `json:"text"`
}
