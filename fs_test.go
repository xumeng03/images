package images

import (
	"fmt"
	"testing"
)

func TestOpen(t *testing.T) {
	fileName := "test.jpeg"
	_, err := Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fileName, "读取成功")
}

func TestClose(t *testing.T) {
	fileName := "test.jpeg"
	quality := 50
	img, err := Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = Close(img, "compress_"+fileName, quality)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fileName, "保存成功")
}
