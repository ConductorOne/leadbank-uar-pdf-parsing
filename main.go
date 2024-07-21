package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/leadbank-uar-pdf-parsing/ocr"
	"github.com/conductorone/leadbank-uar-pdf-parsing/pdf"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Config struct {
	PDFPath     string
	ConvertPath string
	OutputPath  string
}

func loadConfig() (*Config, error) {
	pdfPath := os.Getenv("PDF_PATH")
	if pdfPath == "" {
		return nil, fmt.Errorf("PDF_PATH is not set")
	}

	convertPath := os.Getenv("CONVERT_PATH")
	if convertPath == "" {
		return nil, fmt.Errorf("CONVERT_PATH is not set")
	}

	outputPath := os.Getenv("OUTPUT_PATH")
	if outputPath == "" {
		return nil, fmt.Errorf("OUTPUT_PATH is not set")
	}

	return &Config{
		PDFPath:     pdfPath,
		ConvertPath: convertPath,
		OutputPath:  outputPath,
	}, nil
}

func main() {
	ctx := context.Background()
	l := ctxzap.Extract(ctx)

	l.Info("Preparing configuration")
	cfg, err := loadConfig()
	if err != nil {
		l.Error("Failed to load configuration", zap.Error(err))
		panic(err)
	}

	l.Info("Preparing the PDF")
	pdfWrapper := pdf.NewNativePDF(cfg.PDFPath, cfg.ConvertPath)
	err = pdfWrapper.Convert(ctx)
	if err != nil {
		l.Error("Failed to convert PDF", zap.Error(err))
		panic(err)
	}

	// prepare the OCR client
	ocrWrapper := ocr.Tesseract{}
	ocrWrapper.Prepare(ctx)
	defer ocrWrapper.Teardown(ctx)

	// prepare zones for parsing
	userBaseZones := pdf.NewUserBaseZones()
	permissionZones := pdf.NewPermissionZones()

	// parse the images and extract the text based on the zones
	l.Info("Extracting text from images")
	err = pdfWrapper.ParseImages(ctx, &ocrWrapper, userBaseZones, permissionZones)
	if err != nil {
		l.Error("Failed to parse images", zap.Error(err))
		panic(err)
	}

	// read username and his permissions from mapped zones
	l.Info("User Base Info", zap.String("username", userBaseZones.GetValue("username")), zap.String("email", userBaseZones.GetValue("email")))
	l.Info("Permission Zones PRM", zap.Any("PRM", permissionZones.GetValue("PRM")))
}
