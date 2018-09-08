## VladVM

This is the vlad vm, which executes a json-file describing the prestate, and applies a set of transactions in the order they appear.
See `example.json` for specific syntax.

This awesome VM uses uses `Constantinople` rules!

OBS: The transactions use the same chain id as mainnet, so be careful (because anything can be replayed across them)

The example file contains two identical transactions, after each other. This is fully legit, and demonstrates that the second one is rejected
since it has an invalid nonce.

## Example

When running it, this is the `stderr` output for the example:

```
./vladvm apply example.json
rejected tx: 0x0557bacce3375c98d806609b8d5043072f0b6a8bae45ae5a67a00d3a1a18d673 from 0x8a8eafb1cf62bfbeb1741769dae1a9dd47996192: nonce too low
```

This is the `stdout` output: 
```json
{
  "state": {
    "root": "84208a19bc2b46ada7445180c1db162be5b39b9abc8c0a54b05d32943eae4e13",
    "accounts": {
      "8a8eafb1cf62bfbeb1741769dae1a9dd47996192": {
        "balance": "4276951709",
        "nonce": 1,
        "root": "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
        "codeHash": "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "code": "",
        "storage": {}
      },
      "a94f5374fce5edbc8e2a8697c15331677e6ebf0b": {
        "balance": "6916764286133345652",
        "nonce": 172,
        "root": "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
        "codeHash": "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "code": "",
        "storage": {}
      },
      "c94f5374fce5edbc8e2a8697c15331677e6ebf0b": {
        "balance": "42000",
        "nonce": 0,
        "root": "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
        "codeHash": "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "code": "",
        "storage": {}
      }
    }
  },
  "receipts": [
    {
      "root": "0x",
      "status": "0x1",
      "cumulativeGasUsed": "0x5208",
      "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "logs": null,
      "transactionHash": "0x0557bacce3375c98d806609b8d5043072f0b6a8bae45ae5a67a00d3a1a18d673",
      "contractAddress": "0x0000000000000000000000000000000000000000",
      "gasUsed": "0x5208"
    }
  ],
  "rejected": [
    "0x0557bacce3375c98d806609b8d5043072f0b6a8bae45ae5a67a00d3a1a18d673"
  ]
}
```

## Notes 

Things of note:

- No 'mining-reward' is applied per se, aside from the transaction fees
- Info about rejected transactions wind up on `stderr`, to not 'pollute' the `json` which can be reused for the next state transition
- and, again: TRANSACTIONS USE THE SAME CHAINID AS MAINET - REPLAY IS FULLY POSSIBLE

This written for ethberlin hackathon, do not use this in real life. 
