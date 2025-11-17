package matroska

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"testing"
)

// TestReadVInt tests the readVInt function with various inputs.
func TestReadVInt(t *testing.T) {
	testCases := []struct {
		name             string
		input            []byte
		keepLengthMarker bool
		expectedVal      uint64
		expectErr        bool
	}{
		// 1-byte VINTs
		{"1-byte value", []byte{0x81}, false, 1, false},
		{"1-byte max value", []byte{0xFF}, false, 127, false},
		{"1-byte with length marker", []byte{0x81}, true, 0x81, false},

		// 2-byte VINTs
		{"2-byte value", []byte{0x40, 0x01}, false, 1, false},
		{"2-byte value high", []byte{0x50, 0x11}, false, 0x1011, false},
		{"2-byte max value", []byte{0x7F, 0xFF}, false, (1 << 14) - 1, false},
		{"2-byte with length marker", []byte{0x50, 0x11}, true, 0x5011, false},

		// 4-byte VINTs
		{"4-byte value", []byte{0x10, 0x00, 0x00, 0x01}, false, 1, false},
		{"4-byte value high", []byte{0x1A, 0xBC, 0xDE, 0xF0}, false, 0xABCDEF0, false},
		{"4-byte max value", []byte{0x1F, 0xFF, 0xFF, 0xFF}, false, (1 << 28) - 1, false},
		{"4-byte with length marker", []byte{0x1A, 0xBC, 0xDE, 0xF0}, true, 0x1ABCDEF0, false},

		// 8-byte VINTs
		{"8-byte value", []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, false, 1, false},
		{"8-byte value high", []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}, false, 0x23456789ABCDEF, false},
		{"8-byte max value", []byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, false, (1 << 56) - 1, false},
		{"8-byte with length marker", []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}, true, 0x0123456789ABCDEF, false},

		// 5-byte VINTs
		{"5-byte value", []byte{0x08, 0x00, 0x00, 0x00, 0x01}, false, 1, false},
		{"5-byte with length marker", []byte{0x08, 0x00, 0x00, 0x00, 0x01}, true, 0x0800000001, false},

		// 6-byte VINTs
		{"6-byte value", []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x01}, false, 1, false},
		{"6-byte with length marker", []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x01}, true, 0x040000000001, false},

		// 7-byte VINTs
		{"7-byte value", []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, false, 1, false},
		{"7-byte with length marker", []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, true, 0x02000000000001, false},

		// Error cases
		{"invalid VINT zero byte", []byte{0x00}, false, 0, true},
		{"EOF in second byte", []byte{0x40}, false, 0, true},
		{"EOF in later byte", []byte{0x10, 0x00}, false, 0, true},
		{"invalid VINT no length marker", []byte{0x00, 0x00, 0x00}, true, 0, true},
		{"empty reader", []byte{}, false, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.input)
			reader := NewEBMLReader(r)

			val, err := reader.readVInt(tc.keepLengthMarker)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if val != tc.expectedVal {
					t.Errorf("Expected value %d, but got %d", tc.expectedVal, val)
				}
			}
		})
	}
}

