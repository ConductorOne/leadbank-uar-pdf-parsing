package pdf

import (
	"context"
	"image"

	"github.com/conductorone/leadbank-uar-pdf-parsing/ocr"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Zones interface {
	Parse(ctx context.Context, ocr ocr.OCR, img image.Image) error
}

type UserBaseZones struct {
	zones  map[string]image.Rectangle
	values map[string]string
}

func NewUserBaseZones() *UserBaseZones {
	return &UserBaseZones{
		zones: map[string]image.Rectangle{
			"username": image.Rect(0, 150, 175, 173),
			"name":     image.Rect(0, 175, 280, 196),
			"email":    image.Rect(450, 175, 1440, 196),
			"groups":   image.Rect(180, 850, 1440, 1440),
		},
	}
}

func (ubz *UserBaseZones) Parse(ctx context.Context, ocr ocr.OCR, img image.Image) error {
	l := ctxzap.Extract(ctx)

	for key, zone := range ubz.zones {
		croppedImg := cropImage(img, zone)
		text, err := ocr.Extract(ctx, croppedImg)
		if err != nil {
			return err
		}

		l.Info("User info", zap.String(key, text))

		ubz.values[key] = text
	}

	return nil
}

func (ubz *UserBaseZones) GetValue(key string) string {
	return ubz.values[key]
}

type PermissionZones struct {
	zones  map[string][]image.Rectangle
	values map[string][]string
}

func NewPermissionZones() *PermissionZones {
	return &PermissionZones{
		zones: map[string][]image.Rectangle{
			"PRM": getPermissionZones(32, 355),
			"CIS": getPermissionZones(174, 355),
			// TODO: add other permission zones
		},
	}
}

func (pz *PermissionZones) Parse(ctx context.Context, ocr ocr.OCR, img image.Image) error {
	l := ctxzap.Extract(ctx)

	for key, zones := range pz.zones {
		for permissiomIndex, zone := range zones {
			croppedImg := cropImage(img, zone)
			text, err := ocr.Extract(ctx, croppedImg)
			if err != nil {
				return err
			}

			l.Info("Permission info", zap.String(key, text))

			pz.values[key][permissiomIndex] = text
		}
	}

	return nil
}

func (pz *PermissionZones) GetValue(key string) []string {
	return pz.values[key]
}
