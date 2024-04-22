package codegen

import "testing"

// Run 2nd
func TestGenerateTypescriptFile(t *testing.T) {

	GenerateTypescriptFile("docs.json", "public_structs.json", "./")

}

// Run 1st
func TestGenerateStructs(t *testing.T) {

	ExtractStructs("../", "./")

}