// TestEBMLElementRead_Types tests the type reading methods of EBMLElement.
func TestEBMLElementRead_Types(t *testing.T) {
	t.Run("ReadUInt", func(t *testing.T) {
		testCases := []struct {
			name     string
			data     []byte
			expected uint64
		}{
			{"1 byte", []byte{0x01}, 1},
			{"2 bytes", []byte{0x01, 0x02}, 0x0102},
			{"4 bytes", []byte{0x01, 0x02, 0x03, 0x04}, 0x01020304},
			{"8 bytes", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, 0x0102030405060708},
			{"empty data", []byte{}, 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				el := &EBMLElement{Data: tc.data}
				if val := el.ReadUInt(); val != tc.expected {
					t.Errorf("ReadUInt() = %v, want %v", val, tc.expected)
				}
			})
		}
	})

	t.Run("ReadInt", func(t *testing.T) {
		testCases := []struct {
			name     string
			data     []byte
			expected int64
		}{
			{"1 byte positive", []byte{0x01}, 1},
			{"1 byte negative", []byte{0xFF}, -1},
			{"2 bytes positive", []byte{0x01, 0x02}, 0x0102},
			{"2 bytes negative", []byte{0xFF, 0xFE}, -2},
			{"4 bytes positive", []byte{0x01, 0x02, 0x03, 0x04}, 0x01020304},
			{"4 bytes negative", []byte{0xFF, 0xFF, 0xFF, 0xFE}, -2},
			{"8 bytes positive", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, 0x0102030405060708},
			{"8 bytes negative", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}, -2},
			{"empty data", []byte{}, 0},
			{"3 bytes negative", []byte{0xFF, 0xFF, 0xFE}, -2},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				el := &EBMLElement{Data: tc.data}
				if val := el.ReadInt(); val != tc.expected {
					t.Errorf("ReadInt() = %v, want %v", val, tc.expected)
				}
			})
		}
	})

	t.Run("ReadFloat", func(t *testing.T) {
		testCases := []struct {
			name     string
			data     []byte
			expected float64
			isNaN    bool
		}{
			{"32-bit", float32ToBytes(3.14), 3.140000104904175, false},
			{"64-bit", float64ToBytes(3.1415926535), 3.1415926535, false},
			{"32-bit NaN", float32ToBytes(float32(math.NaN())), 0, true},
			{"64-bit NaN", float64ToBytes(math.NaN()), 0, true},
			{"32-bit +Inf", float32ToBytes(float32(math.Inf(1))), math.Inf(1), false},
			{"64-bit +Inf", float64ToBytes(math.Inf(1)), math.Inf(1), false},
			{"32-bit -Inf", float32ToBytes(float32(math.Inf(-1))), math.Inf(-1), false},
			{"64-bit -Inf", float64ToBytes(math.Inf(-1)), math.Inf(-1), false},
			{"empty data", []byte{}, 0, false},
			{"invalid size", []byte{1, 2, 3}, 0, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				el := &EBMLElement{Data: tc.data}
				val := el.ReadFloat()
				if tc.isNaN {
					if !math.IsNaN(val) {
						t.Errorf("Expected NaN, but got %v", val)
					}
				} else if math.Abs(val-tc.expected) > 1e-9 {
					t.Errorf("ReadFloat() = %v, want %v", val, tc.expected)
				}
			})
		}
	})

	t.Run("ReadString", func(t *testing.T) {
		el := &EBMLElement{Data: []byte("hello")}
		if val := el.ReadString(); val != "hello" {
			t.Errorf("ReadString() = %q, want %q", val, "hello")
		}
		// With null terminator
		elNull := &EBMLElement{Data: []byte("hello\x00")}
		if val := elNull.ReadString(); val != "hello" {
			t.Errorf("ReadString() with null = %q, want %q", val, "hello")
		}
	})

	t.Run("ReadBytes", func(t *testing.T) {
		data := []byte{1, 2, 3}
		el := &EBMLElement{Data: data}
		if val := el.ReadBytes(); !reflect.DeepEqual(val, data) {
			t.Errorf("ReadBytes() = %v, want %v", val, data)
		}
	})
}

func float32ToBytes(f float32) []byte {
	bits := math.Float32bits(f)
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, bits)
	return b
}

func float64ToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, bits)
	return b
}

// TestEBMLReader_ReadElement tests reading a full element.
func TestEBMLReader_ReadElement(t *testing.T) {
	// ID: 0x1A45DFA3 (EBMLHeader), Size: 4, Data: "test"
	input := []byte{0x1A, 0x45, 0xDF, 0xA3, 0x84, 't', 'e', 's', 't'}
	r := bytes.NewReader(input)
	reader := NewEBMLReader(r)

	el, err := reader.ReadElement()
	if err != nil {
		t.Fatalf("ReadElement() failed: %v", err)
	}

	if el.ID != IDEBMLHeader {
		t.Errorf("Expected ID 0x%X, got 0x%X", IDEBMLHeader, el.ID)
	}
	if el.Size != 4 {
		t.Errorf("Expected size 4, got %d", el.Size)
	}
	if string(el.Data) != "test" {
		t.Errorf("Expected data 'test', got %q", string(el.Data))
	}
}

