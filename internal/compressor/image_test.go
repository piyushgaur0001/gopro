package compressor

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestCompressImageJPEG(t *testing.T) {
	input := testJPEG(t, 40, 30)

	result, err := CompressImage(input, ImageOptions{Quality: 75, MaxDimension: 100})
	if err != nil {
		t.Fatalf("CompressImage returned error: %v", err)
	}

	if result.ContentType != "image/jpeg" {
		t.Fatalf("got content type %q, want image/jpeg", result.ContentType)
	}
	if result.Extension != ".jpg" {
		t.Fatalf("got extension %q, want .jpg", result.Extension)
	}
	if len(result.Data) == 0 {
		t.Fatal("compressed data is empty")
	}
}

func TestCompressImagePNG(t *testing.T) {
	input := testPNG(t, 40, 30)

	result, err := CompressImage(input, ImageOptions{Quality: 80, MaxDimension: 100})
	if err != nil {
		t.Fatalf("CompressImage returned error: %v", err)
	}

	if result.ContentType != "image/png" {
		t.Fatalf("got content type %q, want image/png", result.ContentType)
	}
	if result.Extension != ".png" {
		t.Fatalf("got extension %q, want .png", result.Extension)
	}
	if len(result.Data) == 0 {
		t.Fatal("compressed data is empty")
	}
}

func TestCompressImageJPEGTargetSize(t *testing.T) {
	input := testJPEG(t, 240, 180)

	large, err := CompressImage(input, ImageOptions{Quality: 95, MaxDimension: 500})
	if err != nil {
		t.Fatalf("CompressImage returned error: %v", err)
	}

	target := int64(len(large.Data) / 2)
	result, err := CompressImage(input, ImageOptions{Quality: 95, TargetBytes: target, MaxDimension: 500})
	if err != nil {
		t.Fatalf("CompressImage returned error: %v", err)
	}

	if result.QualityUsed < 1 || result.QualityUsed > 100 {
		t.Fatalf("quality used %d is out of range", result.QualityUsed)
	}
	if len(result.Data) > len(large.Data) {
		t.Fatalf("target result size %d should not exceed high quality size %d", len(result.Data), len(large.Data))
	}
}

func TestCompressImageJPEGTinyTargetResizes(t *testing.T) {
	input := testJPEG(t, 1200, 800)

	result, err := CompressImage(input, ImageOptions{Quality: 95, TargetBytes: 1024, MaxDimension: 2000})
	if err != nil {
		t.Fatalf("CompressImage returned error: %v", err)
	}

	if int64(len(result.Data)) > 1024 {
		t.Fatalf("got %d bytes, want at most 1024 bytes", len(result.Data))
	}
}

func testJPEG(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, testImage(width, height), &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := png.Encode(&buf, testImage(width, height)); err != nil {
		t.Fatalf("encode png fixture: %v", err)
	}
	return buf.Bytes()
}

func testImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 120, A: 255})
		}
	}
	return img
}
