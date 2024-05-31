# leadbank-uar-pdf-parsing

## Prerequisites

This script requires:
- pdftoppm or other tool for converting pdf to images
- Tesseract OCR support (more information [here](https://tesseract-ocr.github.io/tessdoc/Compiling.html))
- Output directory for cropped images
- PDF file with UAR data

## Output

Output of this script are in current solution files with cropped images that represent parts of the UAR pdf file. This is for testing purposes only and should be removed in production.

You can also set testing page number with `TESTING_PAGE_NUMBER` environment variable that will use one page of the pdf file for testing purposes. This page is then used as visualisation of what parts of pdf page are being processed.

## Usage

To use this script, you have to specify following parameters:
- `PDF_PATH` - path to the pdf file with UAR data
- `CONVERT_PATH` - path to the directory where the pdf file will be converted to images
- `OUTPUT_PATH` - path to the output directory for cropped images
- (optional) `TESTING_PAGE_NUMBER` - page number of the pdf file that will be used for visualisation of the processing

## Example

### Directly

```bash
# convert pdf to images
pdftoppm -png /path/to/example.pdf /path/to/converted/page
# build the script
go build -o main
# run the script
CONVERT_PATH=/path/to/converted OUTPUT_PATH=/path/to/output TESTING_PAGE_NUMBER=2 ./main
```

### With Docker
```
# build the docker image
docker build --build-arg PDF_PATH=/path/to/example.pdf --build-arg CONVERT_PATH=/path/to/converted -t leadbank-uar-pdf-parsing .
# run the docker image
docker run -it --rm /host/output/directory:/path/to/output -e OUTPUT_PATH=/path/to/output leadbank-uar-pdf-parsing
```
