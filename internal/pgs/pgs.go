package pgs

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
)

// Segment Types
const (
	SegPDS = 0x14 // Palette Definition Segment
	SegODS = 0x15 // Object Definition Segment
	SegPCS = 0x16 // Presentation Composition Segment
	SegWDS = 0x17 // Window Definition Segment
	SegEND = 0x80 // End of Display Set
)

const (
	CompStateNormal           = 0x00 // Normal: New display
	CompStateAcquisitionPoint = 0x40 // Acquisition Point: New epoch start
	CompStateEpochStart       = 0x80 // Epoch Start: Palette update
	CompStateEpochContinue    = 0xC0 // Epoch Continue: Update existing
)

// PgsDecoder decodes PGS packets into images
type PgsDecoder struct {
	Palette            color.Palette
	currentObject      *PgsObject
	currentComposition *PgsComposition
	objects            map[uint16]*PgsObject // Store completed objects by ID
	windows            map[uint8]*WindowDefinition
}

type PgsObject struct {
	ID       uint16
	Version  uint8
	Width    uint16
	Height   uint16
	Data     []byte
	Complete bool
}

type PgsComposition struct {
	PTS              uint64 // Presentation timestamp
	DTS              uint64 // Decode timestamp
	Width            uint16
	Height           uint16
	FrameRate        uint8
	CompositionNum   uint16
	CompositionState uint8
	PaletteUpdate    bool
	PaletteID        uint8
	Objects          []CompositionObject
}

type CompositionObject struct {
	ObjectID   uint16
	WindowID   uint8
	X          uint16
	Y          uint16
	Cropped    bool
	CropX      uint16
	CropY      uint16
	CropWidth  uint16
	CropHeight uint16
}

type WindowDefinition struct {
	WindowID uint8
	X        uint16
	Y        uint16
	Width    uint16
	Height   uint16
}

func NewPgsDecoder() *PgsDecoder {
	palette := make(color.Palette, 256)
	// Initialize with transparent as default
	for i := range palette {
		palette[i] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	}
	return &PgsDecoder{
		Palette: palette,
		objects: make(map[uint16]*PgsObject),
		windows: make(map[uint8]*WindowDefinition),
	}
}

// DecodePacket processes a single raw PGS packet (which may contain multiple segments)
// Returns an image if this packet completed an object, otherwise nil.
func (d *PgsDecoder) DecodePacket(packet []byte) (image.Image, error) {
	var resultImg image.Image
	offset := 0
	segmentCount := 0

	// Process all segments in the packet
	for offset < len(packet) {
		if offset+3 > len(packet) {
			break
		}

		segType := packet[offset]
		segSize := binary.BigEndian.Uint16(packet[offset+1 : offset+3])

		if offset+3+int(segSize) > len(packet) {
			return nil, fmt.Errorf("packet too short: expected %d bytes at offset %d, got %d", segSize+3, offset, len(packet)-offset)
		}

		payload := packet[offset+3 : offset+3+int(segSize)]
		segmentCount++

		var img image.Image
		var err error

		switch segType {
		case SegPDS:
			err = d.parsePDS(payload)
		case SegODS:
			img, err = d.parseODS(payload)
			if img != nil {
				resultImg = img
			}
		case SegPCS:
			err = d.parsePCS(payload)
		case SegWDS:
			err = d.parseWDS(payload)
		case SegEND:
			// End of display set marker
		}

		if err != nil {
			return nil, fmt.Errorf("error processing segment type 0x%02X: %w", segType, err)
		}

		// Move to next segment
		offset += 3 + int(segSize)
	}

	return resultImg, nil
}

// parsePDS parses the YCbCr palette and converts it to Go RGBA
func (d *PgsDecoder) parsePDS(data []byte) error {
	// Header: ID(1) + Version(1)
	if len(data) < 2 {
		return errors.New("PDS too short")
	}

	// Entries start after header. Each entry is 5 bytes: ID, Y, Cr, Cb, Alpha
	for i := 2; i+5 <= len(data); i += 5 {
		id := data[i]
		y, cr, cb, a := data[i+1], data[i+2], data[i+3], data[i+4]

		// Convert YCbCr to RGB
		// Standard formulas (BT.601 or similar usually used in PGS)
		yf := float64(y) - 16
		crf := float64(cr) - 128
		cbf := float64(cb) - 128

		r := clamp(1.164*yf + 1.596*crf)
		g := clamp(1.164*yf - 0.813*crf - 0.391*cbf)
		b := clamp(1.164*yf + 2.018*cbf)

		d.Palette[id] = color.RGBA{R: r, G: g, B: b, A: a}
	}
	return nil
}

