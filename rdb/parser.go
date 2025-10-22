package rdb

import (
	"bufio"
	"fmt"
	"time"
)

const (
	OpSelectDB      = 0xFE
	OpEOF           = 0xFF
	OpExpireTimeSec = 0xFD
	OpExpireTimeMs  = 0xFC
	OpResizeDB      = 0xFB
	OpAux           = 0xFA
)

const (
	TypeString        = 0x00
	TypeList          = 0x01
	TypeSet           = 0x02
	TypeSortedSet     = 0x03
	TypeHash          = 0x04
	TypeZipmap        = 0x09
	TypeZiplist       = 0x0A
	TypeIntset        = 0x0B
	TypeSortedSetZL   = 0x0C
	TypeHashZL        = 0x0D
	TypeListQuicklist = 0x0E
	TypeStream        = 0x0F

	TypeHashListpack    = 0x10
	TypeZSet2           = 0x11
	TypeStreamListpack  = 0x13
	TypeHashZipmap2     = 0x14
	TypeStreamListpack2 = 0x15
)

type KeyValue struct {
	Key       string
	Value     interface{}
	ValueType byte
	Expiry    *time.Time
}

func parseHeader(reader *bufio.Reader) (string, error) {
	// Read magic string "REDIS"
	magicBytes, err := readBytes(reader, 5)
	if err != nil {
		return "", fmt.Errorf("failed to read magic string: %w", err)
	}

	magic := string(magicBytes)
	if magic != "REDIS" {
		return "", fmt.Errorf("invalid RDB file: magic string is '%s', expected 'REDIS'", magic)
	}

	// Read version
	versionBytes, err := readBytes(reader, 4)
	if err != nil {
		return "", fmt.Errorf("failed to read version: %w", err)
	}

	version := string(versionBytes)

	// Validate version format
	for _, c := range version {
		if c < '0' || c > '9' {
			return "", fmt.Errorf("invalid RDB version: '%s'", version)
		}
	}

	return version, nil
}

func skipMetadata(reader *bufio.Reader) error {
	for {
		opcode, err := readByte(reader)
		if err != nil {
			return fmt.Errorf("failed to read opcode while skipping metadata: %w", err)
		}

		if opcode != OpAux {
			// Not an aux field, unread the byte by seeking back
			// Since bufio.Reader doesn't support UnreadByte after ReadByte in all cases,
			// we'll handle this by peeking first in actual usage
			// For now, we assume this byte is the start of database section
			return reader.UnreadByte()
		}

		//Skipping aux field key
		_, err = readString(reader)
		if err != nil {
			return fmt.Errorf("failed to skip aux key: %w", err)
		}

		//Skipping aux field value
		_, err = readString(reader)
		if err != nil {
			return fmt.Errorf("failed to skip aux value: %w", err)
		}
	}
}

func parseDatabaseSelector(reader *bufio.Reader) (int, error) {
	// Read the database number
	dbNum, isEncoded, err := readLength(reader)
	if err != nil {
		return 0, fmt.Errorf("failed to read database number: %w", err)
	}

	if isEncoded {
		return 0, fmt.Errorf("unexpected encoded value for database number")
	}

	return int(dbNum), nil
}

func skipHashTableSize(reader *bufio.Reader) error {
	// Read hash table size
	_, _, err := readLength(reader)
	if err != nil {
		return fmt.Errorf("failed to read hash table size: %w", err)
	}

	// Read expiry hash table size
	_, _, err = readLength(reader)
	if err != nil {
		return fmt.Errorf("failed to read expiry table size: %w", err)
	}

	return nil
}

