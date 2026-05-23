package compressor

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

type ImageResult struct {
	Data        []byte
	Format      string
	ContentType string
	Extension   string
	QualityUsed int
	TargetBytes int64
}

type ImageOptions struct {
	Quality      int
	TargetBytes  int64
	MaxDimension int
}

func CompressImage(input []byte, options ImageOptions) (ImageResult, error) {
	img, format, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return ImageResult{}, err
	}

	if options.MaxDimension > 0 {
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		if width > options.MaxDimension || height > options.MaxDimension {
			img = imaging.Fit(img, options.MaxDimension, options.MaxDimension, imaging.Lanczos)
		}
	}

	var out bytes.Buffer
	switch format {
	case "jpeg":
		quality := options.Quality
		if options.TargetBytes > 0 {
			encoded, usedQuality, err := encodeJPEGForTarget(img, options.TargetBytes)
			if err != nil {
				return ImageResult{}, err
			}
			out.Write(encoded)
			quality = usedQuality
		} else if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: quality}); err != nil {
			return ImageResult{}, err
		}
		return ImageResult{
			Data:        out.Bytes(),
			Format:      "jpeg",
			ContentType: "image/jpeg",
			Extension:   ".jpg",
			QualityUsed: quality,
			TargetBytes: options.TargetBytes,
		}, nil
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&out, img); err != nil {
			return ImageResult{}, err
		}
		return ImageResult{
			Data:        out.Bytes(),
			Format:      "png",
			ContentType: "image/png",
			Extension:   ".png",
			QualityUsed: options.Quality,
			TargetBytes: options.TargetBytes,
		}, nil
	default:
		return ImageResult{}, errors.New("unsupported image format")
	}
}

func encodeJPEGForTarget(img image.Image, targetBytes int64) ([]byte, int, error) {
	low := 1
	high := 100
	bestQuality := 1
	var bestUnder []byte
	var smallest []byte
	smallestQuality := 1

	for low <= high {
		quality := (low + high) / 2
		encoded, err := encodeJPEG(img, quality)
		if err != nil {
			return nil, 0, err
		}

		if smallest == nil || len(encoded) < len(smallest) {
			smallest = encoded
			smallestQuality = quality
		}

		if int64(len(encoded)) <= targetBytes {
			bestUnder = encoded
			bestQuality = quality
			low = quality + 1
		} else {
			high = quality - 1
		}
	}

	if bestUnder != nil {
		return bestUnder, bestQuality, nil
	}
	return smallest, smallestQuality, nil
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