// TestEBMLReader_ReadEBMLHeader tests parsing the EBML header.
func TestEBMLReader_ReadEBMLHeader(t *testing.T) {
	t.Run("Valid EBML header", func(t *testing.T) {
		// EBMLHeader (ID 0x1A45DFA3)
		//   - EBMLVersion (ID 0x4286), Size 1, Value 1
		//   - DocType (ID 0x4282), Size 8, Value "matroska"
		headerData := []byte{
			0x42, 0x86, 0x81, 0x01, // EBMLVersion
			0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a', // DocType
		}
		headerSize := len(headerData)

		buf := new(bytes.Buffer)
		// Write EBML Header ID
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		// Write size
		buf.Write([]byte{byte(0x80 | headerSize)})
		// Write data
		buf.Write(headerData)

		r := bytes.NewReader(buf.Bytes())
		reader := NewEBMLReader(r)

		header, err := reader.ReadEBMLHeader()
		if err != nil {
			t.Fatalf("ReadEBMLHeader() failed: %v", err)
		}
		if header.Version != 1 {
			t.Errorf("Expected Version 1, got %d", header.Version)
		}
		if header.DocType != "matroska" {
			t.Errorf("Expected DocType 'matroska', got %q", header.DocType)
		}
	})

	t.Run("Complete EBML header with all fields", func(t *testing.T) {
		// EBMLHeader with all possible fields
		headerData := []byte{
			0x42, 0x86, 0x81, 0x01, // EBMLVersion = 1
			0x42, 0xF7, 0x81, 0x01, // EBMLReadVersion = 1 (ID: 0x42F7)
			0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a', // DocType = "matroska"
			0x42, 0x87, 0x81, 0x04, // DocTypeVersion = 4 (ID: 0x4287)
			0x42, 0x85, 0x81, 0x02, // DocTypeReadVersion = 2 (ID: 0x4285)
			0x42, 0xF2, 0x81, 0x04, // EBMLMaxIDLength = 4
			0x42, 0xF3, 0x81, 0x08, // EBMLMaxSizeLength = 8
		}
		headerSize := len(headerData)

		buf := new(bytes.Buffer)
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write([]byte{byte(0x80 | headerSize)})
		buf.Write(headerData)

		r := bytes.NewReader(buf.Bytes())
		reader := NewEBMLReader(r)

		header, err := reader.ReadEBMLHeader()
		if err != nil {
			t.Fatalf("ReadEBMLHeader() failed: %v", err)
		}
		if header.Version != 1 {
			t.Errorf("Expected Version 1, got %d", header.Version)
		}
		if header.ReadVersion != 1 {
			t.Errorf("Expected ReadVersion 1, got %d", header.ReadVersion)
		}
		if header.DocType != "matroska" {
			t.Errorf("Expected DocType 'matroska', got %q", header.DocType)
		}
		if header.DocTypeVersion != 4 {
			t.Errorf("Expected DocTypeVersion 4, got %d", header.DocTypeVersion)
		}
		if header.DocTypeReadVersion != 2 {
			t.Errorf("Expected DocTypeReadVersion 2, got %d", header.DocTypeReadVersion)
		}
		if header.MaxIDLength != 4 {
			t.Errorf("Expected MaxIDLength 4, got %d", header.MaxIDLength)
		}
		if header.MaxSizeLength != 8 {
			t.Errorf("Expected MaxSizeLength 8, got %d", header.MaxSizeLength)
		}
	})

	t.Run("Wrong element ID", func(t *testing.T) {
		// Wrong ID (not EBML header)
		input := []byte{0x42, 0x86, 0x81, 0x01} // EBMLVersion element instead of EBMLHeader
		r := bytes.NewReader(input)
		reader := NewEBMLReader(r)

		_, err := reader.ReadEBMLHeader()
		if err == nil {
			t.Errorf("Expected error for wrong element ID, but got nil")
		}
	})

	t.Run("EOF while reading header", func(t *testing.T) {
		r := bytes.NewReader([]byte{})
		reader := NewEBMLReader(r)

		_, err := reader.ReadEBMLHeader()
		if err == nil {
			t.Errorf("Expected error for EOF, but got nil")
		}
	})

	t.Run("Corrupted child element", func(t *testing.T) {
		// EBMLHeader with corrupted child element
		headerData := []byte{
			0x42, 0x86, 0x81, 0x01, // EBMLVersion
			0x00, 0x00, 0x81, 0x01, // Invalid child element ID (all zeros)
		}
		headerSize := len(headerData)

		buf := new(bytes.Buffer)
		buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
		buf.Write([]byte{byte(0x80 | headerSize)})
		buf.Write(headerData)

		r := bytes.NewReader(buf.Bytes())
		reader := NewEBMLReader(r)

		_, err := reader.ReadEBMLHeader()
		if err == nil {
			t.Errorf("Expected error for corrupted child element, but got nil")
		}
	})
}

