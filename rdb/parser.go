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
	TypeListQuicklist2  = 0x12
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
	switch valueType {
	case TypeList:
		return parseSimpleList(reader)
	case TypeListQuicklist, TypeListQuicklist2: // MODIFY THIS LINE - handle both versions
		return parseQuicklist(reader)
	default:
		return nil, fmt.Errorf("unsupported list type: 0x%02X", valueType)
	}
}

func parseSimpleList(reader *bufio.Reader) ([]string, error) {
	length, _, err := readLength(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read list length: %w", err)
	}
	if length == 0 {
		return []string{}, nil
	}

	elements := make([]string, length)
	for i := uint64(0); i < length; i++ {
		element, err := readString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read list element %d: %w", i, err)
		}
		elements[i] = element
	}

	return elements, nil
}

func parseQuicklist(reader *bufio.Reader) ([]string, error) {
	nodeCount, _, err := readLength(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read quicklist node count: %w", err)
	}

	if nodeCount == 0 {
		return []string{}, nil
	}

	fmt.Printf("🔍 Parsing quicklist with %d nodes\n", nodeCount) // ADD THIS

	allElements := make([]string, 0)

	for i := uint64(0); i < nodeCount; i++ {
		containerData, err := readString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read quicklist node %d: %w", i, err)
		}

		fmt.Printf("🔍 Node %d: %d bytes\n", i, len(containerData)) // ADD THIS

		if len(containerData) == 0 {
			continue
		}

		elements, err := parseZiplistOrListpack([]byte(containerData))
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse node %d: %v\n", i, err)
			continue
		}

		fmt.Printf("✅ Node %d: parsed %d elements\n", i, len(elements)) // ADD THIS
		allElements = append(allElements, elements...)
	}

	return allElements, nil
}
func parseZiplistOrListpack(data []byte) ([]string, error) {
	if len(data) < 6 {
		// Try as simple string encoding (fallback)
		return []string{string(data)}, nil
	}

	// Check if it's a listpack (starts with total bytes + num entries)
	// Listpack format: [total_bytes:4][num_entries:2][entries...][end:0xFF]

	// Try parsing as ziplist first (more common)
	if len(data) >= 10 {
		result, err := parseZiplist(data)
		if err == nil && len(result) > 0 {
			return result, nil
		}
	}

	// If ziplist parsing fails, try listpack
	return parseListpack(data)
}

func parseListpack(data []byte) ([]string, error) {
	if len(data) < 7 {
		return nil, fmt.Errorf("listpack too short: %d bytes", len(data))
	}

	// Listpack header: [4 bytes total size][2 bytes num elements]
	totalBytes := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	numElements := uint16(data[4]) | uint16(data[5])<<8

	if totalBytes != uint32(len(data)) {
		return nil, fmt.Errorf("listpack size mismatch: expected %d, got %d", totalBytes, len(data))
	}

	if numElements == 0 {
		return []string{}, nil
	}

	elements := make([]string, 0, numElements)
	offset := 6 // Start after header

	for i := uint16(0); i < numElements && offset < len(data)-1; i++ {
		element, newOffset, err := parseListpackEntry(data, offset)
		if err != nil {
			break
		}
		elements = append(elements, element)
		offset = newOffset
	}

	return elements, nil
}
func parseZiplist(data []byte) ([]string, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("ziplist too short: %d bytes", len(data))
	}

	// Read number of entries (bytes 8-9)
	numEntries := uint16(data[8]) | (uint16(data[9]) << 8)

	// Sanity check
	if numEntries > 10000 {
		return nil, fmt.Errorf("unreasonable entry count: %d", numEntries)
	}

	if numEntries == 0 {
		return []string{}, nil
	}

	elements := make([]string, 0, numEntries)
	offset := 10 // Start after header

	for i := uint16(0); i < numEntries && offset < len(data)-1; i++ {
		element, newOffset, err := parseZiplistEntry(data, offset)
		if err != nil {
			// Stop on error but return what we've parsed so far
			break
		}
		elements = append(elements, element)
		offset = newOffset
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("no valid entries parsed")
	}

	return elements, nil
}
func parseListpackEntry(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, fmt.Errorf("offset out of bounds")
	}

	encoding := data[offset]
	offset++

	switch {
	// Small string (length in lower 6 bits)
	case encoding>>7 == 0:
		strLen := int(encoding & 0x3F)
		if offset+strLen > len(data) {
			return "", offset, fmt.Errorf("string out of bounds")
		}
		str := string(data[offset : offset+strLen])
		offset += strLen
		// Skip backlen
		if offset < len(data) {
			offset++
		}
		return str, offset, nil

	// 12-bit string
	case encoding&0xE0 == 0x80:
		if offset >= len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		strLen := int((encoding&0x1F)<<8) | int(data[offset])
		offset++
		if offset+strLen > len(data) {
			return "", offset, fmt.Errorf("string out of bounds")
		}
		str := string(data[offset : offset+strLen])
		offset += strLen
		if offset < len(data) {
			offset++
		}
		return str, offset, nil

	// 32-bit string
	case encoding&0xF0 == 0xF0 && encoding != 0xFF:
		if offset+4 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		strLen := int(data[offset]) | int(data[offset+1])<<8 | int(data[offset+2])<<16 | int(data[offset+3])<<24
		offset += 4
		if offset+strLen > len(data) {
			return "", offset, fmt.Errorf("string out of bounds")
		}
		str := string(data[offset : offset+strLen])
		offset += strLen
		if offset < len(data) {
			offset++
		}
		return str, offset, nil

	// 7-bit unsigned int
	case encoding == 0xC0:
		if offset >= len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int(data[offset])
		offset++
		if offset < len(data) {
			offset++
		}
		return fmt.Sprintf("%d", val), offset, nil

	// 13-bit signed int
	case encoding&0xE0 == 0xC0:
		if offset >= len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int((encoding&0x1F)<<8) | int(data[offset])
		if val&0x1000 != 0 {
			val -= 0x2000 // Sign extend
		}
		offset++
		if offset < len(data) {
			offset++
		}
		return fmt.Sprintf("%d", val), offset, nil

	// 16-bit integer
	case encoding == 0xD0:
		if offset+2 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int16(data[offset]) | int16(data[offset+1])<<8
		offset += 2
		if offset < len(data) {
			offset++
		}
		return fmt.Sprintf("%d", val), offset, nil

	// 32-bit integer
	case encoding == 0xE0:
		if offset+4 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int32(data[offset]) | int32(data[offset+1])<<8 | int32(data[offset+2])<<16 | int32(data[offset+3])<<24
		offset += 4
		if offset < len(data) {
			offset++
		}
		return fmt.Sprintf("%d", val), offset, nil

	default:
		return "", offset, fmt.Errorf("unknown listpack encoding: 0x%02X", encoding)
	}
}

