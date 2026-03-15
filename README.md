# Pocket

Pocket is a personal finance app: connect a bank account (via Plaid) or upload a CSV statement, and it gives you a running net position, a categorized transaction ledger, per-category budgets, and a chat assistant that can answer questions about your own spending. The backend is a Go API over Postgres; the frontend is a Next.js app; transaction categorization and the chat assistant are backed by Google's Gemini.

## Features

- **Accounts** — connect real banks through Plaid Link (sandbox by default) or upload a CSV bank statement. Net position correctly nets out liabilities (credit cards, loans, mortgages) against assets. Plaid-linked connections can be unlinked from the Accounts page (with a confirmation step), which removes every account and transaction under that connection.
- **Transactions** — a paginated, filterable ledger (by account or category) with consistent income/expense sign conventions across every view.
- **Budgets** — per-category monthly limits with spent/remaining tracking. Categories are restricted to the fixed taxonomy transactions actually use, so a budget can never sit at zero because of a typo or a category no transaction will ever match.
- **Ask Pocket** — a chat assistant (Gemini) with access to your accounts, transactions, and budget data, so you can ask things like "how much did I spend on dining this month?"
- **Demo mode** — a shared, reset-able demo account (`View demo` on the login page) seeded with seven months of realistic transaction history, for trying the app without connecting a real bank or uploading a file.

## Architecture

```
frontend/   Next.js 16 (App Router, Turbopack, React 19, SWR for data fetching)
internal/   Go 1.26 API (net/http, database/sql, no framework)
  api/          HTTP handlers, middleware (auth, CORS, rate limiting, logging)
  db/           Postgres connection + migrations (embedded, run automatically on startup)
  plaid/        Plaid client, sync, transaction categorization
  ai/           Gemini client for the chat assistant
  models/       Shared domain types
cmd/server/ Entry point
```

The frontend talks to the Go API over plain HTTP with a bearer JWT (no server-side rendering of authenticated data — auth state lives in `localStorage`). The Go API owns all business logic: sign conventions, categorization, budget math, and net-position calculation are computed once on the backend and never recomputed differently in the frontend.

## Prerequisites