// parsePCS parses the Presentation Composition Segment
// Contains timing and positioning information
func (d *PgsDecoder) parsePCS(data []byte) error {
	// Header: Width(2) + Height(2) + FrameRate(1) + CompositionNum(2) + CompositionState(1) + PaletteUpdateFlag(1) + PaletteID(1) + NumCompositionObjects(1)
	if len(data) < 11 {
		return errors.New("PCS too short")
	}

	comp := &PgsComposition{
		Width:            binary.BigEndian.Uint16(data[0:2]),
		Height:           binary.BigEndian.Uint16(data[2:4]),
		FrameRate:        data[4],
		CompositionNum:   binary.BigEndian.Uint16(data[5:7]),
		CompositionState: data[7],
		PaletteUpdate:    data[8] == 0x80,
		PaletteID:        data[9],
	}

	numObjects := int(data[10])
	offset := 11

	// Parse composition objects
	for i := 0; i < numObjects; i++ {
		if offset+8 > len(data) {
			break
		}

		obj := CompositionObject{
			ObjectID: binary.BigEndian.Uint16(data[offset : offset+2]),
			WindowID: data[offset+2],
		}

		// Byte at offset+3 contains the cropping flag (bit 7)
		cropFlag := data[offset+3]
		obj.Cropped = (cropFlag & 0x80) != 0

		// X and Y come after the crop flag
		obj.X = binary.BigEndian.Uint16(data[offset+4 : offset+6])
		obj.Y = binary.BigEndian.Uint16(data[offset+6 : offset+8])

		if obj.Cropped {
			// Cropped object needs 8 more bytes for crop info
			if offset+16 > len(data) {
				break
			}
			obj.CropX = binary.BigEndian.Uint16(data[offset+8 : offset+10])
			obj.CropY = binary.BigEndian.Uint16(data[offset+10 : offset+12])
			obj.CropWidth = binary.BigEndian.Uint16(data[offset+12 : offset+14])
			obj.CropHeight = binary.BigEndian.Uint16(data[offset+14 : offset+16])
			offset += 16
		} else {
			// Non-cropped object is 8 bytes
			offset += 8
		}

		comp.Objects = append(comp.Objects, obj)
	}

	d.currentComposition = comp
	return nil
}

// parseWDS parses the Window Definition Segment
// Defines display windows where objects are rendered
func (d *PgsDecoder) parseWDS(data []byte) error {
	if len(data) < 1 {
		return errors.New("WDS too short")
	}

	numWindows := int(data[0])
	offset := 1

	// Each window definition is 9 bytes: WindowID(1) + X(2) + Y(2) + Width(2) + Height(2)
	for i := 0; i < numWindows; i++ {
		if offset+9 > len(data) {
			break
		}

		windowID := data[offset]
		x := binary.BigEndian.Uint16(data[offset+1 : offset+3])
		y := binary.BigEndian.Uint16(data[offset+3 : offset+5])
		width := binary.BigEndian.Uint16(data[offset+5 : offset+7])
		height := binary.BigEndian.Uint16(data[offset+7 : offset+9])

		d.windows[windowID] = &WindowDefinition{
			WindowID: windowID,
			X:        x,
			Y:        y,
			Width:    width,
			Height:   height,
		}

		offset += 9
	}

	return nil
}

