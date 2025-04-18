package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"

	// "time"

	"github.com/klauspost/reedsolomon"
)

type Quorum struct {
	ID          int
	FingerTable []int
}

func main() {
	// rand.Seed(time.Now().UnixNano())

	// Parameters
	m := 4
	ringSize := int(math.Pow(2, float64(m)))
	numQuorums := 8
	parityShards := 2

	fmt.Printf("Ring size: %d, Number of Quorums: %d\n", ringSize, numQuorums)

	// Build chord ring and routing table
	quorums := createRandomQuorums(numQuorums, ringSize)
	buildFingerTables(quorums, ringSize, m)
	printFingerTables(quorums)

	selectedQuorum := quorums[0]
	routingTable := buildRoutingTable(selectedQuorum)

	// Encode routing table
	dataShards := len(routingTable)
	allShards, shardSize := encodeRoutingTable(routingTable, dataShards, parityShards)
	fmt.Printf("Shard size: %d\n", shardSize)

	// Simulate faults
	missingIndices := []int{1, 3}
	corruptedIndices := []int{0, 5}
	fmt.Printf("\nSimulating shard faults...\n")
	fmt.Printf("Missing Shard: %d\n", missingIndices)
	fmt.Printf("Corrupted Shard: %d\n", corruptedIndices)

	simulateMissingShards(allShards, missingIndices)
	simulateCorruptedShardsAndMarkNil(allShards, corruptedIndices)

	// Reconstruct and verify
	reconstructShards(allShards, dataShards, parityShards)

	// Display output
	printRecoveredTable(allShards, dataShards)
}

// ---------- Step 1: Generate Chord Quorums and Routing Tables ----------

func createRandomQuorums(numQuorums, ringSize int) []Quorum {
	quorumMap := make(map[int]bool)
	quorums := make([]Quorum, 0, numQuorums)

	for len(quorums) < numQuorums {
		id := rand.Intn(ringSize)
		if !quorumMap[id] {
			quorumMap[id] = true
			quorums = append(quorums, Quorum{ID: id})
		}
	}

	sort.Slice(quorums, func(i, j int) bool {
		return quorums[i].ID < quorums[j].ID
	})

	fmt.Println("Random Quorum IDs:")
	for _, q := range quorums {
		fmt.Printf("%d ", q.ID)
	}
	fmt.Println()

	return quorums
}

func buildFingerTables(quorums []Quorum, ringSize, m int) {
	ids := make([]int, len(quorums))
	for i, q := range quorums {
		ids[i] = q.ID
	}
	for i := range quorums {
		quorums[i].FingerTable = make([]int, m)
		for j := 0; j < m; j++ {
			target := (quorums[i].ID + int(math.Pow(2, float64(j)))) % ringSize
			quorums[i].FingerTable[j] = findSuccessor(target, ids)
		}
	}
}

func findSuccessor(target int, ids []int) int {
	for _, id := range ids {
		if id >= target {
			return id
		}
	}
	return ids[0]
}

func buildRoutingTable(q Quorum) []string {
	table := make([]string, len(q.FingerTable))
	for i, dest := range q.FingerTable {
		table[i] = fmt.Sprintf("Q%d -> Q%d", q.ID, dest)
	}
	return table
}

func printFingerTables(quorums []Quorum) {
	fmt.Println("\nFinger Tables:")
	for _, q := range quorums {
		fmt.Printf("Quorum %d: %v\n", q.ID, q.FingerTable)
	}
}

// ---------- Step 2: Encode Routing Table ----------
func encodeRoutingTable(table []string, dataShards, parityShards int) ([][]byte, int) {
	shardSize := getMaxEntryLength(table)
	totalShards := dataShards + parityShards

	data := make([][]byte, dataShards)
	for i, entry := range table {
		padded := make([]byte, shardSize)
		copy(padded, []byte(entry))
		data[i] = padded
	}

	allShards := make([][]byte, totalShards)
	copy(allShards, data)
	for i := dataShards; i < totalShards; i++ {
		allShards[i] = make([]byte, shardSize)
	}

	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		log.Fatal(err)
	}
	err = enc.Encode(allShards)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nEncoding complete.")
	printShards(allShards)
	return allShards, shardSize
}

// ---------- Step 3: Fault Injection ----------
func simulateMissingShards(shards [][]byte, indices []int) {
	for _, index := range indices {
		if index < len(shards) {
			shards[index] = nil
			fmt.Printf("Shard %d marked as MISSING\n", index)
		}
	}
}

func simulateCorruptedShardsAndMarkNil(shards [][]byte, indices []int) {
	for _, index := range indices {
		if index < len(shards) && shards[index] != nil {
			before := shards[index][0]
			shards[index][0] ^= 0xFF
			fmt.Printf("Shard %d CORRUPTED (flipped bits), marked as nil\n", index, before, shards[index][0])
			shards[index] = nil
		}
	}
}

// ---------- Step 4: Reconstruction and Output ----------

func reconstructShards(shards [][]byte, dataShards, parityShards int) {
	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		log.Fatal(err)
	}

	err = enc.Reconstruct(shards)
	if err != nil {
		log.Fatalf("Reconstruction failed: %v", err)
	}

	ok, err := enc.Verify(shards)
	if err != nil {
		log.Fatalf("Verification error: %v", err)
	}
	if !ok {
		log.Println("Verification failed — possible corruption")
	} else {
		fmt.Println("Reconstruction successful and verified.")
	}
}

func printRecoveredTable(shards [][]byte, dataShards int) {
	fmt.Println("\nRecovered Routing Table:")
	for i := 0; i < dataShards; i++ {
		fmt.Printf("Entry %d: %s\n", i, string(shards[i]))
	}
}

func printShards(shards [][]byte) {
	for i, shard := range shards {
		if shard == nil {
			fmt.Printf("Shard %d: MISSING\n", i)
		} else {
			fmt.Printf("Shard %d: %s\n", i, string(shard))
		}
	}
}

func getMaxEntryLength(entries []string) int {
	max := 0
	for _, s := range entries {
		if len(s) > max {
			max = len(s)
		}
	}
	return max + 10
}
