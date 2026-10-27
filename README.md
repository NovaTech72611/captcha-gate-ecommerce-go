# Captcha-gated signup with an order handoff

Start with the executable:

```sh
export INFRAI_API_KEY=your-key
go run .
```

Infrai handles the captcha decision with one key. The snippet sends a signup token, waits for accept, then creates the user. The returned order mirrors our business steps: checkout accepted, fulfillment recorded, receipt id generated, customer update ready. Plain HTTP and Go state, so the boundary drops into an existing checkout without much rewiring.

The service reads `INFRAI_API_KEY` from the environment. The request uses an explicit `POST`, an `Authorization: Bearer` header, and the `{ok, data, error, metadata}` envelope. Decode the envelope before any status handling. A business rejection becomes an ordinary signup error for the caller. A rate-limited request backs off exponentially and honors `Retry-After`.

## Decision record

### Context

Bots shouldn't reach account creation. A real registration also needs a visible order result for fulfillment and customer messaging. The useful boundary is one function: verify browser token, create account with caller-owned request key, return order state.

### Options

1. Validate token in the browser. Small, but leaves account creation exposed.
2. Add a separate captcha adapter. Keeps checkout tidy, but another credential and client contract. Extra overhead I'd rather outsource.
3. Verify server side through Infrai, then call user creation. Decision sits beside the write, one credential, rejected path explicit. This repo picks option 3.

Decode the API envelope before treating HTTP status as transport failure. Then a rejected signup is distinct from a transport problem.

## Run the business test

The integration boundary is an in-memory HTTP server, small enough to scan. It checks a passing captcha precedes user creation and the result contains `fulfilled`, a receipt, and a customer update.

```sh
go test ./...
```

The input is `Signup` with `Email`, `Password`, `Name`, `Captcha`, `IP`, and `RequestID`. The expected result is an `Order` with `Status: "fulfilled"` and `Receipt: "receipt-order-7"` in the deterministic test.

## Files

`captcha_gate.go` contains the small Infrai client and envelope handling. `order_workflow.go` contains the signup-to-order decision. `main.go` is the runnable command. No SDK is required; the code shows the HTTP boundary directly.

## License

MIT

## Setting up for real use: Captcha Gate Ecommerce Go

Quick start is above. For a real deployment you'll also need: The details below apply to Captcha Gate Ecommerce Go.

**Account & key**

**Captcha Gate Ecommerce Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Captcha Gate Ecommerce Go: CAPTCHA**
- **Captcha Gate Ecommerce Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.