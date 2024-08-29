package images

import (
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

func Decode(reader io.Reader) (image.Image, error) {
	// 数据写入 PipeWriter 对象后，可以通过相应的 PipeReader 对象进行读取；
	pr, pw := io.Pipe()
	// 创建了一个新的 io.Reader 对象，这个对象能够将从其读取的数据同时写入到另一个 io.Writer 中（如同包装类）
	reader = io.TeeReader(reader, pw)
	done := make(chan struct{})
	var _ orientation
	go func() {
		defer close(done)
		_ = readOrientation(pr)
		io.Copy(io.Discard, pr)
	}()
	img, _, err := image.Decode(reader)
	pw.Close()
	<-done
	//fmt.Println(orient)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func Encode(w io.Writer, img image.Image, t string, quality int) error {
	switch t {
	case "jpg":
		fallthrough
	case "jpeg":
		if nrgba, ok := img.(*image.NRGBA); ok && nrgba.Opaque() {
			rgba := &image.RGBA{
				Pix:    nrgba.Pix,
				Stride: nrgba.Stride,
				Rect:   nrgba.Rect,
			}
			return jpeg.Encode(w, rgba, &jpeg.Options{Quality: quality})
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	default:
		println("type error!")
		return nil
	}
}
