FROM ubuntu:22.04

# Install OpenCV
WORKDIR /opencv

# Prepare dependencies
RUN apt-get update
RUN apt-get install -y \
    cmake \
    build-essential \
    git \
    pkg-config \
    libgtk-3-dev \
    wget \
    unzip \
    python3-dev \
    python3-numpy \
    libavcodec-dev \
    libavformat-dev \
    libavutil-dev \
    libswscale-dev
    # libavcodec-dev libavformat-dev libswscale-dev python-dev python-numpy libtbb2 libtbb-dev \
    # libjpeg-dev libpng-dev libtiff-dev libjasper-dev libdc1394-22-dev unzip

# Download and unzip OpenCV 4.10.0
RUN wget -O opencv-4.10.0.zip https://github.com/opencv/opencv/archive/4.10.0.zip \
    && unzip opencv-4.10.0.zip \
    && wget -O opencv_contrib-4.10.0.zip https://github.com/opencv/opencv_contrib/archive/4.10.0.zip \
    && unzip opencv_contrib-4.10.0.zip \
    && rm opencv-4.10.0.zip opencv_contrib-4.10.0.zip

# Configure and build OpenCV 4.10.0 (with contrib modules)
RUN mkdir opencv-4.10.0/build

RUN cmake \
    -S ./opencv-4.10.0 \
    -B ./opencv-4.10.0/build \
    -DOPENCV_EXTRA_MODULES_PATH=./opencv_contrib-4.10.0/modules/ \
    -DCMAKE_VERBOSE_MAKEFILE:BOOL=ON

RUN cmake \
    --build ./opencv-4.10.0/build \
    -j8 \
    --verbose

RUN cmake \
    --install ./opencv-4.10.0/build \
    --verbose

# Set non-interactive mode to avoid prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install C dependencies
RUN apt-get install -y \
    libopenblas-dev \
    liblapack-dev \
    libjpeg-dev \
    libpng-dev \
    libtiff-dev \
    libglib2.0-dev \
    libpoppler-glib-dev \
    libopencv-dev

ENV DEBIAN_FRONTEND=

# Install GO
WORKDIR /go

# Download and unpack the GO binary
RUN wget https://golang.org/dl/go1.22.4.linux-amd64.tar.gz \
    && tar -xvf go1.22.4.linux-amd64.tar.gz \
    && mv go /usr/local \
    && rm go1.22.4.linux-amd64.tar.gz

# Set env variables necessary for GO
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Prepare PDF parsing 
WORKDIR ${GOPATH}/src/github.com/conductorone/leadbank-uar-pdf-parsing

# Set up environment variables for paths and configurations.
ARG PDF_PATH
ARG CONVERT_PATH

ENV CONVERT_PATH=${CONVERT_PATH}

# Update package lists and install necessary packages for OpenCV, Tesseract, and MuPDF.
RUN apt-get install -y -qq \
    libtesseract-dev \
    libleptonica-dev \
    tesseract-ocr-eng \
    libmupdf-dev \
    poppler-utils

# Copy the local code to the container's workspace.
COPY . .

# Update Go module dependencies.
RUN go mod tidy

# Build the project.
RUN go build -o pdfscript

# Specify the command to run on container start.
CMD ["./pdfscript"]
