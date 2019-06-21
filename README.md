# Stellar withdraw integration module

## Usage

Enviromental variable `KV_VIPER_FILE` must be set and contain path to desired config file.

```bash
stellar-withdraw-svc run withdraw
```

## Watchlist

In order for service to start watching withdrawals in specific asset, asset details in TokenD must have entry of the following form: 
```json
"stellar": {
   "withdraw": true, //
   "asset_code": "USD", // Omit for asset type "native"
   "asset_type": "AlphaNum4",
   },
```
Service will only listen for withdraw requests with `2048` pending tasks flag set and `4096` flag not set.
So, either value by key `withdrawal_tasks:*`, or `withdrawal_tasks:ASSET_CODE`  must contain `2048` flag and must not contain flag `4096`.

## Config

```yaml
stellar:
  is_testnet: true

horizon:
  endpoint:
  signer: "S_SOME_VALID_SECRET_KEY" # Issuer of assets

payment:
  source_signer: "S_SOME_VALID_STELLAR_SECRET_KEY" # Payment sender signer
  source_address: "G_SOME_VALID_STELLAR_ADDRESS" # Payment sender address

withdraw:
  signer: "S_SOME_VALID_SECRET_KEY"
  owner: "G_SOME_VALID_ADDRESS"

log:
  level: debug
  disable_sentry: true
```