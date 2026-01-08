CREATE TABLE accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  plaid_item_id UUID NOT NULL REFERENCES plaid_items(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  plaid_account_id TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  official_name TEXT,
  type TEXT NOT NULL,
  subtype TEXT,
  current_balance NUMERIC(12,2),
  available_balance NUMERIC(12,2),
  currency_code TEXT NOT NULL DEFAULT 'USD',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
