# Go Blobstream verifier

This Go program verifies the inclusion of a blob in the Celestia
network and its commitment in the BlobstreamX contract on the
Ethereum network. It fetches a transaction from the Celestia
network using a specific hash, generates a proof of inclusion
for the blob, and verifies the proof. The program also generates
a proof of inclusion for the data root in the BlobstreamX contract
and verifies it, demonstrating the interoperability between the
Celestia and Ethereum networks.

## Running this demo

1. Clone the repository and change into the directory:

    ```bash
    cd $HOME
    git clone https://github.com/jcstein/blobstream-demo.git
    cd blobstream-demo
    ```

2. Add missing and remove unused modules from `go.mod`:

    ```bash
    go mod tidy
    ```

3. Download modules required for this demo:

    ```bash
    go mod download
    ```

4. Run the demo:

    ```bash
    go run main.go
    ```

5. See the results in your logs:

    ```logs
    3:04PM INF verifying that the blob was committed to by Blobstream height=105058 tx_hash=4B122452FA679F15B458271512816B933803D5870919F67969B4D62221D70346
    3:04PM INF proving shares inclusion to data root end_share=7 height=105058 start_share=5
    3:04PM DBG getting shares proof from tendermint node
    3:04PM DBG verifying shares proofs
    3:04PM INF proofs from shares to data root are valid
    3:04PM INF proving that the data root was committed to in the BlobstreamX contract contract_address=0x046120E6c6C48C05627FB369756F5f44858950a5 fist_block=105001 last_block=106001 nonce=106
    3:04PM DBG getting the data root to commitment inclusion proof
    3:04PM INF verifying that the data root was committed to in the BlobstreamX contract
    3:04PM INF the BlobstreamX contract has committed to the provided blob
    ```

## Dependencies

The program uses several Go packages, including:

- `context`, `encoding/hex`, `fmt`, `math/big`, `os` from the standard library
- `github.com/celestiaorg/blobstream-contracts/v4/wrappers/Blobstream.sol` for interacting with the Blobstream contract
- `github.com/celestiaorg/celestia-app/pkg/square` for calculating the range of shares that the blob occupies in the block
- `github.com/cosmos/cosmos-sdk/server`, `github.com/rs/zerolog`for logging
- `github.com/ethereum/go-ethereum/accounts/abi/bind`, `github.com/ethereum/go-ethereum/common`, `github.com/ethereum/go-ethereum/ethclient` for Ethereum client operations
- `github.com/tendermint/tendermint/crypto/merkle`, `github.com/tendermint/tendermint/rpc/client/http` for Tendermint operations
- `google.golang.org/grpc`, `google.golang.org/grpc/credentials/insecure` for gRPC operations

## Constants

The program uses several constants, including:

- `txHash`: the hash of the transaction to fetch
- `blobIndex`: the index of the blob (currently only the first blob is supported)
- `rpcEndpoint`: the endpoint of the Tendermint RPC server
- `celesGRPC`: the endpoint of the Celestia gRPC server
- `evmRPC`: the endpoint of the EVM chain RPC server
- `contractAddr`: the address of the BlobstreamX contract
- `dataCommitmentStartBlock`, dataCommitmentEndBlock: the range of blocks for the data commitment
- `dataCommitmentNonce`: the nonce of the attestation

## Main function

The `main` function starts the verification process and exits with an error
code if the verification fails.

## Verify function

The `verify` function orchestrates the verification process. It decodes
the transaction hash, fetches the corresponding transaction and block
from the Celestia network, and generates and verifies a proof of inclusion
for the blob in the block. It also interacts with the BlobstreamX
contract on the Ethereum network, generating and verifying a proof of
inclusion for the data root. The function logs the start and result of
each major step in the process.

## VerifyDataRootInclusion function

The `VerifyDataRootInclusion` function verifies the proof of inclusion for
the data root in the BlobstreamX contract. It prepares the data root and the
proof for verification, and calls the `VerifyAttestation` method of the
BlobstreamX contract.