func TestEBMLReader_ReadElementHeader(t *testing.T) {
	t.Run("Valid header", func(t *testing.T) {
		input := []byte{0x1A, 0x45, 0xDF, 0xA3, 0x84, 't', 'e', 's', 't'}
		r := bytes.NewReader(input)
		reader := NewEBMLReader(r)

		id, size, err := reader.ReadElementHeader()
		if err != nil {
			t.Fatalf("ReadElementHeader() failed: %v", err)
		}

		if id != IDEBMLHeader {
			t.Errorf("Expected ID 0x%X, got 0x%X", IDEBMLHeader, id)
		}
		if size != 4 {
			t.Errorf("Expected size 4, got %d", size)
		}

		// Check that we can read the rest of the data
		data := make([]byte, size)
		n, err := io.ReadFull(reader.r, data)
		if err != nil {
			t.Fatalf("Failed to read data after header: %v", err)
		}
		if n != 4 {
			t.Errorf("Expected to read 4 bytes, got %d", n)
		}
		if string(data) != "test" {
			t.Errorf("Expected data 'test', got %q", string(data))
		}
	})

	t.Run("EOF while reading ID", func(t *testing.T) {
		r := bytes.NewReader([]byte{})
		reader := NewEBMLReader(r)
		_, _, err := reader.ReadElementHeader()
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got %v", err)
		}
	})

	t.Run("EOF while reading size", func(t *testing.T) {
		r := bytes.NewReader([]byte{0x1A, 0x45, 0xDF, 0xA3})
		reader := NewEBMLReader(r)
		_, _, err := reader.ReadElementHeader()
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got %v", err)
		}
	})

	t.Run("Failing reader on ID", func(t *testing.T) {
		r := &failingSeeker{bytes.NewReader([]byte{0x1A, 0x45, 0xDF, 0xA3}), true}
		reader := NewEBMLReader(r)
		_, _, err := reader.ReadElementHeader()
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
	})

	t.Run("Failing reader on size", func(t *testing.T) {
		// Create a reader that will fail when trying to read the size
		r := &failingReader{data: []byte{0x1A, 0x45, 0xDF, 0xA3, 0x84}, failAtByte: 4}
		reader := NewEBMLReader(r)
		_, _, err := reader.ReadElementHeader()
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
	})

	t.Run("Invalid VInt ID", func(t *testing.T) {
		// Test with invalid VInt ID (all zeros)
		r := bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x00})
		reader := NewEBMLReader(r)
		_, _, err := reader.ReadElementHeader()
		if err == nil {
			t.Errorf("Expected an error for invalid VInt ID, but got nil")
		}
	})

	t.Run("Invalid VInt size", func(t *testing.T) {
		// Valid ID but invalid size (all zeros)
		r := bytes.NewReader([]byte{0x1A, 0x45, 0xDF, 0xA3, 0x00, 0x00})
		reader := NewEBMLReader(r)
		_, _, err := reader.ReadElementHeader()
		if err == nil {
			t.Errorf("Expected an error for invalid VInt size, but got nil")
		}
	})
}

func TestEBMLReader_SkipElement(t *testing.T) {
	t.Run("Skip known size", func(t *testing.T) {
		input := []byte{
			// First element: ID: 0x4286, Size: 1, Data: 1
			0x42, 0x86, 0x81, 0x01,
			// Second element: ID: 0x4282, Size: 8, Data: "matroska"
			0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a',
		}
		r := bytes.NewReader(input)
		reader := NewEBMLReader(r)

		// Read header of the first element
		id1, size1, err := reader.ReadElementHeader()
		if err != nil {
			t.Fatalf("Failed to read first element header: %v", err)
		}

		// Skip the data of the first element
		el1ToSkip := &EBMLElement{ID: id1, Size: size1}
		err = reader.SkipElement(el1ToSkip)
		if err != nil {
			t.Fatalf("SkipElement() failed: %v", err)
		}

		// Now, the reader should be at the start of the second element.
		// Let's read the second element to verify.
		el2, err := reader.ReadElement()
		if err != nil {
			t.Fatalf("Failed to read second element after skip: %v", err)
		}
		if el2.ID != IDEBMLDocType { // 0x4282
			t.Errorf("Expected second element ID 0x%X, got 0x%X", IDEBMLDocType, el2.ID)
		}
		if el2.ReadString() != "matroska" {
			t.Errorf("Expected second element data 'matroska', got %q", el2.ReadString())
		}
	})

	t.Run("Skip with failing seeker", func(t *testing.T) {
		input := []byte{0x42, 0x86, 0x81, 0x01}
		r := &failingSeeker{bytes.NewReader(input), true}
		reader := NewEBMLReader(r)

		el := &EBMLElement{ID: 0x4286, Size: 1}
		err := reader.SkipElement(el)
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("Expected io.ErrUnexpectedEOF, got %v", err)
		}
	})
}

