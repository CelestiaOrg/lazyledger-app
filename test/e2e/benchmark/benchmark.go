//nolint:staticcheck
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/celestiaorg/celestia-app/v2/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v2/test/e2e/testnet"
	"github.com/celestiaorg/celestia-app/v2/test/util/testnode"
	"github.com/tendermint/tendermint/pkg/trace"
)

type BenchmarkTest struct {
	*testnet.Testnet
	manifest *Manifest
}

func NewBenchmarkTest(name string, manifest *Manifest) (*BenchmarkTest, error) {
	// create a new testnet
	testNet, err := testnet.New(name, seed,
		testnet.GetGrafanaInfoFromEnvVar(), manifest.ChainID,
		manifest.GetGenesisModifiers()...)
	if err != nil {
		return nil, err
	}

	testNet.SetConsensusParams(manifest.GetConsensusParams())
	return &BenchmarkTest{Testnet: testNet, manifest: manifest}, nil
}

// SetupNodes creates genesis nodes and tx clients based on the manifest.
// There will be manifest.Validators validators and manifest.TxClients tx clients.
// Each tx client connects to one validator. If TxClients are fewer than Validators, some validators will not have a tx client.
func (b *BenchmarkTest) SetupNodes() error {
	testnet.NoError("failed to create genesis nodes",
		b.CreateGenesisNodes(b.manifest.Validators,
			b.manifest.CelestiaAppVersion, b.manifest.SelfDelegation,
			b.manifest.UpgradeHeight, b.manifest.ValidatorResource))

	// enable latency if specified in the manifest
	if b.manifest.EnableLatency {
		for _, node := range b.Nodes() {
			if err := node.Instance.EnableBitTwister(); err != nil {
				return fmt.Errorf("failed to enable bit twister: %v", err)
			}
		}
	}
	// obtain the GRPC endpoints of the validators
	gRPCEndpoints, err := b.RemoteGRPCEndpoints()
	testnet.NoError("failed to get validators GRPC endpoints", err)
	log.Println("validators GRPC endpoints", gRPCEndpoints)

	// create tx clients and point them to the validators
	log.Println("Creating tx clients")

	err = b.CreateTxClients(b.manifest.TxClientVersion,
		b.manifest.BlobSequences,
		b.manifest.BlobSizes,
		b.manifest.BlobsPerSeq,
		b.manifest.TxClientsResource, gRPCEndpoints)
	testnet.NoError("failed to create tx clients", err)

	log.Println("Setting up testnet")
	testnet.NoError("failed to setup testnet", b.Setup(
		testnet.WithPerPeerBandwidth(b.manifest.PerPeerBandwidth),
		testnet.WithTimeoutPropose(b.manifest.TimeoutPropose),
		testnet.WithTimeoutCommit(b.manifest.TimeoutCommit),
		testnet.WithPrometheus(b.manifest.Prometheus),
		testnet.WithLocalTracing(b.manifest.LocalTracingType),
	))

	if b.manifest.PushTrace {
		log.Println("reading trace push config")
		if pushConfig, err := trace.GetPushConfigFromEnv(); err == nil {
			log.Print("Setting up trace push config")
			for _, node := range b.Nodes() {
				if err = node.Instance.SetEnvironmentVariable(trace.PushBucketName, pushConfig.BucketName); err != nil {
					return fmt.Errorf("failed to set TRACE_PUSH_BUCKET_NAME: %v", err)
				}
				if err = node.Instance.SetEnvironmentVariable(trace.PushRegion, pushConfig.Region); err != nil {
					return fmt.Errorf("failed to set TRACE_PUSH_REGION: %v", err)
				}
				if err = node.Instance.SetEnvironmentVariable(trace.PushAccessKey, pushConfig.AccessKey); err != nil {
					return fmt.Errorf("failed to set TRACE_PUSH_ACCESS_KEY: %v", err)
				}
				if err = node.Instance.SetEnvironmentVariable(trace.PushKey, pushConfig.SecretKey); err != nil {
					return fmt.Errorf("failed to set TRACE_PUSH_SECRET_KEY: %v", err)
				}
				if err = node.Instance.SetEnvironmentVariable(trace.PushDelay, fmt.Sprintf("%d", pushConfig.PushDelay)); err != nil {
					return fmt.Errorf("failed to set TRACE_PUSH_DELAY: %v", err)
				}
			}
		}
	}
	return nil
}

// Run runs the benchmark test for the specified duration in the manifest.
func (b *BenchmarkTest) Run() error {
	log.Println("Starting testnet")
	err := b.Start()
	if err != nil {
		return fmt.Errorf("failed to start testnet: %v", err)
	}

	// add latency if specified in the manifest
	if b.manifest.EnableLatency {
		for _, node := range b.Nodes() {
			if err = node.Instance.SetLatencyAndJitter(b.manifest.LatencyParams.
				Latency, b.manifest.LatencyParams.Jitter); err != nil {
				return fmt.Errorf("failed to set latency and jitter: %v", err)
			}
		}
	}

	// wait some time for the tx clients to submit transactions
	time.Sleep(b.manifest.TestDuration)

	return nil
}

func (b *BenchmarkTest) CheckResults(expectedBlockSize int64) error {
	log.Println("Checking results")

	// if local tracing is enabled,
	// pull round state traces to confirm tracing is working as expected.
	if b.manifest.LocalTracingType == "local" {
		if _, err := b.Node(0).PullBlockSummaryTraces("."); err != nil {
			return fmt.Errorf("failed to pull round state traces: %w", err)
		}
	}

	// download traces from S3, if enabled
	if b.manifest.PushTrace && b.manifest.DownloadTraces {
		// download traces from S3
		pushConfig, _ := trace.GetPushConfigFromEnv()
		err := trace.S3Download("./traces/", b.manifest.ChainID,
			pushConfig)
		if err != nil {
			return fmt.Errorf("failed to download traces from S3: %w", err)
		}
	}

	log.Println("Reading blockchain")
	blockchain, err := testnode.ReadBlockchain(context.Background(),
		b.Node(0).AddressRPC())
	testnet.NoError("failed to read blockchain", err)

	targetSizeReached := false
	maxBlockSize := int64(0)
	for _, block := range blockchain {
		if appconsts.LatestVersion != block.Version.App {
			return fmt.Errorf("expected app version %d, got %d", appconsts.LatestVersion, block.Version.App)
		}
		size := int64(block.Size())
		if size >= expectedBlockSize {
			targetSizeReached = true
			break
		}
		if size > maxBlockSize {
			maxBlockSize = size
		}
	}
	if !targetSizeReached {
		return fmt.Errorf("max reached block size is %d byte and is not within the expected range of %d  and %d bytes", maxBlockSize, expectedBlockSize, b.manifest.MaxBlockBytes)
	}

	return nil
}
