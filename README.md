# PayAgentic Go SDK

Build AI agents that buy API data using PayAgentic's programmable stablecoin wallets, spend policies and x402 payment APIs.

[Package reference](https://pkg.go.dev/github.com/payagentic/payagentic-go) · [Developer home](https://payagentic.ai/platform/developers) · [SDK documentation](https://payagentic.ai/sdks#go) · [Pricing](https://payagentic.ai/pricing) · [Integration support](mailto:developers@payagentic.ai)

## Install

Requires Go 1.24.3 or later; see `go.mod` for this release's toolchain requirements.

```sh
mkdir payagentic-example
cd payagentic-example
go mod init example/payagentic
go get github.com/payagentic/payagentic-go
```

Pin the installed version in your application's `go.mod` and commit `go.sum`. This is a pre-1.0 SDK: review and test changes before upgrading.

## Check your connection

Set `PAYAGENTIC_API_KEY` and `PAYAGENTIC_BASE_URL` through your server's runtime environment or secret manager. The gateway URL must match your deployment; do not rely on the SDK default without checking it.

Save the following as `main.go` and run `go run .`. It performs a read-only wallet request and prints the HTTP status, without printing wallet data or the API key.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    payagentic "github.com/payagentic/payagentic-go"
)

func main() {
    baseURL := os.Getenv("PAYAGENTIC_BASE_URL")
    if baseURL == "" { log.Fatal("Set PAYAGENTIC_BASE_URL") }
    client, err := payagentic.NewClient(payagentic.WithBaseURL(baseURL))
    if err != nil { log.Fatal("Check SDK configuration") }
    defer client.Close()
    result, err := client.OpenAPI.ListWalletsWithResponse(context.Background(), nil)
    if err != nil { log.Fatal("Gateway request failed") }
    fmt.Println("Gateway HTTP status:", result.StatusCode())
}
```

Expected output with a working gateway and authorized key:

```text
Gateway HTTP status: 200
```

The same complete program is in [examples/connection/main.go](examples/connection/main.go). From a checkout of this repository, run `go run ./examples/connection`.

## Try the buyer and merchant workflow

The [runnable examples guide](https://payagentic.ai/docs/examples) provides a downloadable Node.js buyer and merchant demonstration with setup instructions, expected output and tests. It illustrates the HTTP exchange for developers in any language. Payment authorization is simulated; no funds move.

For actual payments, follow the [quickstart](https://payagentic.ai/docs/quickstart), finish wallet provisioning and configure a funded test wallet and spend policy. Merchants must register their API and endpoint and verify origin ownership to test the registered merchant transaction fee flow. An HTTP 200 response alone does not prove settlement; reconcile gateway transaction status before recording it.

## Troubleshooting and compatibility

- Missing key or configuration error: make the runtime variables available to the process running the example.
- DNS or connection failure: verify `PAYAGENTIC_BASE_URL` with your deployment operator.
- HTTP 401 or 403: check the key's environment and permissions. Do not post the key in an issue or paste it into browser code.
- Documentation mentions a method absent from your installed SDK: confirm the gateway and SDK release versions before changing your integration.

This public repository is a release mirror of the private PayAgentic monorepo. Public releases can lag source development; website examples may target newer source builds. The download guide identifies source-built examples separately from registry releases. Use this repository and pkg.go.dev to inspect the API actually installed.

The SDK uses the MIT license (see [LICENSE](LICENSE)). Platform subscriptions and transaction fees are separate; see [pricing](https://payagentic.ai/pricing).
