package pdf

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/conductorone/leadbank-uar-pdf-parsing/ocr"
	"github.com/gen2brain/go-fitz"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// PDF interface for converting pdf to images
type PDF interface {
	// Convert converts the pdf to images
	Convert(ctx context.Context) error

	// GetImages returns the paths or instances of the images
	GetImages(ctx context.Context) []interface{}

	// ParseImages
	ParseImages(ctx context.Context, ocr ocr.OCR, zones ...Zones) error
}

type NativePDF struct {
	// Path to the pdf file
	Path string
	// Output directory for the images
	OutputDir string
}

var _ PDF = (*NativePDF)(nil)

func (p *NativePDF) Convert(ctx context.Context) error {
	l := ctxzap.Extract(ctx)

	doc, err := fitz.New(p.Path)
	if err != nil {
		return fmt.Errorf("failed to open pdf: %w", err)
	}
	defer doc.Close()

	// check for output directory and create if not exists
	// ideally in one simple function
	err = os.MkdirAll(p.OutputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// extract pages as images
	for i := range doc.NumPage() {
		img, err := doc.ImageDPI(i, 300)
		if err != nil {
			return fmt.Errorf("failed to extract image: %w", err)
		}

		imgPath := path.Join(p.OutputDir, fmt.Sprintf("page-%d.png", i))
		imgFile, err := os.Create(imgPath)
		if err != nil {
			return fmt.Errorf("failed to create image file: %w", err)
		}
		defer imgFile.Close()

		l.Info("Converting page to image", zap.Int("page", i), zap.String("path", imgPath))
		err = png.Encode(imgFile, img)
		if err != nil {
			return fmt.Errorf("failed to encode image: %w", err)
		}
	}

	return nil
}

func (p *NativePDF) GetImages(ctx context.Context) []interface{} {
	files, err := os.ReadDir(p.OutputDir)
	if err != nil {
		panic(err)
	}

	var images []interface{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".png" {
			continue
		}

		imgPath := filepath.Join(p.OutputDir, file.Name())
		images = append(images, imgPath)
	}

	return images
}

func (npdf *NativePDF) ParseImages(ctx context.Context, ocr ocr.OCR, zones ...Zones) error {
	l := ctxzap.Extract(ctx)

	paths := npdf.GetImages(ctx)
	for _, p := range paths {
		imagePath, ok := p.(string)
		if !ok {
			continue
		}

		// open the image file (converted pdf page)
		l.Info("Processing image", zap.String("image", imagePath))
		imgFile, err := os.Open(imagePath)
		if err != nil {
			return err
		}
		defer imgFile.Close()

		// decode the image
		img, _, err := image.Decode(imgFile)
		if err != nil {
			return err
		}

		// extract text from the image in the zone
		for _, zone := range zones {
			err = zone.Parse(ctx, ocr, img)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NewPDF creates a new PDF instance
func NewNativePDF(path, outputDir string) PDF {
	return &NativePDF{
		Path:      path,
		OutputDir: outputDir,
	}
}
