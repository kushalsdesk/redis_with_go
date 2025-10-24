package rdb

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

// LoadRDB is the main entry point for loading an RDB file
func LoadRDB(filepath string) error {
	// Checking if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("RDB file not found: %s", filepath)
	}

	// Opening the file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open RDB file: %w", err)
	}
	defer file.Close()

	// Creating buffered reader for efficient reading
	reader := bufio.NewReader(file)

	// Parsing header
	version, err := parseHeader(reader)
	if err != nil {
		return fmt.Errorf("invalid RDB header: %w", err)
	}
	fmt.Printf("📋 RDB version: %s\n", version)

	// Skipping metadata
	err = skipMetadata(reader)
	if err != nil {
		return fmt.Errorf("failed to skip metadata: %w", err)
	}

	// Load databases
	totalKeys := 0
	skippedKeys := 0
	currentDB := -1

	for {
		opcode, err := readByte(reader)
		if err != nil {
			return fmt.Errorf("failed to read opcode: %w", err)
		}

		switch opcode {
		case OpEOF:
			fmt.Printf("✅ RDB loading complete: loaded %d keys, skipped %d expired keys\n",
				totalKeys, skippedKeys)

			// TODO: Validate CRC64 checksum
			// For now, we skip checksum validation
			return nil

		case OpSelectDB:
			// Database selector
			dbNum, err := parseDatabaseSelector(reader)
			if err != nil {
				return fmt.Errorf("failed to parse database selector: %w", err)
			}

			currentDB = dbNum
			fmt.Printf("📂 Loading database %d\n", dbNum)

			if dbNum != 0 {
				fmt.Printf("⚠️  Warning: RDB contains database %d, but only database 0 is supported\n", dbNum)
			}
			continue

		case OpResizeDB:
			err = skipHashTableSize(reader)
			if err != nil {
				return fmt.Errorf("failed to skip hash table size: %w", err)
			}
			continue

		case OpAux:
			_, err = readString(reader)
			if err != nil {
				return fmt.Errorf("failed to skip aux key: %w", err)
			}
			_, err = readString(reader)
			if err != nil {
				return fmt.Errorf("failed to skip aux value: %w", err)
			}
			continue

		default:
			err = reader.UnreadByte()
			if err != nil {
				return fmt.Errorf("failed to unread byte: %w", err)
			}

			// Parsing key-value pair
			kv, err := parseKeyValuePair(reader)
			if err != nil {
				return fmt.Errorf("failed to parse key-value pair at database %d: %w", currentDB, err)
			}

			if kv == nil {
				// EOF or database selector encountered
				continue
			}

			// Store the key-value pair
			loaded, err := storeKeyValue(kv)
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to store key '%s': %v\n", kv.Key, err)
				continue
			}

			if loaded {
				totalKeys++
				// if totalKeys%1000 == 0 {
				// 	fmt.Printf("⏳ Loaded %d keys...\n", totalKeys)
				// }
			} else {
				skippedKeys++
			}
		}
	}
}

func storeKeyValue(kv *KeyValue) (bool, error) {
	if kv.Expiry != nil && time.Now().After(*kv.Expiry) {
		return false, nil
	}

	// Calculate TTL if expiry exists
	var ttl time.Duration
	if kv.Expiry != nil {
		ttl = time.Until(*kv.Expiry)
		if ttl < 0 {
			return false, nil
		}
	}

	switch kv.ValueType {
	case TypeString:
		value, ok := kv.Value.(string)
		if !ok {
			return false, fmt.Errorf("expected string value, got %T", kv.Value)
		}

		store.Set(kv.Key, value, ttl)
		return true, nil

	case TypeList, TypeListQuicklist, TypeListQuicklist2: // MODIFY THIS LINE
		elements, ok := kv.Value.([]string)
		if !ok {
			return false, fmt.Errorf("expected list value, got %T", kv.Value)
		}

		if len(elements) == 0 {
			// Empty list - still create it
			store.CreateEmptyList(kv.Key, ttl)
			return true, nil
		}

		// Store list using RPUSH to maintain order
		length := store.ListPushBulk(kv.Key, elements, false, ttl)
		if length == -1 {
			return false, fmt.Errorf("failed to create list")
		}

		return true, nil

	case TypeStream, TypeStreamListpack, TypeStreamListpack2:
		return false, fmt.Errorf("stream type not implemented yet")

	default:
		return false, fmt.Errorf("unsupported value type: 0x%02X", kv.ValueType)
	}
}

func LoadDatabase(reader *bufio.Reader) (int, int, error) {
	totalKeys := 0
	skippedKeys := 0

	for {
		kv, err := parseKeyValuePair(reader)
		if err != nil {
			return totalKeys, skippedKeys, err
		}

		if kv == nil {
			// EOF or database selector
			break
		}

		loaded, err := storeKeyValue(kv)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to store key '%s': %v\n", kv.Key, err)
			continue
		}

		if loaded {
			totalKeys++
		} else {
			skippedKeys++
		}
	}

	return totalKeys, skippedKeys, nil
}
