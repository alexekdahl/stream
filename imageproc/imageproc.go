package imageproc

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"os"
	"sync"
)

const grayChars = " .:;coPBO?#%@"

var (
	topLeft   = []byte("\x1b[H")
	resetCode = []byte("\x1b[0m")
)

type streamer interface {
	NextPart() (io.Reader, error)
}

// ProcessVideoStream reads and processes the image stream from a multipart reader.
func ProcessVideoStream(stream streamer) error {
	var bufferPool sync.Pool
	bufferPool.New = func() interface{} {
		buf := make([]byte, 0, 4096) // Pre-allocated buffer size
		return &buf
	}

	for {
		part, err := stream.NextPart()
		if err != nil {
			return fmt.Errorf("error reading next part: %w", err)
		}

		img, err := jpeg.Decode(part)
		if err != nil {
			log.Printf("error decoding image: %v", err)
			continue
		}

		processAndWriteFrame(img, &bufferPool)
	}
}

func processAndWriteFrame(img image.Image, bufferPool *sync.Pool) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	aspectRatio := float64(width) / float64(height)
	newWidth := 200
	newHeight := int(float64(newWidth) / aspectRatio / 2)

	scaleX := float64(width) / float64(newWidth)
	scaleY := float64(height) / float64(newHeight)

	// Precalculate buffer size: (width * height) + newHeight for newlines
	bufferSize := (newWidth * newHeight) + newHeight
	buf := bufferPool.Get().(*[]byte)
	if cap(*buf) < bufferSize {
		*buf = make([]byte, 0, bufferSize)
	}
	defer bufferPool.Put(buf)

	*buf = (*buf)[:0] // Reset buffer
	os.Stdout.Write(topLeft)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			pixel := img.At(srcX, srcY)
			rgba := color.RGBAModel.Convert(pixel).(color.RGBA)
			grayScale := color.GrayModel.Convert(pixel).(color.Gray).Y
			scale := int(grayScale) * len(grayChars) / 255
			if scale >= len(grayChars) {
				scale = len(grayChars) - 1
			}

			char := grayChars[scale]
			writeInColor(buf, rgba, char)
		}
		*buf = append(*buf, '\n')
	}

	os.Stdout.Write(*buf)
}

func writeInColor(buffer *[]byte, rgba color.RGBA, char byte) {
	r, g, b, _ := rgba.RGBA()

	// ANSI escape sequence for the color
	colorCode := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8)

	*buffer = append(*buffer, colorCode...)
	*buffer = append(*buffer, char)
	*buffer = append(*buffer, resetCode...)
}
