package util

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestDetectImageFormatAndDimensions_StandardPNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 20))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	w, h, format, err := DetectImageFormatAndDimensions(buf.Bytes(), "http://example.com/test.png")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w != 10 || h != 20 {
		t.Errorf("expected size 10x20, got %dx%d", w, h)
	}
	if format != "png" {
		t.Errorf("expected png, got '%s'", format)
	}
}

func TestDetectImageFormatAndDimensions_StandardJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 15, 25))
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		t.Fatalf("failed to encode jpeg: %v", err)
	}

	w, h, format, err := DetectImageFormatAndDimensions(buf.Bytes(), "http://example.com/test.jpg")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w != 15 || h != 25 {
		t.Errorf("expected size 15x25, got %dx%d", w, h)
	}
	if format != "jpeg" {
		t.Errorf("expected jpeg, got '%s'", format)
	}
}

func TestDetectImageFormatAndDimensions_FallbackJPEG(t *testing.T) {
	data := []byte{
		0xFF, 0xD8,
		0xFF, 0xE0, 0x00, 0x04, 0x00, 0x00,
		0xFF, 0xC0,
		0x00, 0x0B,
		0x08,
		0x01, 0x00,
		0x02, 0x00,
		0x03,
		0x01, 0x02, 0x11,
	}

	_, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil {
		t.Fatalf("expected image.DecodeConfig to fail")
	}

	w, h, format, err := DetectImageFormatAndDimensions(data, "http://example.com/mock.jpg")
	if err != nil {
		t.Fatalf("expected fallback to succeed, but got error: %v", err)
	}
	if w != 512 || h != 256 {
		t.Errorf("expected fallback parsed size 512x256, got %dx%d", w, h)
	}
	if format != "jpeg" {
		t.Errorf("expected jpeg, got '%s'", format)
	}
}

func TestDetectImageFormatAndDimensions_FallbackURLGuessing(t *testing.T) {
	data := []byte("not-an-image")

	w, h, format, err := DetectImageFormatAndDimensions(data, "http://example.com/chapter-1/page-2.webp")
	if err != nil {
		t.Fatalf("Expected fallback URL guessing to succeed, but got error: %v", err)
	}
	if w != 0 || h != 0 {
		t.Errorf("Expected dimensions 0x0 for guessed format, got %dx%d", w, h)
	}
	if format != "webp" {
		t.Errorf("Expected webp, got '%s'", format)
	}
}

func TestDetectImageFormatAndDimensions_AVIF(t *testing.T) {
	// ftyp box with major brand avif
	data := []byte{
		0x00, 0x00, 0x00, 0x1c, // size 28
		'f', 't', 'y', 'p', // ftyp
		'a', 'v', 'i', 'f', // major brand
		0x00, 0x00, 0x00, 0x00, // minor version
		'a', 'v', 'i', 'f', // compatible brand
		// custom ispe box
		0x00, 0x00, 0x00, 0x14, // box size 20
		'i', 's', 'p', 'e', // ispe
		0x00, 0x00, 0x00, 0x00, // version & flags
		0x00, 0x00, 0x04, 0x00, // width 1024
		0x00, 0x00, 0x03, 0x00, // height 768
	}

	w, h, format, err := DetectImageFormatAndDimensions(data, "http://example.com/test.avif")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if format != "avif" {
		t.Errorf("expected avif, got '%s'", format)
	}
	if w != 1024 || h != 768 {
		t.Errorf("expected 1024x768, got %dx%d", w, h)
	}
}
