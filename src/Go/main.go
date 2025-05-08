package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/klauspost/reedsolomon"
)

type Quorum struct {
	ID          int
	FingerTable []int
}

type BenchmarkResult struct {
	NumQuorums       int     `json:"num_quorums"`
	ParityShards     int     `json:"parity_shards"`
	DataShards       int     `json:"data_shards"`
	NumMissingShards int     `json:"num_missing_shards"`
	EncodingTimeMS   float64 `json:"encoding_time_ms"`
	DecodingTimeMS   float64 `json:"decoding_time_ms"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	configs := []struct {
		m                int
		numQuorums       int
		dataShards       int
		parityShards     int
		numMissingShards int
	}{
		{4, 10, 5, 3, 1},
		{5, 30, 10, 5, 3},
		{9, 50, 16, 8, 5},
		{10, 100, 20, 10, 5},
		{12, 500, 48, 16, 8},
		{13, 1000, 64, 24, 10},
		{15, 5000, 96, 32, 12},
		{17, 10000, 128, 32, 16},
		{18, 15000, 128, 32, 16},
		{19, 25000, 128, 32, 16},
	}

	var results []BenchmarkResult

	for _, cfg := range configs {
		ringSize := int(math.Pow(2, float64(cfg.m)))
		quorums := createRandomQuorums(cfg.numQuorums, ringSize)
		buildFingerTables(quorums, ringSize, cfg.m)

		selected := quorums[0]
		routingTable := buildRoutingTable(selected)

		dataShards := cfg.dataShards
		if dataShards > len(routingTable) {
			dataShards = len(routingTable)
		}
		if dataShards == 0 {
			fmt.Println("Skipping config: not enough routing entries for encoding")
			continue
		}
		tableToEncode := routingTable[:dataShards]

		startEnc := time.Now()
		allShards, _ := encodeRoutingTable(tableToEncode, dataShards, cfg.parityShards)
		encodingDuration := time.Since(startEnc).Seconds() * 1000 // ms

		if cfg.numMissingShards > cfg.parityShards {
			fmt.Println("Skipping: can't recover that many data shard losses")
			continue
		}
		missingIndices := pickRandomIndices(dataShards+cfg.parityShards, cfg.numMissingShards)
		simulateMissingShards(allShards, missingIndices)

		startDec := time.Now()
		reconstructShards(allShards, dataShards, cfg.parityShards)
		decodingDuration := time.Since(startDec).Seconds() * 1000 // ms

		results = append(results, BenchmarkResult{
			NumQuorums:       cfg.numQuorums,
			ParityShards:     cfg.parityShards,
			DataShards:       dataShards, // Use adjusted dataShards
			NumMissingShards: cfg.numMissingShards,
			EncodingTimeMS:   encodingDuration,
			DecodingTimeMS:   decodingDuration,
		})
	}

	file, err := os.Create("benchmark_results.json")
	if err != nil {
		log.Fatalf("Cannot create JSON file: %v", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(results)
	if err != nil {
		log.Fatalf("Cannot write JSON: %v", err)
	}

	fmt.Println("Benchmarking complete. Results saved to benchmark_results.json")
}

// [Rest of the functions unchanged: createRandomQuorums, buildFingerTables, findSuccessor, buildRoutingTable, encodeRoutingTable, simulateMissingShards, reconstructShards, pickRandomIndices, getMaxEntryLength]
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

	return allShards, shardSize
}

func simulateMissingShards(shards [][]byte, indices []int) {
	for _, index := range indices {
		if index < len(shards) {
			shards[index] = nil
		}
	}
}

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
	}
}

func pickRandomIndices(length, count int) []int {
	if count > length {
		count = length
	}
	perm := rand.Perm(length)
	return perm[:count]
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
