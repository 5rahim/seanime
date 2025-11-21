package matroska

import (
	"bytes"
	"io"
	"testing"
)

// TestFakeSeeker tests the behavior of the fakeSeeker.
func TestFakeSeeker(t *testing.T) {
	data := []byte("hello world")
	r := bytes.NewReader(data)
	fs := &fakeSeeker{r: r}

	// Test Read
	t.Run("Read", func(t *testing.T) {
		buf := make([]byte, 5)
		n, err := fs.Read(buf)
		if err != nil {
			t.Fatalf("Read() failed: %v", err)
		}
		if n != 5 {
			t.Errorf("Expected to read 5 bytes, got %d", n)
		}
		if string(buf) != "hello" {
			t.Errorf("Expected to read 'hello', got %q", string(buf))
		}
	})

	// Test Read to EOF
	t.Run("Read_EOF", func(t *testing.T) {
		// Drain the rest of the reader
		_, _ = io.ReadAll(fs)

		buf := make([]byte, 1)
		n, err := fs.Read(buf)
		if err != io.EOF {
			t.Errorf("Expected EOF, got %v", err)
		}
		if n != 0 {
			t.Errorf("Expected to read 0 bytes at EOF, got %d", n)
		}
	})

	// Test Seek
	t.Run("Seek", func(t *testing.T) {
		pos, err := fs.Seek(0, io.SeekStart)
		if err == nil {
			t.Error("Seek() should always return an error")
		}
		if pos != -1 {
			t.Errorf("Seek() should return position -1 on error, got %d", pos)
		}
	})
}

// TestVintEncode_AllLengths verifies vintEncode produces correctly-sized VINTs
// across all supported length buckets (1..8 bytes), and that parseVInt can
// decode those values back to the original.
func TestVintEncode_AllLengths(t *testing.T) {
	cases := []struct {
		name   string
		value  uint64
		expLen int
	}{
		{"len1_max", 0x7F, 1},               // < 0x80
		{"len2_max", 0x3FFF, 2},             // < 0x4000
		{"len3_max", 0x1FFFFF, 3},           // < 0x200000
		{"len4_max", 0x0FFFFFFF, 4},         // < 0x10000000
		{"len5_max", 0x07FFFFFFFF, 5},       // < 0x800000000
		{"len6_max", 0x03FFFFFFFFFF, 6},     // < 0x40000000000
		{"len7_max", 0x01FFFFFFFFFFFF, 7},   // < 0x2000000000000
		{"len8_max", 0x00FFFFFFFFFFFFFF, 8}, // < 0x100000000000000
	}

	mp := &MatroskaParser{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := vintEncode(tc.value)
			if len(enc) != tc.expLen {
				t.Fatalf("vintEncode(%#x) length = %d, want %d", tc.value, len(enc), tc.expLen)
			}

			v, n := mp.parseVInt(enc)
			if n != tc.expLen {
				t.Fatalf("parseVInt length = %d, want %d (value=%#x)", n, tc.expLen, tc.value)
			}
			if v != tc.value {
				t.Fatalf("parseVInt value = %#x, want %#x", v, tc.value)
			}
		})
	}
}

// TestVintEncode_Length9 verifies we encode very large values using 9-byte VINTs
// (reserved leading 0x00 marker) and validates the bytes layout.
func TestVintEncode_Length9(t *testing.T) {
	val := uint64(0x100000000000000) // == 1<<56, first value requiring 9-byte encoding
	enc := vintEncode(val)
	if len(enc) != 9 {
		t.Fatalf("vintEncode length for %#x = %d, want 9", val, len(enc))
	}
	if enc[0] != 0x00 {
		t.Fatalf("vintEncode[0] = %#x, want 0x00", enc[0])
	}
	// Expect [0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00]
	expected := []byte{0x01, 0, 0, 0, 0, 0, 0, 0}
	for i, b := range expected {
		if enc[1+i] != b {
			t.Fatalf("vintEncode[%d] = %#x, want %#x (val=%#x)", 1+i, enc[1+i], b, val)
		}
	}
	// parseVInt cannot decode 9-byte (leading 0) by design; ensure it refuses
	mp := &MatroskaParser{}
	if v, n := mp.parseVInt(enc); n != 0 || v != 0 {
		t.Fatalf("parseVInt should fail for 9-byte encoding, got v=%#x n=%d", v, n)
	}
}
