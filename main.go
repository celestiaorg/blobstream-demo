package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/celestiaorg/celestia-app/pkg/square"
	wrapper "github.com/celestiaorg/quantum-gravity-bridge/v2/wrappers/QuantumGravityBridge.sol"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// Pay for blob transaction hash
	txHash = "4B122452FA679F15B458271512816B933803D5870919F67969B4D62221D70346"
	// Blob index (only the first is supported currently)
	blobIndex = 0
	// Celestia RPC endpoint
	rpcEndpoint = "tcp://consensus.lunaroasis.net:26657"
	// Celestia gRPC endpoint
	celesGRPC = "consensus.lunaroasis.net:9090"
	// EVM chain RPC endpoint. Goerli in this case
	evmRPC = "https://eth-goerli.public.blastapi.io"
	// BlobstreamX contract address
	contractAddr = "0x046120E6c6C48C05627FB369756F5f44858950a5"
	// Begining block of the data commitment range. Can be retrieved by manually checking the events emitted by
	// the BlobstreamX contract: https://goerli.etherscan.io/address/0x6e4f1e9ea315ebfd69d18c2db974eef6105fb803#events
	// You can see the block ranges posted from the event DataCommitmentStored, which emits the proof nonce, start block, end block
	// and data commitment for a valid header range proof.
	// For this deployment, the contract attested to the ranges [96001, 109001] in increments of 1K blocks window.
	// For the previous heights, they are attested to, but not sure which are the ranges, or they're included in a
	// single large data commitment [1, 96001[.
	dataCommitmentStartBlock = 105001
	dataCommitmentEndBlock   = 106001
	// Nonce of the attestation, should also be gotten from checking the contract events manually.
	dataCommitmentNonce = 106
)

func main() {
	if err := verify(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func verify() error {
	txHashBz, err := hex.DecodeString(txHash)
	if err != nil {
		return err
	}

	logger := server.ZeroLogWrapper{Logger: zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()}

	trpc, err := http.New(rpcEndpoint, "/websocket")
	if err != nil {
		return err
	}
	err = trpc.Start()
	if err != nil {
		return err
	}
	defer func(trpc *http.HTTP) {
		err := trpc.Stop()
		if err != nil {
			logger.Debug("error closing connection", "err", err.Error())
		}
	}(trpc)

	ctx := context.Background()

	tx, err := trpc.Tx(ctx, txHashBz, true)
	if err != nil {
		return err
	}

	logger.Info("verifying that the blob was committed to by Blobstream", "tx_hash", txHash, "height", tx.Height)

	blockRes, err := trpc.Block(ctx, &tx.Height)
	if err != nil {
		return err
	}

	blobShareRange, err := square.BlobShareRange(blockRes.Block.Txs.ToSliceOfBytes(), int(tx.Index), int(blobIndex), blockRes.Block.Header.Version.App)
	if err != nil {
		return err
	}

	logger.Info(
		"proving shares inclusion to data root",
		"height",
		tx.Height,
		"start_share",
		blobShareRange.Start,
		"end_share",
		blobShareRange.End,
	)

	logger.Debug("getting shares proof from tendermint node")
	sharesProofs, err := trpc.ProveShares(ctx, uint64(tx.Height), uint64(blobShareRange.Start), uint64(blobShareRange.End))
	if err != nil {
		return err
	}

	logger.Debug("verifying shares proofs")
	// checks if the shares proof is valid.
	// the shares proof is self verifiable because it contains also the rows roots
	// which the nmt shares proof is verified against.
	if !sharesProofs.VerifyProof() {
		logger.Info("proofs from shares to data root are invalid")
		return err
	}

	logger.Info("proofs from shares to data root are valid")

	coreGRPC, err := grpc.Dial(celesGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer func(coreGRPC *grpc.ClientConn) {
		err := coreGRPC.Close()
		if err != nil {
			logger.Debug("error closing connection", "err", err.Error())
		}
	}(coreGRPC)

	contractAddress := ethcmn.HexToAddress(contractAddr)

	logger.Info(
		"proving that the data root was committed to in the BlobstreamX contract",
		"contract_address",
		contractAddress.Hex(),
		"fist_block",
		dataCommitmentStartBlock,
		"last_block",
		dataCommitmentEndBlock,
		"nonce",
		dataCommitmentNonce,
	)

	logger.Debug("getting the data root to commitment inclusion proof")
	dcProof, err := trpc.DataRootInclusionProof(ctx, uint64(tx.Height), dataCommitmentStartBlock, dataCommitmentEndBlock)
	if err != nil {
		return err
	}

	block, err := trpc.Block(ctx, &tx.Height)
	if err != nil {
		return err
	}

	ethClient, err := ethclient.Dial(evmRPC)
	if err != nil {
		return err
	}
	defer ethClient.Close()

	blobstreamWrapper, err := wrapper.NewWrappers(contractAddress, ethClient)
	if err != nil {
		return err
	}

	logger.Info("verifying that the data root was committed to in the BlobstreamX contract")
	isCommittedTo, err := VerifyDataRootInclusion(
		ctx,
		blobstreamWrapper,
		dataCommitmentNonce,
		uint64(tx.Height),
		block.Block.DataHash,
		dcProof.Proof,
	)
	if err != nil {
		return err
	}

	if isCommittedTo {
		logger.Info("the BlobstreamX contract has committed to the provided blob")
	} else {
		logger.Info("the BlobstreamX contract didn't commit to the provided blob")
	}
	return nil
}

func VerifyDataRootInclusion(
	_ context.Context,
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
	if err != nil {
		return false, err
	}
	return valid, nil
}
