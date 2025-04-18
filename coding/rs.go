// coding/rs.go
package coding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

// ShardMetadata holds info required for decoding
type ShardMetadata struct {
	DataShards   int
	ParityShards int
	OriginalSize int
}

// EncodedRoutingTable wraps the shards and metadata
type EncodedRoutingTable struct {
	Metadata ShardMetadata
	Shards   [][]byte
}

// EncodeRoutingTableWithParams encodes the routing table using given shard parameters
func EncodeRoutingTableWithParams(rt map[string]string, dataShards, parityShards int) (*EncodedRoutingTable, error) {
	dataBytes, err := json.Marshal(rt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal routing table: %w", err)
	}

	rsEnc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reedsolomon encoder: %w", err)
	}

	shards, err := rsEnc.Split(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to split data: %w", err)
	}

	err = rsEnc.Encode(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to encode parity shards: %w", err)
	}

	meta := ShardMetadata{
		DataShards:   dataShards,
		ParityShards: parityShards,
		OriginalSize: len(dataBytes),
	}

	return &EncodedRoutingTable{Metadata: meta, Shards: shards}, nil
}

// DecodeRoutingTable reconstructs the routing table from shards
func DecodeRoutingTable(encoded *EncodedRoutingTable) (map[string]string, error) {
	meta := encoded.Metadata
	shards := encoded.Shards
	totalShards := meta.DataShards + meta.ParityShards

	if len(shards) != totalShards {
		return nil, errors.New("shard count mismatch")
	}

	rsEnc, err := reedsolomon.New(meta.DataShards, meta.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("decoder init failed: %w", err)
	}

	// Replace zero-length slices with nil to indicate missing
	for i, shard := range shards {
		if shard != nil && len(shard) == 0 {
			shards[i] = nil
		}
	}

	if err := rsEnc.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("reconstruction failed: %w", err)
	}

	ok, err := rsEnc.Verify(shards)
	if err != nil {
		return nil, fmt.Errorf("verification error: %w", err)
	}
	if !ok {
		return nil, errors.New("verification failed: data integrity compromised")
	}

	var buf bytes.Buffer
	if err := rsEnc.Join(&buf, shards, meta.OriginalSize); err != nil {
		return nil, fmt.Errorf("failed to join shards: %w", err)
	}

	var rt map[string]string
	if err := json.Unmarshal(buf.Bytes(), &rt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return rt, nil
}
