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

1. [Clone this repository](../README.md#running-these-demos)
and change into the verifier directory:

    ```bash
    cd $HOME/blobstream-demo/verifier
    ```

2. Add missing and remove unused modules from `go.mod`, and download modules required for this demo:

    ```bash
    go mod tidy
    go mod download
    ```

3. Run the demo:

    ```bash
    go run main.go
    ```

4. See the results in your logs:

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

## Notes

Alternatively to this demo, celestia-app can be used to verify that
a transaction hash, a range of shares, or a blob referenced by its
transaction hash were committed to by the Blobstream contract.

Run this command to see the menu:

```bash
celestia-appd verify --help
Verifies that a transaction hash, a range of shares, or a blob referenced by its transaction hash were committed to by the QGB contract

Usage:
  celestia-appd verify [command]

Available Commands:
  blob        Verifies that a blob, referenced by its transaction hash, in hex format, has been committed to by the QGB contract
  shares      Verifies that a range of shares has been committed to by the QGB contract. The range should be end exclusive.
  tx          Verifies that a transaction hash, in hex format, has been committed to by the QGB contract

Flags:
  -h, --help   help for verify

Global Flags:
      --home string          directory for config and data (default "/Users/joshstein/.celestia-app")
      --log-to-file string   Write logs directly to a file. If empty, logs are written to stderr
      --log_format string    The logging format (json|plain) (default "plain")
      --log_level string     The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace                print out full stack trace on errors

Use "celestia-appd verify [command] --help" for more information about a command.
```

Here is the same example, queried through `celestia-appd`, as is
used in the rest of this example:

```bash
celestia-appd verify blob \
    4B122452FA679F15B458271512816B933803D5870919F67969B4D62221D70346 0 \
    --contract-address 0x046120E6c6C48C05627FB369756F5f44858950a5 \
    --celes-grpc consensus.lunaroasis.net:9090 --chain-id celestia \
    --evm-chain-id 0x5 --evm-rpc https://eth-goerli.public.blastapi.io \
    --node https://rpc.celestia.pops.one:443
```

Results in:

```logs
I[2023-11-21|17:48:38.249] verifying that the blob was committed to by the QGB tx_hash=4B122452FA679F15B458271512816B933803D5870919F67969B4D62221D70346 height=105058
I[2023-11-21|17:48:38.937] proving shares inclusion to data root        height=105058 start_share=5 end_share=7
D[2023-11-21|17:48:38.937] getting shares proof from tendermint node
D[2023-11-21|17:48:39.341] verifying shares proofs
I[2023-11-21|17:48:39.341] proofs from shares to data root are valid
I[2023-11-21|17:48:39.630] proving that the data root was committed to in the QGB contract contract_address=0x046120E6c6C48C05627FB369756F5f44858950a5 fist_block=104801 last_block=105201 nonce=360
D[2023-11-21|17:48:39.630] getting the data root to commitment inclusion proof
I[2023-11-21|17:48:40.095] verifying that the data root was committed to in the QGB contract
I[2023-11-21|17:48:40.341] the QGB contract didn't commit to the provided shares
```

Looks similar to running `main.go`, right?

![IMG_9185](https://github.com/jcstein/blobstream-demo/assets/46639943/3f157a9e-7b84-4c90-b110-01d1d3ae068a)