// parseODS parses the RLE bitmap data
func (d *PgsDecoder) parseODS(data []byte) (image.Image, error) {
	// Header: ID(2) + Version(1) + Sequence(1) + DataLength(3) + Width(2) + Height(2)
	if len(data) < 11 {
		return nil, errors.New("ODS too short")
	}

	objectID := binary.BigEndian.Uint16(data[0:2])
	version := data[2]
	sequenceFlag := data[3]
	dataLength := binary.BigEndian.Uint32([]byte{0, data[4], data[5], data[6]})
	width := binary.BigEndian.Uint16(data[7:9])
	height := binary.BigEndian.Uint16(data[9:11])

	// RLE data starts after header
	rleData := data[11:]

	// Handle multi-segment objects
	if sequenceFlag&0x80 != 0 {
		// First segment, initialize object
		d.currentObject = &PgsObject{
			ID:       objectID,
			Version:  version,
			Width:    width,
			Height:   height,
			Data:     make([]byte, 0, dataLength),
			Complete: false,
		}
		d.currentObject.Data = append(d.currentObject.Data, rleData...)
	} else if sequenceFlag&0x40 != 0 {
		// Last segment, complete object
		if d.currentObject != nil && d.currentObject.ID == objectID {
			d.currentObject.Data = append(d.currentObject.Data, rleData...)
			d.currentObject.Complete = true

			// Decode complete object
			img := image.NewPaletted(
				image.Rect(0, 0, int(d.currentObject.Width), int(d.currentObject.Height)),
				d.Palette,
			)

			err := decodeRLE(d.currentObject.Data, img.Pix, int(d.currentObject.Width), int(d.currentObject.Height))
			if err != nil {
				return nil, fmt.Errorf("failed to decode RLE: %w", err)
			}

			// Store completed object for potential composition use
			d.objects[d.currentObject.ID] = d.currentObject
			d.currentObject = nil
			return img, nil
		}
	} else {
		// Middle segment - append data
		if d.currentObject != nil && d.currentObject.ID == objectID {
			d.currentObject.Data = append(d.currentObject.Data, rleData...)
		}
	}

	// Single segment object (0xC0)
	if sequenceFlag == 0xC0 {
		img := image.NewPaletted(image.Rect(0, 0, int(width), int(height)), d.Palette)
		err := decodeRLE(rleData, img.Pix, int(width), int(height))
		if err != nil {
			return nil, fmt.Errorf("failed to decode RLE: %w", err)
		}
		return img, nil
	}

	return nil, nil
}

// decodeRLE implements the specific Run-Length Encoding used in PGS
func decodeRLE(data []byte, pix []byte, width, height int) error {
	buf := bytes.NewReader(data)
	idx := 0
	limit := width * height

	for idx < limit && buf.Len() > 0 {
		b, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read byte at idx %d: %w", idx, err)
		}

		if b != 0 {
			// Non-zero value is a single pixel of that color index
			if idx >= len(pix) {
				return fmt.Errorf("pixel buffer overflow at idx %d", idx)
			}
			pix[idx] = b
			idx++
		} else {
			// Zero starts an escape sequence
			flag, err := buf.ReadByte()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to read flag byte at idx %d: %w", idx, err)
			}

			if flag == 0 {
				// 0x00 0x00 = End of Line (Fill rest of line with transparent)
				col := idx % width
				if col > 0 {
					lineEnd := idx + (width - col)
					if lineEnd > limit {
						lineEnd = limit
					}
					for idx < lineEnd {
						if idx >= len(pix) {
							return fmt.Errorf("pixel buffer overflow at idx %d during EOL", idx)
						}
						pix[idx] = 0
						idx++
					}
				}
			} else {
				// Parse Run Length and Color
				var runLength int
				var colorIndex byte

				if (flag & 0xC0) == 0 {
					// 0x00 0x0L -> 'L' transparent pixels
					runLength = int(flag & 0x3F)
					colorIndex = 0
				} else if (flag & 0xC0) == 0x40 {
					// 0x00 0x4H 0xLL -> Zeros with 2-byte length
					nextByte, err := buf.ReadByte()
					if err != nil {
						return fmt.Errorf("failed to read extended run length at idx %d: %w", idx, err)
					}
					runLength = (int(flag&0x3F) << 8) | int(nextByte)
					colorIndex = 0
				} else if (flag & 0xC0) == 0x80 {
					// 0x00 0x8L 0xCC -> 'L' pixels of color 'CC'
					runLength = int(flag & 0x3F)
					colorIndex, err = buf.ReadByte()
					if err != nil {
						return fmt.Errorf("failed to read color index at idx %d: %w", idx, err)
					}
				} else if (flag & 0xC0) == 0xC0 {
					// 0x00 0xCL 0xLL 0xCC -> 'LLL' pixels of color 'CC'
					nextByte, err := buf.ReadByte()
					if err != nil {
						return fmt.Errorf("failed to read run length at idx %d: %w", idx, err)
					}
					runLength = (int(flag&0x3F) << 8) | int(nextByte)
					colorIndex, err = buf.ReadByte()
					if err != nil {
						return fmt.Errorf("failed to read color index at idx %d: %w", idx, err)
					}
				}

				// Fill pixels with bounds checking
				endIdx := idx + runLength
				if endIdx > limit {
					endIdx = limit
				}
				for idx < endIdx {
					if idx >= len(pix) {
						return fmt.Errorf("pixel buffer overflow at idx %d during run fill", idx)
					}
					pix[idx] = colorIndex
					idx++
				}
			}
		}
	}

	return nil
}

