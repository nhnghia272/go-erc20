# Golang ERC20

## Download from Github
From your project directory:
```
go get github.com/nhnghia272/goerc20
```

## Example
```go
package main

import (
	"github.com/nhnghia272/goerc20"
)

func main() {
	// Initialize
	erc20 := goerc20.New(goerc20.Config{ChainRpc: "CHAIN_RPC", Contract: "YOUR_CONTRACT_ADDRESS", Decimals: 18, PrivateKey: "YOUR_PRIVATE_KEY"})

	// Send Native Token
	erc20.SendTo("TO_ADDRESS", 1)

	// Send ERC20 Token
	erc20.Erc20SendTo("TO_ADDRESS", 1)

	// Send ERC20 Token
	erc20.Erc20SendFrom("OWNER_ADDRESS", "TO_ADDRESS", 1)
}
```