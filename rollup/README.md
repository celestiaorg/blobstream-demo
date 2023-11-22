# Blobstream rollup verification demo

This program is a basic demonstration of rollup block verification.
In the context of blockchain technology, a rollup is a Layer 2 solution
that increases transaction throughput by bundling multiple transactions
into a single rollup block. This rollup block is then verified and
added to the Layer 1 blockchain.

The program specifically verifies a single rollup block by interacting
with the Celestia and Ethereum networks. It fetches a transaction from the Celestia network, generates a data root inclusion proof for that transaction,
and then verifies this proof on the BlobstreamX contract on the Ethereum
network. This process ensures the integrity and validity of the rollup block,
providing a crucial component of the security and scalability of the rollup
solution.

This program verifies a transaction using a rollup. It fetches the transaction
from the Celestia network, generates a data root inclusion proof, and verifies
the proof on the BlobstreamX contract on the Ethereum network.

## Prerequisites

- Go (version 1.21 or later)
- Access to a Celestia node (for fetching transactions and blocks)
  - This is optional, only necessary if you want to change the endpoint
  that is used.
- Access to an Ethereum node (for interacting with the BlobstreamX contract)
  - This is optional, only necessary if you want to change the endpoint 
  that is used.
  
## Running this demo

1. [Clone this repository](../README.md#running-these-demos)
and change into the rollup directory:

    ```bash
    cd $HOME/blobstream-demo/rollup
    ```

2. Run the program with the make command:

    ```bash
    make start
    ```

### How it works

The program performs the following steps:

1. Decodes the transaction hash.
2. Establishes a connection to the Celestia network.
3. Fetches the transaction with the decoded hash from Celestia.
4. Fetches the block containing the transaction from Celestia.
5. Generates a data root inclusion proof.
6. Establishes a connection to the Ethereum network.
7. Fetches the BlobstreamX contract.
8. Verifies the data root inclusion on the BlobstreamX contract.

If the verification is successful, the program prints "✅ Verification process completed successfully." If the verification fails, the program prints an error message.
