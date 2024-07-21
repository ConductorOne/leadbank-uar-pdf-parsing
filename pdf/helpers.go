package pdf

import (
	"image"
	"path/filepath"
)

func cropImage(img image.Image, rect image.Rectangle) image.Image {
	// Calculate the width and height of the rectangle
	width := rect.Dx()
	height := rect.Dy()

	// Create a new RGBA image with the correct dimensions
	croppedImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy pixels from the original image to the new cropped image
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			// Set the pixel at the new location in the cropped image
			// Adjusted to start from (0, 0)
			croppedImg.Set(x-rect.Min.X, y-rect.Min.Y, img.At(x, y))
		}
	}

	return croppedImg
}

// each permission zone is a rectangle with the same width and height
// the zones are placed in a grid with a fixed distance between them
// the distance between the zones is following:
// - the width of the zone is 10 pixels
// - the height of the zone is 40 pixels
func getPermissionZones(startX, startY int) []image.Rectangle {
	// the width and height of the zone
	width := 11
	height := 40

	// the number of zones in the grid
	cols := 12

	// the distance between the zones
	dx := width

	// the list of zones
	zones := make([]image.Rectangle, 0)

	// iterate over the rows and columns
	for col := 0; col < cols; col++ {
		// calculate the coordinates of the top-left corner of the zone
		x := startX + col*dx

		// create the zone rectangle
		zone := image.Rect(x, startY, x+width, startY+height)

		// add the zone to the list
		zones = append(zones, zone)
	}

	return zones
}

func trimExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
