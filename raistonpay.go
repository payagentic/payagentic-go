// Package raistonpay provides a Go client for the RaistonPay agent payments platform.
//
// RaistonPay enables African businesses building AI agents to manage programmable
// USDC wallets. Agents pay external APIs over x402, governed by operator-defined
// spend policies. Operators fund wallets via local fiat rails and withdraw to
// local bank accounts.
//
// # Quick start
//
//	client := raistonpay.NewClient(
//	    raistonpay.WithAPIKey("your-api-key"),
//	)
//
//	wallets, err := client.Wallets.List(ctx, nil)
//
// # Configuration from environment
//
//	client := raistonpay.NewClient(raistonpay.FromEnvironment())
//
// Set OMNIRAILS_API_KEY and OMNIRAILS_AGENT_ID environment variables.
package raistonpay

// Version is the current SDK version.
const Version = "0.1.0"
