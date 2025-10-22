# eth2usd

**eth2usd** is a CLI tool that connects to an Ethereum RPC endpoint, reads token balances for a given address, retrieves real-time USD prices from the Chainlink Feed Registry, and outputs the portfolio valuation in both text and JSON formats.

---

## 🔧 Features

- Fetches on-chain balances (ETH + ERC-20)
- Converts token values to USD via Chainlink oracles
- Supports custom token lists
- Works with any Ethereum RPC (Infura, Alchemy, local node)
- Simple configuration via `.env` file

---

## 🚀 Quick Start

### 1. Clone and build
```bash
git clone https://github.com/yourorg/eth2usd.git
cd eth2usd
go mod tidy && make build
````

### 2. Create `.env`

> Replace `YOUR_INFURA_KEY_HERE` with your actual Infura API key.

```bash
INFURA_KEY="YOUR_INFURA_KEY_HERE"

bash -lc 'cat > .env <<EOF
RPC_URL=https://mainnet.infura.io/v3/'"$INFURA_KEY"'
FEED_REGISTRY=0x47Fb2585D2C56Fe188D0E6ec628a38b74fCeeeDf
ACCOUNT=0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7
TOKENS_FILE=./tokens.json
FORMAT=text
TIMEOUT=30s
EOF
[ -f .gitignore ] || : > .gitignore; grep -qxF ".env" .gitignore || echo ".env" >> .gitignore'
```

### 3. Create `tokens.json`

```bash
bash -lc 'cat > tokens.json <<EOF
[
  { "address": "eth://native", "symbol": "ETH",  "decimals": 18 },
  { "address": "0x6B175474E89094C44Da98b954EedeAC495271d0F", "symbol": "DAI"  },
  { "address": "0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "symbol": "USDC" },
  { "address": "0xdAC17F958D2ee523a2206206994597C13D831ec7", "symbol": "USDT" }
]
EOF'
```

### 4. Verify RPC connection

```bash
bash -lc 'set -a; source .env; set +a; curl -s -X POST "$RPC_URL" -H "Content-Type: application/json" \
--data '"'"'{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'"'"' && echo'
```

### 5. Run test (text output)

```bash
bash -lc 'set -a; source .env; set +a; ./bin/eth2usd \
  --rpc-url "$RPC_URL" \
  --chainlink-registry "$FEED_REGISTRY" \
  --account "$ACCOUNT" \
  --tokens-file "$TOKENS_FILE" \
  --format "$FORMAT" \
  --timeout "$TIMEOUT"'
```

Expected output example:

```
ASSET   AMOUNT          USD          SOURCE
ETH     0               0            chainlink
DAI     73454530.43     73446081.69  chainlink
USDC    76011867.77     76001226.12  chainlink
USDT    26794272.13     26801421.65  chainlink
TOTAL USD: 176248729.46
```

---

## 🔁 Change Address and Re-Test

Now let’s change the address to another known wallet that holds ETH and re-run the tests.

### 1. Update `.env` with a new address (Vitalik's wallet)

```bash
bash -lc 'python3 - <<PY
import re; p=".env"; t=open(p).read()
t=re.sub(r"^ACCOUNT=.*$", "ACCOUNT=0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", t, flags=re.M)
open(p,"w").write(t)
print("ACCOUNT updated to Vitalik.eth")
PY'
```

### 2. Re-run test (text)

```bash
bash -lc 'set -a; source .env; set +a; ./bin/eth2usd \
  --rpc-url "$RPC_URL" \
  --chainlink-registry "$FEED_REGISTRY" \
  --account "$ACCOUNT" \
  --tokens-file "$TOKENS_FILE" \
  --format "$FORMAT" \
  --timeout "$TIMEOUT"'
```

**Expected output example (non-zero ETH balance):**

```
ASSET   AMOUNT      USD        SOURCE
ETH     2500.15     8250000.00 chainlink
TOTAL USD: 8250000.00
```

### 3. Re-run test (JSON)

```bash
bash -lc 'set -a; source .env; set +a; ./bin/eth2usd \
  --rpc-url "$RPC_URL" \
  --chainlink-registry "$FEED_REGISTRY" \
  --account "$ACCOUNT" \
  --tokens-file "$TOKENS_FILE" \
  --format json \
  --timeout "$TIMEOUT" | jq .'
```

**Expected JSON output example (non-zero ETH balance):**

```json
{
  "Rows": [
    {
      "Symbol": "ETH",
      "Amount": "2500.15",
      "USD": "8250000.00",
      "Source": "chainlink",
      "Err": ""
    }
  ],
  "TotalUSD": "8250000.00"
}
```

---

## 🧭 Testing Roadmap

| Step | Goal                              | Command                                                                                                                                |
| ---- | --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | Verify RPC works                  | `curl -X POST "$RPC_URL" -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'` |
| 2    | Build binary                      | `make build`                                                                                                                           |
| 3    | Test single address (Curve 3pool) | `.env` default address                                                                                                                 |
| 4    | Change to Vitalik address         | `ACCOUNT=0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045`                                                                                   |
| 5    | Change output format              | `FORMAT=json` in `.env`                                                                                                                |
| 6    | Save results                      | Add `--out result.txt`                                                                                                                 |
| 7    | Add/remove tokens                 | Edit `tokens.json`                                                                                                                     |

---

## 🧰 Example Useful Addresses

| Name                    | Address                                      | Notes                        |
| ----------------------- | -------------------------------------------- | ---------------------------- |
| Curve 3pool             | `0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7` | Large DAI/USDC/USDT balances |
| Vitalik                 | `0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045` | ETH holder                   |
| Chainlink Feed Registry | `0x47Fb2585D2C56Fe188D0E6ec628a38b74fCeeeDf` | Mainnet registry             |

---

## 🧩 Output Formats

**Text**

```
ASSET   AMOUNT      USD       SOURCE
ETH     0.1234      398.22    chainlink
TOTAL USD: 398.22
```

**JSON**

```json
{
  "Rows": [
    { "Symbol": "ETH", "Amount": "0.1234", "USD": "398.22", "Source": "chainlink" }
  ],
  "TotalUSD": "398.22"
}
```

---