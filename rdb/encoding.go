package rdb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// 6-bit length (0-63)
	LEN_6BIT = 0x00
	// 14-bit length (0-16383)
	LEN_14BIT = 0x01
	// 32-bit length
	LEN_32BIT = 0x02
	// Special encoding (integer or compressed string)
	LEN_SPECIAL = 0x03

	// Special encoding types
	ENC_INT8  = 0 // 8-bit integer
	ENC_INT16 = 1 // 16-bit integer
	ENC_INT32 = 2 // 32-bit integer
	ENC_LZF   = 3 // LZF compressed string
)

func readByte(reader *bufio.Reader) (byte, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("failed to read byte: %w", err)
	}
	return b, nil
}

func readBytes(reader *bufio.Reader, n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("invalid byte count: %d", n)
	}

	if n == 0 {
		return []byte{}, nil
	}

	buf := make([]byte, n)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read %d bytes: %w", n, err)
	}
	return buf, nil
}

func readLength(reader *bufio.Reader) (uint64, bool, error) {
	firstByte, err := readByte(reader)
	if err != nil {
		return 0, false, err
	}

	encType := (firstByte & 0xC0) >> 6

	switch encType {
	case LEN_6BIT:
		length := uint64(firstByte & 0x3F)
		return length, false, nil

	case LEN_14BIT:
		secondByte, err := readByte(reader)
		if err != nil {
			return 0, false, err
		}
		length := uint64(firstByte&0x3F)<<8 | uint64(secondByte)
		return length, false, nil

	case LEN_32BIT:
		buf, err := readBytes(reader, 4)
		if err != nil {
			return 0, false, err
		}
		length := uint64(binary.BigEndian.Uint32(buf))
		return length, false, nil

	case LEN_SPECIAL:
		encoding := firstByte & 0x3F
		return uint64(encoding), true, nil

	default:
		return 0, false, fmt.Errorf("invalid length encoding: %d", encType)
	}
}

func readString(reader *bufio.Reader) (string, error) {
	length, isEncoded, err := readLength(reader)
	if err != nil {
		return "", err
	}

	if isEncoded {
		intVal, err := readEncodedInteger(reader, byte(length))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", intVal), nil
	}

	if length == 0 {
		return "", nil
	}

	buf, err := readBytes(reader, int(length))
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func readEncodedInteger(reader *bufio.Reader, encoding byte) (int64, error) {
	switch encoding {
	case ENC_INT8:
		buf, err := readByte(reader)
		if err != nil {
			return 0, err
		}
		return int64(int8(buf)), nil

	case ENC_INT16:
		buf, err := readBytes(reader, 2)
		if err != nil {
			return 0, err
		}
		val := binary.LittleEndian.Uint16(buf)
		return int64(int16(val)), nil

	case ENC_INT32:
		buf, err := readBytes(reader, 4)
		if err != nil {
			return 0, err
		}
		val := binary.LittleEndian.Uint32(buf)
		return int64(int32(val)), nil

	case ENC_LZF:
		return 0, fmt.Errorf("LZF compression not supported yet")

	default:
		return 0, fmt.Errorf("unknown integer encoding: %d", encoding)
	}
}

func readInteger(reader *bufio.Reader, encoding byte) (int64, error) {
	return readEncodedInteger(reader, encoding)
}

func readUint32(reader *bufio.Reader) (uint32, error) {
	buf, err := readBytes(reader, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

func readUint64(reader *bufio.Reader) (uint64, error) {
	buf, err := readBytes(reader, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf), nil
}
