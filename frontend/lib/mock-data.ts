import type { Account, Budget, Summary, Transaction } from './api'

export const MOCK_SUMMARY: Summary = {
  totalBalance: 22_313.11,
  monthlySpend: 2_847.40,
  lastMonthSpend: 2_541.18,
  transactionCount: 47,
}

export const MOCK_ACCOUNTS: Account[] = [
  {
    id: 'chk_001',
    name: 'Chase Checking ••4821',
    type: 'checking',
    institution: 'Chase',
    balance: 4_821.33,
    lastSynced: '2026-07-04T10:00:00Z',
  },
  {
    id: 'sav_001',
    name: 'Ally High-Yield Savings',
    type: 'savings',
    institution: 'Ally',
    balance: 18_340.00,
    lastSynced: '2026-07-04T10:00:00Z',
  },
  {
    id: 'cc_001',
    name: 'Amex Gold ••9012',
    type: 'credit',
    institution: 'American Express',
    balance: -847.22,
    lastSynced: '2026-07-03T22:00:00Z',
  },
]

export const MOCK_TRANSACTIONS: Transaction[] = [
  { id: 'txn_001', accountId: 'chk_001', date: '2026-07-04', merchant: 'Whole Foods Market', category: 'Groceries', amount: -94.38 },
  { id: 'txn_002', accountId: 'cc_001',  date: '2026-07-03', merchant: 'Clio Restaurant', category: 'Dining', amount: -76.50 },
  { id: 'txn_003', accountId: 'chk_001', date: '2026-07-03', merchant: 'Netflix', category: 'Subscriptions', amount: -15.99 },
  { id: 'txn_004', accountId: 'chk_001', date: '2026-07-02', merchant: 'Direct Deposit — Employer', category: 'Income', amount: 3_100.00 },
  { id: 'txn_005', accountId: 'cc_001',  date: '2026-07-02', merchant: 'Shell', category: 'Gas', amount: -62.40 },
  { id: 'txn_006', accountId: 'cc_001',  date: '2026-07-01', merchant: 'Amazon', category: 'Shopping', amount: -43.99 },
  { id: 'txn_007', accountId: 'chk_001', date: '2026-07-01', merchant: 'Starbucks', category: 'Coffee', amount: -7.25 },
  { id: 'txn_008', accountId: 'chk_001', date: '2026-06-30', merchant: 'Rent — July', category: 'Housing', amount: -2_100.00 },
  { id: 'txn_009', accountId: 'chk_001', date: '2026-06-29', merchant: 'Venmo — Alex K.', category: 'Transfer', amount: 120.00 },
  { id: 'txn_010', accountId: 'cc_001',  date: '2026-06-29', merchant: 'CVS Pharmacy', category: 'Health', amount: -28.14 },
  { id: 'txn_011', accountId: 'cc_001',  date: '2026-06-28', merchant: "Trader Joe's", category: 'Groceries', amount: -67.44 },
  { id: 'txn_012', accountId: 'chk_001', date: '2026-06-27', merchant: 'Spotify', category: 'Subscriptions', amount: -10.99 },
  { id: 'txn_013', accountId: 'cc_001',  date: '2026-06-26', merchant: 'Sweetgreen', category: 'Dining', amount: -18.75 },
  { id: 'txn_014', accountId: 'chk_001', date: '2026-06-25', merchant: 'MBTA CharlieCard', category: 'Transit', amount: -25.00 },
  { id: 'txn_015', accountId: 'cc_001',  date: '2026-06-24', merchant: 'REI', category: 'Shopping', amount: -134.95 },
  { id: 'txn_016', accountId: 'chk_001', date: '2026-06-23', merchant: 'Duane Reade', category: 'Health', amount: -21.50 },
  { id: 'txn_017', accountId: 'cc_001',  date: '2026-06-22', merchant: 'United Airlines', category: 'Travel', amount: -312.00 },
  { id: 'txn_018', accountId: 'chk_001', date: '2026-06-20', merchant: 'Direct Deposit — Employer', category: 'Income', amount: 3_100.00 },
  { id: 'txn_019', accountId: 'cc_001',  date: '2026-06-19', merchant: "Murray's Cheese", category: 'Groceries', amount: -33.60 },
  { id: 'txn_020', accountId: 'chk_001', date: '2026-06-18', merchant: 'Con Edison', category: 'Utilities', amount: -87.44 },
]

export const MOCK_BUDGETS: Budget[] = [
  { id: 'bud_001', category: 'Housing',       limit: 2_200, spent: 2_100 },
  { id: 'bud_002', category: 'Groceries',     limit: 400,   spent: 309  },
  { id: 'bud_003', category: 'Dining',        limit: 300,   spent: 247  },
  { id: 'bud_004', category: 'Transportation',limit: 150,   spent: 87   },
  { id: 'bud_005', category: 'Subscriptions', limit: 60,    spent: 52   },
  { id: 'bud_006', category: 'Shopping',      limit: 200,   spent: 179  },
  { id: 'bud_007', category: 'Health',        limit: 100,   spent: 49   },
  { id: 'bud_008', category: 'Coffee',        limit: 40,    spent: 22   },
]
