package onepay

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	qrcode "github.com/yeqown/go-qrcode/v2"
)

type qrEncoder func(string, int) ([]byte, error)

func encodePNG(text string, size int) ([]byte, error) {
	code, err := qrcode.NewWith(text, qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium))
	if err != nil {
		return nil, err
	}
	w := &pngWriter{size: size}
	if err := code.Save(w); err != nil {
		return nil, err
	}
	return w.data, nil
}

type pngWriter struct {
	size int
	data []byte
}

func (w *pngWriter) Close() error { return nil }
func (w *pngWriter) Write(mat qrcode.Matrix) error {
	const quiet = 4
	modules := mat.Width() + quiet*2
	moduleSize := w.size / modules
	if moduleSize < 1 {
		return fmt.Errorf("QR matrix requires at least %d pixels", modules)
	}
	drawnSize := modules * moduleSize
	offset := (w.size - drawnSize) / 2
	img := image.NewGray(image.Rect(0, 0, w.size, w.size))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	bitmap := mat.Bitmap()
	for y, row := range bitmap {
		for x, set := range row {
			if set {
				x0 := offset + (x+quiet)*moduleSize
				y0 := offset + (y+quiet)*moduleSize
				x1, y1 := x0+moduleSize, y0+moduleSize
				for py := y0; py < y1; py++ {
					for px := x0; px < x1; px++ {
						img.SetGray(px, py, color.Gray{Y: 0})
					}
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	w.data = buf.Bytes()
	return nil
}