- Go 1.26+
- Node 20+ and pnpm
- Postgres 14+ (running locally, or a connection string to a hosted instance)
- A [Plaid](https://dashboard.plaid.com/signup) sandbox account (free) — only needed to test the "connect a bank" flow; CSV import and demo mode work without it
- A [Gemini API key](https://aistudio.google.com/apikey) — only needed for the "Ask Pocket" chat feature

## Setup

```sh
git clone <this repo>
cd pocket-api

# Backend config
cp .env.example .env
# edit .env: set DATABASE_URL, JWT_SECRET (openssl rand -base64 32), and
# Plaid/Gemini keys if you want those features

# Frontend config
cp frontend/.env.example frontend/.env.local
# defaults to http://localhost:8080, which matches the backend's default port

# Install dependencies
go mod download
pnpm --dir frontend install
```

Create the database referenced by `DATABASE_URL` (e.g. `createdb pocket`) — migrations run automatically the first time the server starts.

### Running locally

```sh
./dev.sh
```

This is the canonical way to run Pocket locally — one command, both servers, in the right order. It starts the backend on `:8080`, waits for its health check, then starts the frontend on `:3000`. Logs go to `/tmp/pocket-backend.log` and `/tmp/pocket-frontend.log`. Ctrl+C stops both. Prefer this over starting the two servers separately by hand, since it's easy to bring one up and forget the other.

You can still run them separately if you need independent control of each (see `Procfile` for the same two commands, usable with `honcho start` / `goreman start` if you have one installed):

```sh
# terminal 1
go run ./cmd/server

# terminal 2
pnpm --dir frontend dev
```

Both are plain foreground processes — nothing restarts them if the terminal closes.

### Production build

```sh
pnpm --dir frontend build
pnpm --dir frontend start
```

### Tests

```sh
go test ./...              # Go unit + integration tests
cd frontend
npx tsc --noEmit            # type check
pnpm build                  # production build (also catches build-time errors)
```

Some Go tests (`internal/db/queries/reconciliation_test.go`) connect to the real database in `DATABASE_URL` to exercise actual SQL aggregation logic; they skip automatically if `DATABASE_URL` isn't set.

## Demo account

Click "View demo" on the login page. This logs into (or registers, on first use) a single shared account (`demo@example.com`) seeded with ~85 days of recurring and randomized transactions across every budget category. "Load demo data" repopulates it; "Reset demo" wipes it. These two operations are locked to that one email server-side — no other account can trigger them, even by calling the endpoint directly.

## CSV import format

Upload a CSV bank statement from the Accounts page. Column headers are matched case-insensitively against common aliases, so most bank exports work without renaming columns:

| Field | Required | Recognized headers |
|---|---|---|
| Date | yes | date, transaction date, posted date, posting date, trans date |
| Description | yes | description, details, memo, payee, narrative, name, transaction, reference, merchant, vendor |
| Amount | one of amount, or debit+credit | amount, value, transaction amount |
| Debit | | debit, withdrawal, money out, paid out |
| Credit | | credit, deposit, money in, paid in |
| Category | no | category, type |

- **Dates**: most common formats are accepted (`YYYY-MM-DD`, `MM/DD/YYYY`, `DD/MM/YYYY`, `Jan 2, 2006`, `2006/01/02`, etc.).
- **Amount sign convention**: if your file has a single "amount" column, it must follow the standard bank-statement convention — **negative = money out, positive = money in**. If your file has separate debit/credit columns instead, no sign is needed; the columns are unambiguous.
- **Category**: if your file has no category column, Pocket infers one from the merchant/description text. If it does, values are mapped onto Pocket's fixed category taxonomy (see below) — recognized values map directly (e.g. "gas" → Transport, "rent" → Housing); anything unrecognized keeps your original label rather than collapsing to "Other".
- Re-uploading the same file is safe — rows are deduplicated by content (account + date + description + amount + occurrence), so it won't double your transactions or your account balance.
- Skipped rows (bad dates, missing amounts) are reported individually after upload, numbered by transaction (not raw file line).

The canonical category taxonomy — used by CSV import, Plaid sync, and budgets alike — is: Groceries, Dining, Shopping, Transport, Travel, Subscriptions, Entertainment, Health, Housing, Income, Transfers, Other.

## Deployment checklist

This repo only contains app config, not a hosting setup — pick your own host for the Go binary, the Next.js app, and Postgres. Before deploying:

1. **Database**: provision Postgres, set `DATABASE_URL` to point at it. Migrations run automatically on server startup — no separate migration step needed, but the server must be able to reach the database before it will start.
2. **Backend env vars** (see `.env.example`): `JWT_SECRET` (required — the server refuses to start without one; generate with `openssl rand -base64 32`), `PLAID_CLIENT_ID` / `PLAID_SECRET` / `PLAID_ENV=production` (for real bank connections), `GEMINI_API_KEY` (for chat), `CORS_ALLOWED_ORIGIN` (set to your real frontend origin — defaults to `*`, which is fine for local dev but should be locked down in production).
3. **Frontend env vars**: `NEXT_PUBLIC_API_URL` set to your deployed backend's public URL.
4. **Build and run**: `pnpm --dir frontend build && pnpm --dir frontend start` for the frontend; `go build -o pocket-server ./cmd/server && ./pocket-server` (or `go run ./cmd/server`) for the backend.
5. Confirm the deployed frontend can reach the deployed backend (check `GET /health`) and that a real signup → login round-trip works before switching Plaid to production mode.

## Known limitations

- **Rate limiting is in-memory and per-process** (`internal/api/middleware/ratelimit.go`). It resets on restart and isn't shared across instances — fine for a single-instance deployment, not correct if you horizontally scale the backend without moving this to a shared store (e.g. Redis).
- **The demo account is shared** — anyone who clicks "View demo" logs into the same account and sees whatever the last person left there (or resets it). It's meant for quick evaluation, not multi-user demoing.
- **Plaid OAuth institutions** (banks requiring a redirect step) aren't configured — Link works with sandbox and most non-OAuth institutions, but a redirect URI would need to be added for full OAuth bank support.
- **Categories are fixed**, not user-customizable. This is intentional (it's what makes budget-to-transaction matching reliable), but it means you can't add your own category names.
- **Transaction category edits in the UI are local-only** — the transactions page lets you preview a category change, but there's no backend endpoint yet to persist it.
