package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/otiai10/gosseract/v2"
)

func config() (string, string, int, error) {
	convertPath := os.Getenv("CONVERT_PATH")
	if convertPath == "" {
		return "", "", 0, fmt.Errorf("CONVERT_PATH is not set")
	}

	outputPath := os.Getenv("OUTPUT_PATH")
	if outputPath == "" {
		return "", "", 0, fmt.Errorf("OUTPUT_PATH is not set")
	}

	testPage := os.Getenv("TESTING_PAGE_NUMBER")
	if testPage == "" {
		return "", "", 0, fmt.Errorf("TESTING_PAGE_NUMBER is not set")
	}

	testPageNum, err := strconv.Atoi(testPage)
	if err != nil {
		return "", "", 0, fmt.Errorf("TESTING_PAGE_NUMBER is not a number")
	}

	return convertPath, outputPath, testPageNum, nil
}

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

func trimExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func extract(client *gosseract.Client, img image.Image, zone image.Rectangle, outputPath, imgName, key string, fileIndex, zoneIndex int) {
	// crop the image to the zone
	croppedImg := cropImage(img, zone)

	// convert to bytes
	buf := new(bytes.Buffer)
	err := png.Encode(buf, croppedImg)
	if err != nil {
		panic(err)
	}

	// set the image
	client.SetImageFromBytes(buf.Bytes())

	// perform OCR
	text, err := client.Text()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Extracted text (%s, %d): %s\n", key, zoneIndex, text)

	// save each cropped image for verification
	imgOutputPath := path.Join(outputPath, imgName, fmt.Sprintf("%s_%d.png", key, zoneIndex))
	outZoneFile, err := os.Create(imgOutputPath)
	if err != nil {
		panic(err)
	}
	defer outZoneFile.Close()

	// save the cropped image
	err = png.Encode(outZoneFile, croppedImg)
	if err != nil {
		panic(err)
	}
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

func main() {
	convertPath, outputPath, testPageNum, err := config()

	// read converted pdf
	files, err := os.ReadDir(convertPath)
	if err != nil {
		panic(err)
	}

	client := gosseract.NewClient()
	defer client.Close()

	for fileIndex, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".png" {
			continue
		}

		// open the image file (converted pdf page)
		imgName := trimExt(file.Name())
		imgPath := filepath.Join(convertPath, file.Name())
		fmt.Printf("Processing image: %s\n", imgPath)

		imgFile, err := os.Open(imgPath)
		if err != nil {
			panic(err)
		}
		defer imgFile.Close()

		// decode the image
		img, _, err := image.Decode(imgFile)
		if err != nil {
			panic(err)
		}

		// define zones (each page have identical structure)
		baseZones := map[string]image.Rectangle{
			"username": image.Rect(0, 150, 175, 173),
			"name":     image.Rect(0, 175, 280, 196),
			"email":    image.Rect(450, 175, 1440, 196),
			"groups":   image.Rect(180, 850, 1440, 1440),
		}
		permissionZones := map[string][]image.Rectangle{
			"PRM": getPermissionZones(32, 355),
			"CIS": getPermissionZones(174, 355),
			// TODO: add other permission zones
		}

		// make a folder for each page to visualize the zones and output the cropped images
		imgOutputPath := path.Join(outputPath, imgName)
		err = os.Mkdir(imgOutputPath, os.ModePerm)
		if err != nil {
			panic(err)
		}

		isTestPage := fileIndex == testPageNum

		// test the visualization of the zones on one page
		var testImage *image.RGBA
		if isTestPage {
			// copy the original image to draw the zones
			testImage = image.NewRGBA(img.Bounds())
			draw.Draw(testImage, img.Bounds(), img, image.Point{}, draw.Src)
		}

		// parse the base zones (username, email, ...)
		for key, zone := range baseZones {
			// extract text from the zone
			extract(client, img, zone, outputPath, imgName, key, fileIndex, 0)

			// visualize the zones on one image
			if isTestPage {
				// Draw each zone on the copied image
				red := color.RGBA{255, 0, 0, 255}
				draw.Draw(testImage, zone, &image.Uniform{red}, image.Point{}, draw.Over)
			}
		}

		// parse the permission zones (PRM, CIS, ...)
		for key, zones := range permissionZones {
			for zoneIndex, zone := range zones {
				// extract text from the zone
				extract(client, img, zone, outputPath, imgName, key, fileIndex, zoneIndex)

				// visualize the zones on one image
				if isTestPage {
					// Draw each zone on the copied image
					blue := color.RGBA{0, 0, 255, 255}
					draw.Draw(testImage, zone, &image.Uniform{blue}, image.Point{}, draw.Over)
				}
			}
		}

		if isTestPage {
			// Create or open the file where the visualized zones will be saved
			imgOutputPath := path.Join(outputPath, imgName, "output.png")
			vizFile, err := os.Create(imgOutputPath)
			if err != nil {
				panic(err)
			}
			defer vizFile.Close()

			err = png.Encode(vizFile, testImage)
			if err != nil {
				panic(err)
			}
		}
	}
}
