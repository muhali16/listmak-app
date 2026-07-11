# Pakasir Integration Guide

**Last Updated:** 03 Jul 2026  
**Service:** Payment link solution (QRIS + Virtual Account)  
**Operator:** PT. Geksa (Bank Indonesia licensed)

---

## Table of Contents

1. [Setup](#setup)
2. [Integration Methods](#integration-methods)
3. [URL-Based Integration](#url-based-integration)
4. [API Integration](#api-integration)
5. [Webhook Handling](#webhook-handling)
6. [Transaction Status Queries](#transaction-status-queries)
7. [Plugins & Packages](#plugins--packages)
8. [Implementation Checklist](#implementation-checklist)

---

## Setup

### A.1 Create a Project

1. Register/login to Pakasir dashboard
2. Create a new Project (one per application/website)
3. **Note these values** — you'll need them for all integrations:
   - `slug` — project identifier
   - `api_key` — for API requests & webhook validation

Each project in Pakasir is independent, allowing you to manage multiple applications under one account.

---

## Integration Methods

Pakasir offers three integration approaches:

| Method | Use Case | Complexity |
|--------|----------|-----------|
| **URL Redirect** | Simple payment links, redirect user to hosted payment page | Low |
| **API + Webhook** | Full control, custom UI, background transactions | High |
| **Plugins** | WordPress/WooCommerce, EDD, FluentCart | Medium |

---

## URL-Based Integration

### Basic Payment Link

Redirect customers to the payment page using this URL:

```
https://app.pakasir.com/pay/{slug}/{amount}?order_id={order_id}
```

**Parameters:**
- `{slug}` — Your project slug
- `{amount}` — Transaction amount in IDR, no decimals/spaces (e.g., `100000` or `25000`)
- `{order_id}` — Your internal transaction/invoice ID (e.g., `INV20240910-123456` or `1298`)

**Example:**
```
https://app.pakasir.com/pay/depodomain/22000?order_id=240910
```

### Option: Custom Redirect (Callback URL)

After successful payment, user sees a "Return to Merchant" button. By default, they return to previous page.

To redirect to a specific URL:

```
https://app.pakasir.com/pay/depodomain/22000?order_id=240910&redirect=https://websitekamu.com/success
```

⚠️ **Note:** This redirect is **not secure** as a sole verification method. Always validate payment status via webhook or API query (Section 5–6).

### Option: QRIS-Only Display

Force QRIS payment method only (user cannot switch to bank transfer):

```
https://app.pakasir.com/pay/depodomain/22000?order_id=240910&qris_only=1
```

---

## API Integration

### Prerequisites

- API Key from project settings
- HTTPS endpoint for webhook receiver (required for production)
- Idempotency handling in your system

### C.1 API Response Structure

When you call Pakasir API, they send you:
1. **QR String** or Virtual Account Number
2. **Total Payment Amount** (includes fee)
3. **Expiry Time** (deadline for payment)

Your responsibility: Display QR code/account info to customer and track payment status.

### C.2 Create Transaction

**Endpoint:** `POST https://app.pakasir.com/api/transactioncreate/{method}`

**Request Body:**
```json
{
  "project": "depodomain",
  "order_id": "INV123123",
  "amount": 99000,
  "api_key": "xxx123"
}
```

**cURL Example:**
```bash
curl -L 'https://app.pakasir.com/api/transactioncreate/qris' \
  -H 'Content-Type: application/json' \
  -d '{
    "project": "depodomain",
    "order_id": "INV123123",
    "amount": 99000,
    "api_key": "xxx123"
  }'
```

**Response (Success):**
```json
{
  "payment": {
    "project": "depodomain",
    "order_id": "INV123123",
    "amount": 99000,
    "fee": 1003,
    "total_payment": 100003,
    "payment_method": "qris",
    "payment_number": "00020101021226610016ID.CO.SHOPEE",
    "expired_at": "2025-09-19T01:18:49.678622564Z"
  }
}
```

⚠️ **Implementation Notes:**
- Fee is automatically added by Pakasir. Decide: absorb it or pass to customer?
- `expired_at` is when payment link expires. Store this in your DB.
- `payment_number` is the QR string — use a QR library to render it as image
- Always save transaction details locally before showing to customer

### C.3 Available Payment Methods

```
- cimb_niaga_va
- bni_va
- qris
- sampoerna_va
- bnc_va
- maybank_va
- permata_va
- atm_bersama_va
- artha_graha_va
- bri_va
```

For each method, create separate transactions or detect customer preference.

### C.4 Payment Simulation (Sandbox Only)

Test webhook in sandbox mode:

**Endpoint:** `POST https://app.pakasir.com/api/paymentsimulation`

**Request:**
```json
{
  "project": "depodomain",
  "order_id": "INV123123",
  "amount": 99000,
  "api_key": "xxx123"
}
```

**cURL:**
```bash
curl -L 'https://app.pakasir.com/api/paymentsimulation' \
  -H 'Content-Type: application/json' \
  -d '{
    "project": "depodomain",
    "order_id": "INV123123",
    "amount": 99000,
    "api_key": "xxx123"
  }'
```

Use this to test your webhook handler before production.

### C.5 Cancel Transaction

Cancel a pending transaction (before payment):

**Endpoint:** `POST https://app.pakasir.com/api/transactioncancel`

**Request:**
```json
{
  "project": "depodomain",
  "order_id": "INV123123",
  "amount": 99000,
  "api_key": "xxx123"
}
```

**cURL:**
```bash
curl -L 'https://app.pakasir.com/api/transactioncancel' \
  -H 'Content-Type: application/json' \
  -d '{
    "project": "depodomain",
    "order_id": "INV123123",
    "amount": 99000,
    "api_key": "xxx123"
  }'
```

⚠️ **Critical:** Only cancel if transaction is still pending. Once payment arrives, use manual refund via dashboard.

---

## Webhook Handling

### How It Works

When customer payment succeeds and funds enter Pakasir's system, they send an HTTP POST to your webhook URL with payment details.

**Register Webhook URL:** Dashboard → Edit Project → Webhook URL field

### Webhook Payload

```json
{
  "amount": 22000,
  "order_id": "240910HDE7C9",
  "project": "depodomain",
  "status": "completed",
  "payment_method": "qris",
  "completed_at": "2024-09-10T08:07:02.819+07:00"
}
```

### Required Validations (⚠️ Security Critical)

```plaintext
1. Verify amount matches your internal transaction record
2. Verify order_id matches your internal transaction record
3. Verify project slug matches your configured project
4. Idempotency: Check if order_id already processed → only process once
5. ⚠️ NO WEBHOOK SIGNATURE PROVIDED — Assume all webhooks could be spoofed
```

**Important:** Pakasir's official docs recommend using their Transaction Detail API (Section 6) for definitive status, not just webhook trust.

### Implementation Strategy

```pseudocode
webhook_handler(payload):
  // Step 1: Quick validation
  if payload.project != EXPECTED_PROJECT:
    return 400  // Reject wrong project
  
  if not order_exists(payload.order_id):
    return 404  // Unknown order
  
  // Step 2: Database lock/transaction
  db.transaction {
    order = db.find(payload.order_id)
    
    // Idempotency check
    if order.status == "completed":
      return 200  // Already processed, don't double-charge
    
    // Step 3: Verify amount
    if order.total_amount != payload.amount:
      log.error("Amount mismatch", order.order_id, payload)
      return 400  // Amount tampering detected
    
    // Step 4: Query Pakasir API for definitive status
    api_response = pakasir_api.get_transaction_detail(
      project: payload.project,
      order_id: payload.order_id
    )
    
    if api_response.status != "completed":
      log.warn("Webhook says completed, API says not completed")
      return 202  // Accepted but not processed yet
    
    // Step 5: Mark as completed
    order.status = "completed"
    order.payment_method = payload.payment_method
    order.completed_at = payload.completed_at
    order.save()
    
    // Step 6: Trigger business logic (send email, fulfill order, etc)
    order.fulfill()
  }
  
  return 200  // Success
```

### Webhook Best Practices

- **Use database transaction** to prevent race conditions
- **Idempotency key** = order_id (treat duplicate webhooks as safe)
- **Timeout:** Respond within 5–10 seconds; use background job for long operations
- **Retry logic:** Assume Pakasir may retry if you don't return 200
- **Logging:** Log all webhook attempts for debugging

---

## Transaction Status Queries

### Query Transaction Detail

For definitive payment status, query Pakasir API:

**Endpoint:** `GET https://app.pakasir.com/api/transactiondetail?project={project}&order_id={order_id}&api_key={api_key}`

**cURL Example:**
```bash
curl 'https://app.pakasir.com/api/transactiondetail?project=depodomain&order_id=240910HDE7C9&api_key=xxx123'
```

**Response:**
```json
{
  "transaction": {
    "amount": 22000,
    "order_id": "240910HDE7C9",
    "project": "depodomain",
    "status": "completed",
    "payment_method": "qris",
    "completed_at": "2024-09-10T08:07:02.819+07:00"
  }
}
```

### Use Cases for Status Query

1. **Webhook verification** — Double-check webhook data against API
2. **Webhook timeout** — If webhook doesn't arrive, query manually after waiting
3. **Polling fallback** — If customer asks "where's my payment?", query API instead of database
4. **Reconciliation jobs** — Scheduled background task to sync unpaid orders

### Implementation Example (Polling Fallback)

```pseudocode
check_payment_status(order_id):
  // First, check local DB status
  order = db.find(order_id)
  if order.status == "completed":
    return "paid"
  
  // If not completed yet, ask Pakasir
  api_response = pakasir_api.get_transaction_detail(order_id)
  
  if api_response.status == "completed":
    // Update our DB
    order.status = "completed"
    order.save()
    return "paid"
  else:
    return "pending"
```

---

## Implementation Checklist

### Pre-Development

- [ ] Pakasir account created
- [ ] Project slug & API Key obtained and stored in `.env`
- [ ] Decision made: absorb Pakasir fee or pass to customer?
- [ ] Decision made: URL redirect or full API integration?
- [ ] HTTPS endpoint prepared for webhook receiver

### During Development (API Integration)

- [ ] Create transaction endpoint built
- [ ] QR code rendering library added (e.g., `qrcode.js`, `qr-code`)
- [ ] Webhook receiver endpoint built with:
  - [ ] Request validation (project, order_id, amount)
  - [ ] Idempotency handling (no double-processing)
  - [ ] Database locking during processing
  - [ ] Backup API query for status verification
  - [ ] Error logging
- [ ] Transaction detail query endpoint built
- [ ] Payment simulation tested in sandbox mode

### Before Production

- [ ] Webhook URL registered in Pakasir dashboard
- [ ] API key stored securely (not in code, use env vars)
- [ ] Test with live Pakasir account:
  - [ ] Create transaction via API
  - [ ] Make actual payment (QRIS or VA)
  - [ ] Verify webhook received & processed
  - [ ] Verify order status updated in your DB
- [ ] Monitoring/alerting setup:
  - [ ] Failed webhooks logged
  - [ ] Amount mismatches alerted
  - [ ] Pending payments older than expiry alerted
- [ ] Error handling for edge cases:
  - [ ] Webhook arrives before transaction created
  - [ ] Webhook arrives multiple times
  - [ ] Customer cancels payment mid-process
  - [ ] API timeout during transaction creation

### Post-Production

- [ ] Monitoring dashboard for webhook health
- [ ] Daily reconciliation job (query Pakasir for unpaid orders > 24h)
- [ ] Refund workflow documented & tested
- [ ] Support runbook for payment issues

---

## Security Notes

### API Key Protection

- Never commit `api_key` to version control
- Use environment variables: `process.env.PAKASIR_API_KEY`
- Rotate keys periodically
- Use different keys for sandbox vs production

### Webhook Spoofing Risk

⚠️ **Pakasir does not provide request signing/HMAC verification.**

Mitigations:
1. Verify `amount` & `order_id` match your DB
2. Query Pakasir API to confirm status (authoritative source)
3. Only mark order complete after API verification
4. Log all discrepancies for audit

### Amount Tampering

- Never trust webhook amount alone
- Always compare against your stored order amount
- If mismatch detected, log and alert (potential fraud)

### Idempotency

- Webhook may arrive multiple times
- Use order_id as idempotency key in your DB
- Check status before processing
- Use database transaction to prevent duplicates

---

## Common Issues & Solutions

### Issue: Webhook Not Received

1. Check webhook URL in dashboard (must be HTTPS, reachable)
2. Verify firewall/WAF not blocking Pakasir IPs
3. Check server logs for incoming requests
4. Test with payment simulation API (Section C.4)
5. Implement polling fallback (Section 6)

### Issue: Amount Mismatch

1. Verify fee calculation: `total = amount + fee`
2. Compare local amount vs webhook amount
3. Check if currency conversion issue (Pakasir uses IDR)

### Issue: Webhook Timeout

1. Move long operations to background job
2. Respond 200 immediately, process asynchronously
3. Implement retry logic with exponential backoff

### Issue: Payment Completed But Order Not Updated

1. Check webhook logs (was webhook received?)
2. Query Pakasir API manually for payment status
3. If API shows completed but DB doesn't, re-run webhook handler
4. Check for database locks/deadlocks

---

## References

- **Official Docs:** https://pakasir.com/p/docs
- **Service Type:** Payment link (QRIS + Virtual Account)
- **Operator:** PT. Geksa (Bank Indonesia licensed)
- **Support:** Check Pakasir dashboard for contact info

---

## Changelog

| Date | Change |
|------|--------|
| 2026-07-11 | Initial conversion to Markdown with backend implementation notes |
