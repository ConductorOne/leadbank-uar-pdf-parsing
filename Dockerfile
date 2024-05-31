FROM golang:latest

ARG PDF_PATH
ARG CONVERT_PATH
ARG TESTING_PAGE_NUMBER=-1

ENV CONVERT_PATH=${CONVERT_PATH}
ENV TESTING_PAGE_NUMBER=${TESTING_PAGE_NUMBER}

RUN apt-get update -qq

# You need librariy files and headers of tesseract and leptonica.
RUN apt-get install -y -qq \
    libtesseract-dev \
    libleptonica-dev \
    poppler-utils 

# Load languages.
RUN apt-get install -y -qq tesseract-ocr-eng 

# Setup your cool project with go.mod.
WORKDIR ${GOPATH}/src/github.com/conductorone/leadbank-uar-pdf-parsing
COPY . .
RUN go mod tidy
RUN go build -o pdfscript

# convert the pdf into images
RUN pdftoppm -png ${PDF_PATH} ${CONVERT_PATH}/page

CMD ["./pdfscript"]
