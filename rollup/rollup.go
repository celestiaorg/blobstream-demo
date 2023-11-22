package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	wrapper "github.com/celestiaorg/blobstream-contracts/v4/wrappers/Blobstream.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/rpc/client/http"
)

const (
	txHash                   = "4B122452FA679F15B458271512816B933803D5870919F67969B4D62221D70346"
	rpcEndpoint              = "tcp://consensus.lunaroasis.net:26657"
	evmRPC                   = "https://eth-goerli.public.blastapi.io"
	contractAddr             = "0x046120E6c6C48C05627FB369756F5f44858950a5"
	dataCommitmentStartBlock = 105001
	dataCommitmentEndBlock   = 106001
	dataCommitmentNonce      = 106
)

func main() {
	log.Println("🚀 Starting the verification process...")

	if err := verify(); err != nil {
		log.Fatalf("❌ Verification process failed: %v", err)
	}

	log.Println("✅ Verification process completed successfully.")
}

func verify() error {
	log.Println("🔍 Decoding transaction hash...")
	txHashBz, err := hex.DecodeString(txHash)
	if err != nil {
		return fmt.Errorf("failed to decode transaction hash: %w", err)
	}

	log.Println("🔗 Establishing connection to Celestia...")
	trpc, err := http.New(rpcEndpoint, "/websocket")
	if err != nil {
		return fmt.Errorf("failed to connect to Celestia: %w", err)
	}
	defer trpc.Stop()

	ctx := context.Background()

	log.Println("📦 Fetching transaction with decoded hash from Celestia...")
	tx, err := trpc.Tx(ctx, txHashBz, true)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction: %w", err)
	}

	log.Println("📦 Fetching block from Celestia...")
	block, err := trpc.Block(ctx, &tx.Height)
	if err != nil {
		return fmt.Errorf("failed to fetch block: %w", err)
	}

	log.Println("🔍 Generating data root inclusion proof...")
	dcProof, err := trpc.DataRootInclusionProof(ctx, uint64(tx.Height), dataCommitmentStartBlock, dataCommitmentEndBlock)
	if err != nil {
		return fmt.Errorf("failed to generate data root inclusion proof: %w", err)
	}

	log.Println("🔗 Establishing connection to Ethereum client...")
	ethClient, err := ethclient.Dial(evmRPC)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}
	defer ethClient.Close()

	log.Println("📦 Fetching BlobstreamX contract...")
	contractAddress := ethcmn.HexToAddress(contractAddr)
	blobstreamWrapper, err := wrapper.NewWrappers(contractAddress, ethClient)
	if err != nil {
		return fmt.Errorf("failed to fetch BlobstreamX contract: %w", err)
	}

	log.Println("🔍 Verifying data root inclusion on BlobstreamX contract...")
	valid, err := VerifyDataRootInclusion(
		ctx,
		blobstreamWrapper,
		dataCommitmentNonce,
		uint64(tx.Height),
		block.Block.DataHash,
		dcProof.Proof,
	)
	if err != nil {
		return fmt.Errorf("failed to verify data root inclusion on BlobstreamX contract: %w", err)
	}

	if !valid {
		return fmt.Errorf("data root inclusion verification failed")
	}

	return nil
}

func VerifyDataRootInclusion(
	ctx context.Context,
	blobstreamWrapper *wrapper.Wrappers,
	nonce uint64,
	height uint64,
	dataRoot []byte,
	proof merkle.Proof,
) (bool, error) {
	tuple := wrapper.DataRootTuple{
		Height:   big.NewInt(int64(height)),
		DataRoot: *(*[32]byte)(dataRoot),
	}

	sideNodes := make([][32]byte, len(proof.Aunts))
	for i, aunt := range proof.Aunts {
		sideNodes[i] = *(*[32]byte)(aunt)
	}
	wrappedProof := wrapper.BinaryMerkleProof{
		SideNodes: sideNodes,
		Key:       big.NewInt(proof.Index),
		NumLeaves: big.NewInt(proof.Total),
	}

	valid, err := blobstreamWrapper.VerifyAttestation(
		&bind.CallOpts{},
		big.NewInt(int64(nonce)),
		tuple,
		wrappedProof,
	)
	return valid, err
}
