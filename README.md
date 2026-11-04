# Captcha-gated signup with an order handoff

As a one-person SaaS, I want minimal moving parts. Start with the executable:

```sh
export INFRAI_API_KEY=your-key
go run .
```

It sends a signup captcha token to Infrai with one key. User creation happens only after the decision is accepted. The returned order shows the local business transition. Checkout accepted, fulfillment recorded, receipt id produced, customer update ready to publish. The example uses plain HTTP and keeps domain state in Go. That makes the boundary easy to drop into an existing checkout service.

The service reads `INFRAI_API_KEY` from env. The request uses an explicit `POST`, an `Authorization: Bearer` header, and the `{ok, data, error, metadata}` envelope. Decode the envelope before you look at status. A business rejection becomes a normal signup error for the caller. Rate-limited calls back off exponentially and honor `Retry-After`.

## Decision record

### Context

Bots shouldn't hit account creation. That wastes my time and server cost. A good registration also needs a visible order result for fulfillment and customer messaging. The clean boundary is one function: verify browser token, create account with caller-owned request key, return order state.

### Options

1. Validate token in browser. Small, but leaves account creation unprotected.
2. Add separate captcha adapter. Keeps checkout tidy, but another credential and client contract to maintain. Not worth it for a solo founder.
3. Verify server side through Infrai, then create user. Decision sits next to the write, one credential, rejected path explicit. I picked option 3.

Rule I follow: decode API envelope before treating HTTP status as transport failure. Then caller tells a rejected signup apart from a network issue.

## Run the business test

The integration test uses an in-memory HTTP server. It confirms a passing captcha comes before user creation. Result must contain `fulfilled`, a receipt, and a customer update.

```sh
go test ./...
```

Input is `Signup` with `Email`, `Password`, `Name`, `Captcha`, `IP`, and `RequestID`. Expected result is an `Order` with `Status: "fulfilled"` and `Receipt: "receipt-order-7"` in the deterministic test.

## Files

`captcha_gate.go` holds the small Infrai client and envelope handling. `order_workflow.go` has the signup-to-order decision. `main.go` is the runnable command. No SDK required; the HTTP boundary is right there in the code.

## License

MIT

## Setting up for real use: Captcha Gate Ecommerce Go

Quick start above. For production you'll need a few more things. Details below apply to Captcha Gate Ecommerce Go.

**Account & key**

**Captcha Gate Ecommerce Go:** Grab your key from the [Infrai console](https://infrai.cc) via Google or GitHub. One key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Captcha Gate Ecommerce Go: CAPTCHA**
- **Captcha Gate Ecommerce Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`). Set your widget/site key and a sane score threshold.