func TestEBMLReader_ReadElement_Advanced(t *testing.T) {
	t.Run("Read different types", func(t *testing.T) {
		input := []byte{
			// UInt element
			0x42, 0x86, 0x81, 0x01,
			// String element
			0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'o', 's', 'k', 'a',
		}
		r := bytes.NewReader(input)
		reader := NewEBMLReader(r)

		// First element (UInt)
		el1, err1 := reader.ReadElement()
		if err1 != nil {
			t.Fatalf("Failed to read first element: %v", err1)
		}
		if el1.ID != IDEBMLVersion {
			t.Errorf("Expected ID 0x%X, got 0x%X", IDEBMLVersion, el1.ID)
		}
		if el1.ReadUInt() != 1 {
			t.Errorf("Expected uint value 1, got %d", el1.ReadUInt())
		}

		// Second element (String)
		el2, err2 := reader.ReadElement()
		if err2 != nil {
			t.Fatalf("Failed to read second element: %v", err2)
		}
		if el2.ID != IDEBMLDocType {
			t.Errorf("Expected ID 0x%X, got 0x%X", IDEBMLDocType, el2.ID)
		}
		if el2.ReadString() != "matroska" {
			t.Errorf("Expected string value 'matroska', got %q", el2.ReadString())
		}
	})

	t.Run("EOF while reading ID", func(t *testing.T) {
		r := bytes.NewReader([]byte{})
		reader := NewEBMLReader(r)
		_, err := reader.ReadElement()
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got %v", err)
		}
	})

	t.Run("EOF while reading size", func(t *testing.T) {
		r := bytes.NewReader([]byte{0x1A, 0x45, 0xDF, 0xA3})
		reader := NewEBMLReader(r)
		_, err := reader.ReadElement()
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got %v", err)
		}
	})

	t.Run("EOF while reading data", func(t *testing.T) {
		r := bytes.NewReader([]byte{0x42, 0x86, 0x84, 0x01, 0x02, 0x03})
		reader := NewEBMLReader(r)
		_, err := reader.ReadElement()
		if err == nil || err.Error() != "failed to read element data: unexpected EOF" {
			t.Errorf("Expected unexpected EOF error, got %v", err)
		}
	})

	t.Run("Unknown size element", func(t *testing.T) {
		r := bytes.NewReader([]byte{0x42, 0x86, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
		reader := NewEBMLReader(r)
		_, err := reader.ReadElement()
		if err == nil || err.Error() != "unknown size elements not supported" {
			t.Errorf("Expected unknown size error, got %v", err)
		}
	})

	t.Run("Zero size element", func(t *testing.T) {
		r := bytes.NewReader([]byte{0x42, 0x86, 0x80})
		reader := NewEBMLReader(r)
		el, err := reader.ReadElement()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if el.ID != IDEBMLVersion {
			t.Errorf("Expected ID 0x%X, got 0x%X", IDEBMLVersion, el.ID)
		}
		if el.Size != 0 {
			t.Errorf("Expected size 0, got %d", el.Size)
		}
		if len(el.Data) != 0 {
			t.Errorf("Expected empty data, got %d bytes", len(el.Data))
		}
	})
}

type failingSeeker struct {
	*bytes.Reader
	failOnSeek bool
}

func (s *failingSeeker) Seek(offset int64, whence int) (int64, error) {
	if s.failOnSeek {
		return 0, io.ErrUnexpectedEOF
	}
	return s.Reader.Seek(offset, whence)
}

type failingReader struct {
	data       []byte
	pos        int
	failAtByte int
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.pos >= r.failAtByte {
		return 0, io.ErrUnexpectedEOF
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.pos:])
	if r.pos+n > r.failAtByte {
		n = r.failAtByte - r.pos
	}
	r.pos += n

	if r.pos >= r.failAtByte {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}

func (r *failingReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.pos = int(offset)
	case io.SeekCurrent:
		r.pos += int(offset)
	case io.SeekEnd:
		r.pos = len(r.data) + int(offset)
	}
	return int64(r.pos), nil
}

