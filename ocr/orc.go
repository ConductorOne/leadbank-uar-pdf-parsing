package ocr

import (
	"bytes"
	"context"
	"image"
	"image/png"

	"github.com/otiai10/gosseract/v2"
)

type OCR interface {
	Prepare(ctx context.Context) error
	Extract(ctx context.Context, img image.Image) (string, error)
	Teardown(ctx context.Context) error
}

type Tesseract struct {
	tesseract *gosseract.Client
}

func (t *Tesseract) Prepare(ctx context.Context) error {
	t.tesseract = gosseract.NewClient()

	return nil
}

func (t *Tesseract) Teardown(ctx context.Context) error {
	return t.tesseract.Close()
}

func (t *Tesseract) Extract(ctx context.Context, img image.Image) (string, error) {
	// convert to bytes
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		panic(err)
	}

	// set the image
	t.tesseract.SetImageFromBytes(buf.Bytes())

	// perform OCR
	text, err := t.tesseract.Text()
	if err != nil {
		panic(err)
	}
	return text, nil

	// // save each cropped image for verification
	// imgOutputPath := path.Join(outputPath, imgName, fmt.Sprintf("%s_%d.png", key, zoneIndex))
	// outZoneFile, err := os.Create(imgOutputPath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer outZoneFile.Close()

	// // save the cropped image
	// err = png.Encode(outZoneFile, croppedImg)
	// if err != nil {
	// 	panic(err)
	// }
}