func parseZiplistEntry(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, fmt.Errorf("offset out of bounds")
	}

	// Skip previous entry length (variable length encoding)
	prevLen := int(data[offset])
	if prevLen == 254 {
		offset += 5 // 0xFE + 4 bytes
	} else {
		offset += 1
	}

	if offset >= len(data) {
		return "", offset, fmt.Errorf("unexpected end of ziplist")
	}

	// Read encoding byte
	encoding := data[offset]
	offset++

	// Parse based on encoding
	switch {
	case encoding>>6 == 0: // String with length < 64
		strLen := int(encoding & 0x3F)
		if offset+strLen > len(data) {
			return "", offset, fmt.Errorf("string data out of bounds")
		}
		str := string(data[offset : offset+strLen])
		return str, offset + strLen, nil

	case encoding>>6 == 1: // String with length < 16384
		if offset >= len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		strLen := int((encoding&0x3F)<<8) | int(data[offset])
		offset++
		if offset+strLen > len(data) {
			return "", offset, fmt.Errorf("string data out of bounds")
		}
		str := string(data[offset : offset+strLen])
		return str, offset + strLen, nil

	case encoding>>6 == 2: // String with length >= 16384
		if offset+4 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		strLen := int(data[offset]) | int(data[offset+1])<<8 | int(data[offset+2])<<16 | int(data[offset+3])<<24
		offset += 4
		if offset+strLen > len(data) {
			return "", offset, fmt.Errorf("string data out of bounds")
		}
		str := string(data[offset : offset+strLen])
		return str, offset + strLen, nil

	case encoding == 0xC0: // 16-bit integer
		if offset+2 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int16(data[offset]) | int16(data[offset+1])<<8
		return fmt.Sprintf("%d", val), offset + 2, nil

	case encoding == 0xD0: // 32-bit integer
		if offset+4 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int32(data[offset]) | int32(data[offset+1])<<8 | int32(data[offset+2])<<16 | int32(data[offset+3])<<24
		return fmt.Sprintf("%d", val), offset + 4, nil

	case encoding == 0xE0: // 64-bit integer
		if offset+8 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int64(data[offset]) | int64(data[offset+1])<<8 | int64(data[offset+2])<<16 | int64(data[offset+3])<<24 |
			int64(data[offset+4])<<32 | int64(data[offset+5])<<40 | int64(data[offset+6])<<48 | int64(data[offset+7])<<56
		return fmt.Sprintf("%d", val), offset + 8, nil

	case encoding == 0xF0: // 24-bit integer
		if offset+3 > len(data) {
			return "", offset, fmt.Errorf("unexpected end")
		}
		val := int32(data[offset]) | int32(data[offset+1])<<8 | int32(data[offset+2])<<16
		if val&0x800000 != 0 {
			val |= ^int32(0xFFFFFF) // Sign extend by setting upper 8 bits to 1
		}
		return fmt.Sprintf("%d", val), offset + 3, nil

	case encoding>>4 == 0xF && encoding != 0xFF: // Small integers 0-12
		val := int(encoding & 0x0F)
		return fmt.Sprintf("%d", val-1), offset, nil

	default:
		return "", offset, fmt.Errorf("unknown ziplist encoding: 0x%02X", encoding)
	}
}

func parseStreamValue(reader *bufio.Reader) (interface{}, error) {
	return nil, fmt.Errorf("stream parsing not implemented yet")
}

func parseValue(reader *bufio.Reader, valueType byte) (interface{}, error) {
	switch valueType {
	case TypeString:
		return parseStringValue(reader)

	case TypeList, TypeListQuicklist, TypeListQuicklist2: // MODIFY THIS LINE
		return parseListValue(reader, valueType)

	case TypeStream, TypeStreamListpack, TypeStreamListpack2:
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