func TestEBMLReader_Seek(t *testing.T) {
	input := []byte("abcdefghijklmnopqrstuvwxyz")

	t.Run("Successful seek", func(t *testing.T) {
		r := bytes.NewReader(input)
		reader := NewEBMLReader(r)

		// Seek from start
		pos, err := reader.Seek(10, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek from start failed: %v", err)
		}
		if pos != 10 {
			t.Errorf("Expected position 10, got %d", pos)
		}
		if reader.Position() != 10 {
			t.Errorf("Expected internal position 10, got %d", reader.Position())
		}

		// Seek from current
		pos, err = reader.Seek(5, io.SeekCurrent)
		if err != nil {
			t.Fatalf("Seek from current failed: %v", err)
		}
		if pos != 15 {
			t.Errorf("Expected position 15, got %d", pos)
		}

		// Seek from end
		pos, err = reader.Seek(-5, io.SeekEnd)
		if err != nil {
			t.Fatalf("Seek from end failed: %v", err)
		}
		if pos != int64(len(input)-5) {
			t.Errorf("Expected position %d, got %d", len(input)-5, pos)
		}

		// Read after seek
		b := make([]byte, 5)
		_, err = reader.r.Read(b)
		if err != nil {
			t.Fatalf("Read after seek failed: %v", err)
		}
		if string(b) != "vwxyz" {
			t.Errorf("Expected to read 'vwxyz', got %q", string(b))
		}
	})

	t.Run("Failing seek", func(t *testing.T) {
		r := &failingSeeker{bytes.NewReader(input), true}
		reader := NewEBMLReader(r)

		_, err := reader.Seek(10, io.SeekStart)
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("Expected io.ErrUnexpectedEOF, got %v", err)
		}
	})
}
func TestSeekableReader_Seek(t *testing.T) {
	input := []byte("abcdefghijklmnopqrstuvwxyz")

	t.Run("Successful seek operations", func(t *testing.T) {
		r := bytes.NewReader(input)
		sr := &seekableReader{r}

		// Seek from start
		pos, err := sr.Seek(10, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek from start failed: %v", err)
		}
		if pos != 10 {
			t.Errorf("Expected position 10, got %d", pos)
		}

		// Seek from current
		pos, err = sr.Seek(5, io.SeekCurrent)
		if err != nil {
			t.Fatalf("Seek from current failed: %v", err)
		}
		if pos != 15 {
			t.Errorf("Expected position 15, got %d", pos)
		}

		// Seek from end
		pos, err = sr.Seek(-5, io.SeekEnd)
		if err != nil {
			t.Fatalf("Seek from end failed: %v", err)
		}
		if pos != int64(len(input)-5) {
			t.Errorf("Expected position %d, got %d", len(input)-5, pos)
		}

		// Read after seek to verify position
		b := make([]byte, 5)
		_, err = sr.Read(b)
		if err != nil {
			t.Fatalf("Read after seek failed: %v", err)
		}
		if string(b) != "vwxyz" {
			t.Errorf("Expected to read 'vwxyz', got %q", string(b))
		}
	})

	t.Run("Seek beyond bounds", func(t *testing.T) {
		r := bytes.NewReader(input)
		sr := &seekableReader{r}

		// Seek beyond end
		pos, err := sr.Seek(1000, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek beyond end failed: %v", err)
		}
		if pos != 1000 {
			t.Errorf("Expected position 1000, got %d", pos)
		}

		// Try to read - should get EOF
		b := make([]byte, 1)
		_, err = sr.Read(b)
		if err != io.EOF {
			t.Errorf("Expected EOF when reading beyond end, got %v", err)
		}
	})

	t.Run("Seek to negative position", func(t *testing.T) {
		r := bytes.NewReader(input)
		sr := &seekableReader{r}

		// Seek to negative position from start should fail
		_, err := sr.Seek(-10, io.SeekStart)
		if err == nil {
			t.Errorf("Expected error when seeking to negative position, but got nil")
		}
	})

	t.Run("Multiple seek operations", func(t *testing.T) {
		r := bytes.NewReader(input)
		sr := &seekableReader{r}

		// Test multiple seeks
		positions := []struct {
			offset int64
			whence int
			expect int64
		}{
			{0, io.SeekStart, 0},
			{10, io.SeekStart, 10},
			{5, io.SeekCurrent, 15},
			{-3, io.SeekCurrent, 12},
			{0, io.SeekEnd, int64(len(input))},
			{-1, io.SeekEnd, int64(len(input) - 1)},
		}

		for i, pos := range positions {
			result, err := sr.Seek(pos.offset, pos.whence)
			if err != nil {
				t.Fatalf("Seek operation %d failed: %v", i, err)
			}
			if result != pos.expect {
				t.Errorf("Seek operation %d: expected position %d, got %d", i, pos.expect, result)
			}
		}
	})
}