// parseExpiry reads expiry time based on the opcode
func parseExpiry(reader *bufio.Reader, opcode byte) (*time.Time, error) {
	var expiry time.Time

	switch opcode {
	case OpExpireTimeSec:
		timestamp, err := readUint32(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read expiry time (seconds): %w", err)
		}
		expiry = time.Unix(int64(timestamp), 0)

	case OpExpireTimeMs:
		timestamp, err := readUint64(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read expiry time (milliseconds): %w", err)
		}
		expiry = time.Unix(int64(timestamp/1000), int64((timestamp%1000)*1000000))

	default:
		return nil, fmt.Errorf("invalid expiry opcode: 0x%02X", opcode)
	}

	return &expiry, nil
}

func parseStringValue(reader *bufio.Reader) (string, error) {
	value, err := readString(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse string value: %w", err)
	}
	return value, nil
}

func parseListValue(reader *bufio.Reader, valueType byte) ([]string, error) {
	return nil, fmt.Errorf("stream parsing not implemented yet")
	//		switch valueType {
	//		case TypeList:
	//			return parseSimpleList(reader)
	//		case TypeListQuicklist:
	//			return parseQuickList(reader)
	//		default:
	//			return nil, fmt.Errorf("unsupported list type: 0x%02X", valueType)
	//		}
	//	}
	//
	//	func parseSimpleList(reader *bufio.Reader) ([]string, error) {
	//		length, _, err := readLength(reader)
	//		if err != nil {
	//			return nil, fmt.Errorf("failed to read list length: %w", err)
	//		}
	//		if length == 0 {
	//			return []string{}, nil
	//		}
	//
	//		elements := make([]string, length)
	//		for i:= uint64(0); i< length ; i++{
	//			element, err:=
	//		}
	//
	// }
	// func parseQuickList(reader *bufio.Reader) ([]string, error) {
	//
	// }
}
func parseStreamValue(reader *bufio.Reader) (interface{}, error) {
	return nil, fmt.Errorf("stream parsing not implemented yet")
}

func parseValue(reader *bufio.Reader, valueType byte) (interface{}, error) {
	switch valueType {
	case TypeString:
		return parseStringValue(reader)

	case TypeList, TypeListQuicklist:
		return parseListValue(reader, valueType)

	case TypeStream:
		return parseStreamValue(reader)

	case TypeSet, TypeSortedSet, TypeHash:
		return nil, fmt.Errorf("type 0x%02X not implemented yet", valueType)

	case TypeZiplist, TypeZipmap, TypeIntset, TypeSortedSetZL, TypeHashZL:
		return nil, fmt.Errorf("compressed type 0x%02X not implemented yet", valueType)

	case TypeHashListpack, TypeZSet2, TypeHashZipmap2:
		return nil, fmt.Errorf("listpack/zipmap type 0x%02X not implemented yet", valueType)

	default:
		return nil, fmt.Errorf("unknown value type: 0x%02X", valueType)
	}
}

func parseKeyValuePair(reader *bufio.Reader) (*KeyValue, error) {
	var expiry *time.Time

	//Read the first opcode/value-type byte
	opcode, err := readByte(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read opcode: %w", err)
	}

	switch opcode {
	case OpEOF:
		return nil, nil

	case OpSelectDB:
		reader.UnreadByte()
		return nil, nil

	case OpExpireTimeSec, OpExpireTimeMs:
		expiry, err = parseExpiry(reader, opcode)
		if err != nil {
			return nil, err
		}
		//Read the actual value type that follows the expiry
		opcode, err = readByte(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read value type after expiry: %w", err)
		}

	case OpResizeDB:
		err = skipHashTableSize(reader)
		if err != nil {
			return nil, err
		}
		return parseKeyValuePair(reader)

	case OpAux:
		_, err = readString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to skip aux key: %w", err)
		}
		_, err = readString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to skip aux value: %w", err)
		}
		return parseKeyValuePair(reader)
	}

	valueType := opcode

	// reading the key
	key, err := readString(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}

	value, err := parseValue(reader, valueType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value for key '%s': %w", key, err)
	}

	return &KeyValue{
		Key:       key,
		Value:     value,
		ValueType: valueType,
		Expiry:    expiry,
	}, nil
}
