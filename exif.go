package images

import (
	"encoding/binary"
	"io"
)

type orientation int

const (
	// 方向未指定
	orientationUnspecified = 0
	// 正常方向
	orientationNormal = 1
	// 需水平翻转
	orientationFlipH = 2
	// 需旋转180度
	orientationRotate180 = 3
	// 需垂直翻转
	orientationFlipV = 4
	// 需对角线翻转（左上到右下）
	orientationTranspose = 5
	// 需逆时针旋转270度
	orientationRotate270 = 6
	// 需对角线翻转（右上到左下）
	orientationTransverse = 7
	// 需顺时针旋转90度
	orientationRotate90 = 8
)

const (
	// Start Of Image：表示 JPEG 图片流的起始
	markerSOI = 0xffd8
	// Application Segment 1：表示 APP1 区块，EXIF 信息通常存储在 APP1 区块内
	markerAPP1 = 0xffe1
	// Exif Header：表示 APP1 区块确实包含了 EXIF 信息（紧跟在 APP1 区块标识后），且后面通常跟着两个填充字节
	exifHeader = 0x45786966
	// Big Endian byte order mark：如果 EXIF 段使用大端字节序，那么其字节序标记为 'MM' (0x4D4D)，即（高位字节排在前）
	byteOrderBE = 0x4d4d
	// Little Endian byte order mark：如果 EXIF 段使用小端字节序，那么其字节序标记为 'II' (0x4949)，即（低位字节排在前）
	byteOrderLE = 0x4949
	// Orientation Tag：表示图像的方向
	orientationTag = 0x0112
)

func readOrientation(reader io.Reader) orientation {
	// 检查 JPEG 开始标记（PNG 和 GIF 等格式不是传统意义上的摄影，图像元数据一般不包括拍摄方向信息。处理这些图像文件时，通常没有必要读取或调整图像方向）
	var soi uint16
	if binary.Read(reader, binary.BigEndian, &soi) != nil {
		return orientationUnspecified
	}
	if soi != markerSOI {
		return orientationUnspecified
	}

	for {
		var marker, size uint16
		if err := binary.Read(reader, binary.BigEndian, &marker); err != nil {
			return orientationUnspecified
		}
		if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
			return orientationUnspecified
		}
		// 检查是否是有效的 JPEG 标记
		if marker>>8 != 0xff {
			return orientationUnspecified
		}
		// 检查是否为 APP1 标记
		if marker == markerAPP1 {
			break
		}
		// 对于任何 JPEG 数据块，其报告的大小应至少为2字节
		if size < 2 {
			return orientationUnspecified
		}
		// 这里的减2表示减去size本身占用的2字节（size表示的是从size开始这个段还有几个字节）
		if _, err := io.CopyN(io.Discard, reader, int64(size-2)); err != nil {
			return orientationUnspecified
		}
	}

	// 检查 exifHeader 标记
	var header uint32
	if err := binary.Read(reader, binary.BigEndian, &header); err != nil {
		return orientationUnspecified
	}
	if header != exifHeader {
		return orientationUnspecified
	}
	if _, err := io.CopyN(io.Discard, reader, 2); err != nil {
		return orientationUnspecified
	}

	// 从文件中读取的字节序标识
	var byteOrderTag uint16
	var byteOrder binary.ByteOrder
	if err := binary.Read(reader, binary.BigEndian, &byteOrderTag); err != nil {
		return orientationUnspecified
	}
	switch byteOrderTag {
	case byteOrderBE:
		byteOrder = binary.BigEndian
	case byteOrderLE:
		byteOrder = binary.LittleEndian
	default:
		return orientationUnspecified
	}
	if _, err := io.CopyN(io.Discard, reader, 2); err != nil {
		return orientationUnspecified
	}

	// 跳过 exif 段
	var offset uint32
	if err := binary.Read(reader, binary.BigEndian, &offset); err != nil {
		return orientationUnspecified
	}
	if offset < 8 {
		// 在 TIFF 格式中，如果 offset 小于 8（byteOrderTag、填充字节、offset字节），那么它指向的位置是不合逻辑的，表明可能是一个损坏或非法格式的文件。
		return orientationUnspecified
	}
	if _, err := io.CopyN(io.Discard, reader, int64(offset-8)); err != nil {
		return orientationUnspecified
	}

	// 获取标签数
	var numTags uint16
	if err := binary.Read(reader, byteOrder, &numTags); err != nil {
		return orientationUnspecified
	}

	for i := 0; i < int(numTags); i++ {
		var tag uint16
		if err := binary.Read(reader, binary.BigEndian, &tag); err != nil {
			return orientationUnspecified
		}
		if tag != orientationTag {
			// 10 = 2（数据类型）+ 4（计数）+ 4（值或值偏移量）
			if _, err := io.CopyN(io.Discard, reader, 10); err != nil {
				return orientationUnspecified
			}
			continue
		}

		// 跳过2字节（数据类型）+ 4字节（计数）
		if _, err := io.CopyN(io.Discard, reader, 6); err != nil {
			return orientationUnspecified
		}

		// 读取方向值（在 TIFF 中，实际的方向值可以直接存放在“值或值偏移量”的位置，并且仅占用前两字节，剩余的两字节则不会包含任何重要信息）
		var direction uint16
		if err := binary.Read(reader, binary.BigEndian, &direction); err != nil {
			return orientationUnspecified
		}

		if direction < 1 || direction > 8 {
			// EXIF 规范定义的图像方向值应该在 1 到 8 之间
			return orientationUnspecified
		}
		return orientation(direction)
	}
	return orientationUnspecified
}