// Additional advanced tests for EBMLReader.ReadElement and edge cases
func TestEBMLReader_ReadElement_MoreCases(t *testing.T) {
	// Build a buffer with multiple elements of different types
	buf := new(bytes.Buffer)

	// Element 1: ID 0x1A45DFA3 (EBML header child-like), size 1, data 0x01 (uint)
	buf.Write([]byte{0x1A, 0x45, 0xDF, 0xA3})
	buf.Write(vintEncode(1))
	buf.Write([]byte{0x01})

	// Element 2: ID 0x4282 (DocType), size 6, data "abc123" (string)
	buf.Write([]byte{0x42, 0x82})
	buf.Write(vintEncode(6))
	buf.Write([]byte("abc123"))

	// Element 3: ID 0x4489 (Duration), size 4, float32 1.5
	buf.Write([]byte{0x44, 0x89})
	buf.Write(vintEncode(4))
	var f32 [4]byte
	binary.BigEndian.PutUint32(f32[:], math.Float32bits(1.5))
	buf.Write(f32[:])

	r := NewEBMLReader(bytes.NewReader(buf.Bytes()))
	el1, err := r.ReadElement()
	if err != nil {
		t.Fatalf("ReadElement 1 failed: %v", err)
	}
	if el1.ReadUInt() != 1 {
		t.Errorf("expected uint 1, got %d", el1.ReadUInt())
	}

	el2, err := r.ReadElement()
	if err != nil {
		t.Fatalf("ReadElement 2 failed: %v", err)
	}
	if el2.ReadString() != "abc123" {
		t.Errorf("expected string 'abc123', got %q", el2.ReadString())
	}

	el3, err := r.ReadElement()
	if err != nil {
		t.Fatalf("ReadElement 3 failed: %v", err)
	}
	if math.Abs(float64(el3.ReadFloat())-1.5) > 1e-6 {
		t.Errorf("expected float 1.5, got %f", el3.ReadFloat())
	}

	// Next should hit EOF cleanly
	if el4, errReadElement := r.ReadElement(); errReadElement != io.EOF || el4 != nil {
		t.Errorf("expected io.EOF, got el=%v err=%v", el4, errReadElement)
	}
}

func TestEBMLReader_ReadElement_UnknownSizeAndInvalidID(t *testing.T) {
	// Unknown size element should error
	buf := new(bytes.Buffer)
	// ID 0x42 0x82
	buf.Write([]byte{0x42, 0x82})
	// Write 8-byte unknown size VINT: 0x01 followed by seven 0xFF
	buf.Write([]byte{0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	// No data needed; ReadElement should reject unknown-size elements
	r := NewEBMLReader(bytes.NewReader(buf.Bytes()))
	if _, err := r.ReadElement(); err == nil {
		t.Fatalf("expected error for unknown size element, got nil")
	}

	// Invalid ID: first byte 0x00 should cause VINT error
	buf2 := bytes.NewBuffer([]byte{0x00})
	r2 := NewEBMLReader(bytes.NewReader(buf2.Bytes()))
	if _, err := r2.ReadElement(); err == nil {
		t.Fatalf("expected error for invalid ID VINT, got nil")
	}
}
