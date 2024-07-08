package imageproc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
)

const grayChars = " .:coP#%O?@"

func ProcessStream(stream io.Reader, contentType string) error {
	boundary := extractBoundary(contentType)
	reader := multipart.NewReader(stream, boundary)

	topLeft := []byte("\033[H")
	prevFrame := []byte{}
	for {
		part, err := reader.NextPart()
		if err != nil {
			return fmt.Errorf("error reading next part: %v", err)
		}

		img, err := jpeg.Decode(part)
		if err != nil {
			log.Printf("error decoding image: %v", err)
			continue
		}
		frame := convertToASCII(img)
		if compareSlices(prevFrame, frame) {
			continue
		}
		os.Stdout.Write(topLeft)
		os.Stdout.Write(frame)
		prevFrame = frame
	}
}

func extractBoundary(contentType string) string {
	parts := strings.Split(contentType, "boundary=")
	if len(parts) != 2 {
		log.Fatal("invalid content type")
	}
	return parts[1]
}

func scaleAndConvertToASCII(img image.Image, targetWidth, targetHeight int) []byte {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	scaleX := float64(originalWidth) / float64(targetWidth)
	scaleY := float64(originalHeight) / float64(targetHeight)

	buffer := bytes.Buffer{}
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			pixel := img.At(srcX, srcY)
			rgba := color.RGBAModel.Convert(pixel).(color.RGBA)
			grayScale := color.GrayModel.Convert(pixel).(color.Gray).Y
			scale := int(grayScale) * len(grayChars) / 255
			if scale > 10 {
				scale = 10
			}

			writeInColor(&buffer, rgba, rune(grayChars[scale]))
		}
		buffer.WriteString("\n")
	}

	return buffer.Bytes()
}

// func to compare two byte slices
func compareSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Convert RGBA to colored ASCII character and write to buffer
func writeInColor(buffer *bytes.Buffer, rgba color.RGBA, char rune) {
	r, g, b, _ := rgba.RGBA()

	// ANSI escape sequence for the color
	colorCode := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8)
	resetCode := "\x1b[0m"

	// escape sequence followed by the character and the reset code
	buffer.WriteString(colorCode)
	buffer.WriteRune(char)
	buffer.WriteString(resetCode)
}

func convertToASCII(img image.Image) []byte {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	aspectRatio := float64(width) / float64(height)
	newWidth := 200
	newHeight := int(float64(newWidth) / aspectRatio / 2)

	frame := scaleAndConvertToASCII(img, newWidth, newHeight)

	return frame
}