func clamp(f float64) uint8 {
	if f < 0 {
		return 0
	}
	if f > 255 {
		return 255
	}
	return uint8(f)
}

// EncodePgsImageToBase64PNG encodes a PGS image to a base64-encoded PNG string
func EncodePgsImageToBase64PNG(img image.Image, compressionLevel png.CompressionLevel) (string, error) {
	if img == nil {
		return "", errors.New("image is nil")
	}

	var buf bytes.Buffer
	encoder := &png.Encoder{
		CompressionLevel: compressionLevel,
	}

	if err := encoder.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encoded, nil
}

// GetCurrentComposition returns the current composition information
// This includes positioning, canvas size, and object layout data
func (d *PgsDecoder) GetCurrentComposition() *PgsComposition {
	return d.currentComposition
}

// GetWindow returns the window definition for a given window ID
func (d *PgsDecoder) GetWindow(windowID uint8) *WindowDefinition {
	return d.windows[windowID]
}

// GetObject returns a stored object by its ID
func (d *PgsDecoder) GetObject(objectID uint16) *PgsObject {
	return d.objects[objectID]
}

// GetPgsSegmentTypeName returns the name of a PGS segment type for logging
func GetPgsSegmentTypeName(data []byte) string {
	if len(data) < 1 {
		return "UNKNOWN"
	}
	switch data[0] {
	case SegPDS:
		return "PDS (Palette Definition)"
	case SegODS:
		return "ODS (Object Definition)"
	case SegPCS:
		return "PCS (Presentation Composition)"
	case SegWDS:
		return "WDS (Window Definition)"
	case SegEND:
		return "END (End of Display Set)"
	default:
		return fmt.Sprintf("UNKNOWN (0x%02X)", data[0])
	}
}

// ListPgsSegments returns a list of all segment types in a PGS packet
func ListPgsSegments(packet []byte) []string {
	var segments []string
	offset := 0

	for offset < len(packet) {
		if offset+3 > len(packet) {
			break
		}

		segType := packet[offset]
		segSize := binary.BigEndian.Uint16(packet[offset+1 : offset+3])

		if offset+3+int(segSize) > len(packet) {
			break
		}

		var segName string
		switch segType {
		case SegPDS:
			segName = "PDS"
		case SegODS:
			segName = "ODS"
			// Add sequence flag info
			if offset+3+4 < len(packet) {
				seqFlag := packet[offset+3+3]
				if seqFlag == 0xC0 {
					segName += "(single)"
				} else if seqFlag&0x80 != 0 {
					segName += "(first)"
				} else if seqFlag&0x40 != 0 {
					segName += "(last)"
				} else {
					segName += "(middle)"
				}
			}
		case SegPCS:
			segName = "PCS"
		case SegWDS:
			segName = "WDS"
		case SegEND:
			segName = "END"
		default:
			segName = fmt.Sprintf("0x%02X", segType)
		}

		segments = append(segments, segName)
		offset += 3 + int(segSize)
	}

	return segments
}

// IsClearCommand checks if the current composition is a clear command
// A clear command is a PCS with Normal state (0x00) or Acquisition Point state (0x40) with no objects
func (d *PgsDecoder) IsClearCommand() bool {
	if d.currentComposition == nil {
		return false
	}

	comp := d.currentComposition
	// Clear command: Normal or Acquisition Point state with no objects to display
	if (comp.CompositionState == CompStateNormal || comp.CompositionState == CompStateAcquisitionPoint) && len(comp.Objects) == 0 {
		return true
	}

	return false
}

// GetCompositionState returns the current composition state, or -1 if no composition exists
func (d *PgsDecoder) GetCompositionState() int {
	if d.currentComposition == nil {
		return -1
	}
	return int(d.currentComposition.CompositionState)
}

// ClearCompositionState clears the current composition and associated state
// This should be called when a new display set begins
func (d *PgsDecoder) ClearCompositionState() {
	d.currentComposition = nil
}
