package main

import (
	"log"
	"time"

	"github.com/celestiaorg/celestia-app/v2/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v2/test/e2e/testnet"
)

const (
	seed = 42

	toMiB = 1024 * 1024
	toMB  = 1000 * 1000
)

var bigBlockManifest = Manifest{
	ChainID:    "test",
	Validators: 2,
	TxClients:  2,
	ValidatorResource: testnet.Resources{
		MemoryRequest: "12Gi",
		MemoryLimit:   "12Gi",
		CPU:           "8",
		Volume:        "20Gi",
	},
	TxClientsResource: testnet.Resources{
		MemoryRequest: "1Gi",
		MemoryLimit:   "3Gi",
		CPU:           "2",
		Volume:        "1Gi",
	},
	SelfDelegation:     10000000,
	CelestiaAppVersion: "pr-3541", // "pr-3261",
	TxClientVersion:    "pr-3541", //"pr-3261",
	EnableLatency:      false,
	LatencyParams:      LatencyParams{0, 0}, // in  milliseconds
	BlobSequences:      60,
	BlobsPerSeq:        6,
	BlobSizes:          "200000",
	PerPeerBandwidth:   5 * toMB,
	UpgradeHeight:      0,
	TimeoutCommit:      11 * time.Second,
	TimeoutPropose:     80 * time.Second,
	Mempool:            "v1", // ineffective as it always defaults to v1
	BroadcastTxs:       true,
	Prometheus:         false,
	GovMaxSquareSize:   512,
	MaxBlockBytes:      7800000,
	TestDuration:       10 * time.Minute,
	LocalTracingType:   "local",
	PushTrace:          true,
}

func TwoNodeSimple(logger *log.Logger) error {
	latestVersion, err := testnet.GetLatestVersion()
	testnet.NoError("failed to get latest version", err)

	logger.Println("=== RUN TwoNodeSimple", "version:", latestVersion)

	manifest := Manifest{
		ChainID:            "test-e2e-two-node-simple",
		Validators:         2,
		ValidatorResource:  testnet.DefaultResources,
		TxClientsResource:  testnet.DefaultResources,
		SelfDelegation:     10000000,
		CelestiaAppVersion: "pr-3541", //latestVersion,
		TxClientVersion:    "pr-3541", // testnet.TxsimVersion,
		EnableLatency:      true,
		LatencyParams:      LatencyParams{100, 10}, // in  milliseconds
		BlobsPerSeq:        6,
		BlobSequences:      60,
		BlobSizes:          "200000",
		PerPeerBandwidth:   5 * 1024 * 1024,
		UpgradeHeight:      0,
		TimeoutCommit:      1 * time.Second,
		TimeoutPropose:     1 * time.Second,
		Mempool:            "v1",
		BroadcastTxs:       true,
		Prometheus:         true,
		GovMaxSquareSize:   appconsts.DefaultGovMaxSquareSize,
		MaxBlockBytes:      appconsts.DefaultMaxBytes,
		LocalTracingType:   "local",
		PushTrace:          false,
		DownloadTraces:     false,
		TestDuration:       120 * time.Second,
		TxClients:          2,
	}

	benchTest, err := NewBenchmarkTest("E2EThroughput", &manifest)
	testnet.NoError("failed to create benchmark test", err)

	defer func() {
		log.Print("Cleaning up testnet")
		benchTest.Cleanup()
	}()

	testnet.NoError("failed to setup nodes", benchTest.SetupNodes())

	testnet.NoError("failed to run the benchmark test", benchTest.Run())

	testnet.NoError("failed to check results", benchTest.CheckResults())

	return nil
}

func TwoNodeBigBlock_8MiB(logger *log.Logger) error {
	logger.Println("Running TwoNodeBigBlock_8MiB")
	manifest := bigBlockManifest
	manifest.MaxBlockBytes = 8 * toMB
	manifest.ChainID = "two-node-big-block-8mb"
	logger.Println("ChainID: ", manifest.ChainID)

	benchTest, err := NewBenchmarkTest("TwoNodeBigBlock_8MiB", &manifest)
	testnet.NoError("failed to create benchmark test", err)

	defer func() {
		log.Print("Cleaning up testnet")
		benchTest.Cleanup()
	}()

	testnet.NoError("failed to setup nodes", benchTest.SetupNodes())
	testnet.NoError("failed to run the benchmark test", benchTest.Run())
	testnet.NoError("failed to check results", benchTest.CheckResults())

	return nil
}

func TwoNodeBigBlock_8MiB_Latency(logger *log.Logger) error {
	logger.Println("Running TwoNodeBigBlock_8MiB_Latency")
	manifest := bigBlockManifest
	manifest.ChainID = "two-node-big-block-8mib-latency"
	manifest.MaxBlockBytes = 8 * toMiB
	manifest.EnableLatency = true

	benchTest, err := NewBenchmarkTest("TwoNodeBigBlock_8MiB_Latency", &manifest)
	testnet.NoError("failed to create benchmark test", err)

	defer func() {
		log.Print("Cleaning up testnet")
		benchTest.Cleanup()
	}()

	testnet.NoError("failed to setup nodes", benchTest.SetupNodes())
	testnet.NoError("failed to run the benchmark test", benchTest.Run())
	testnet.NoError("failed to check results", benchTest.CheckResults())

	return nil
}

func TwoNodeBigBlock_32MiB(logger *log.Logger) error {
	logger.Println("Running TwoNodeBigBlock_32MiB")
	manifest := bigBlockManifest
	manifest.ChainID = "two-node-big-block-32mb"
	manifest.MaxBlockBytes = 32 * toMB

	benchTest, err := NewBenchmarkTest("TwoNodeBigBlock_32MiB", &manifest)
	testnet.NoError("failed to create benchmark test", err)

	defer func() {
		log.Print("Cleaning up testnet")
		benchTest.Cleanup()
	}()

	testnet.NoError("failed to setup nodes", benchTest.SetupNodes())
	testnet.NoError("failed to run the benchmark test", benchTest.Run())
	testnet.NoError("failed to check results", benchTest.CheckResults())

	return nil
}

func TwoNodeBigBlock_64MiB(logger *log.Logger) error {
	logger.Println("Running TwoNodeBigBlock_64MiB")
	manifest := bigBlockManifest
	manifest.ChainID = "two-node-big-block-64mb"
	manifest.MaxBlockBytes = 64 * toMB

	benchTest, err := NewBenchmarkTest("TwoNodeBigBlock_64MiB", &manifest)
	testnet.NoError("failed to create benchmark test", err)

	defer func() {
		log.Print("Cleaning up testnet")
		benchTest.Cleanup()
	}()

	testnet.NoError("failed to setup nodes", benchTest.SetupNodes())
	testnet.NoError("failed to run the benchmark test", benchTest.Run())
	testnet.NoError("failed to check results", benchTest.CheckResults())

	return nil
}
