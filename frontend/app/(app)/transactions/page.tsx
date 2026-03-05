'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { Check, ChevronDown, X } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api, type Transaction } from '@/lib/api'
import { MOCK_ACCOUNTS, MOCK_TRANSACTIONS } from '@/lib/mock-data'
import { LedgerAmount } from '@/components/amount'

const ALL_CATEGORIES = [
  'All categories',
  'Coffee', 'Dining', 'Gas', 'Groceries', 'Health',
  'Housing', 'Income', 'Shopping', 'Subscriptions', 'Transfer',
  'Transit', 'Travel', 'Utilities',
]

function formatDate(iso: string) {
  return new Date(iso + 'T00:00:00').toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })
}

interface EditCategoryRowProps {
  tx: Transaction
  onSave: (id: string, newCategory: string) => void
  onCancel: () => void
}

function EditCategoryRow({ tx, onSave, onCancel }: EditCategoryRowProps) {
  const [value, setValue] = useState(tx.category)
  return (
    <div className="bg-muted px-3 py-3">
      <p className="mb-2 text-xs text-muted-foreground">
        Change category for <strong className="text-foreground">{tx.merchant}</strong>
      </p>
      <div className="flex gap-2">
        <div className="relative flex-1">
          <select
            value={value}
            onChange={e => setValue(e.target.value)}
            className="w-full appearance-none rounded border border-input bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
          >
            {ALL_CATEGORIES.filter(c => c !== 'All categories').map(c => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          <ChevronDown size={13} className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
        </div>
        <button
          onClick={() => onSave(tx.id, value)}
          className="flex items-center gap-1.5 rounded bg-primary px-3 py-2 text-xs font-medium text-primary-foreground transition-all duration-150 hover:opacity-90"
        >
          <Check size={12} /> Save
        </button>
        <button
          onClick={onCancel}
          className="flex items-center gap-1.5 rounded border border-border px-3 py-2 text-xs text-muted-foreground transition-all duration-150 hover:text-foreground"
        >
          <X size={12} /> Cancel
        </button>
      </div>
    </div>
  )
}

export default function TransactionsPage() {
  const { token, isDemo } = useAuth()
  const [accountFilter, setAccountFilter] = useState('all')
  const [categoryFilter, setCategoryFilter] = useState('All categories')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [localCategories, setLocalCategories] = useState<Record<string, string>>({})

  const params: Record<string, string> = {}
  if (accountFilter !== 'all') params.accountId = accountFilter
  if (categoryFilter !== 'All categories') params.category = categoryFilter

  const { data: transactions, error } = useSWR(
    !isDemo && token ? ['transactions', token, JSON.stringify(params)] : null,
    ([, t]) => api.transactions(t, params),
    { fallbackData: isDemo ? MOCK_TRANSACTIONS : undefined },
  )

  const { data: accounts } = useSWR(
    !isDemo && token ? ['accounts', token] : null,
    ([, t]) => api.accounts(t),
    { fallbackData: isDemo ? MOCK_ACCOUNTS : undefined },
  )

  let displayTx = transactions ?? (isDemo ? MOCK_TRANSACTIONS : [])

  // Client-side filter for demo mode
  if (isDemo) {
    if (accountFilter !== 'all') displayTx = displayTx.filter(t => t.accountId === accountFilter)
    if (categoryFilter !== 'All categories') displayTx = displayTx.filter(t => t.category === categoryFilter)
  }

  // Apply local category overrides (for demo edits)
  displayTx = displayTx.map(t => ({
    ...t,
    category: localCategories[t.id] ?? t.category,
  }))

  const displayAccounts = accounts ?? (isDemo ? MOCK_ACCOUNTS : [])

  function handleSaveCategory(id: string, newCategory: string) {
    setLocalCategories(prev => ({ ...prev, [id]: newCategory }))
    setEditingId(null)
    // In real mode, you'd PATCH /transactions/:id { category } here
  }

  return (
    <div className="mx-auto max-w-2xl px-8 py-12">

      {/* ── Header ─────────────────────────────────────────── */}
      <header className="mb-8 border-b border-border pb-8">
        <p className="mb-2 text-xs uppercase tracking-widest text-muted-foreground">
          Transactions
        </p>
        <h1 className="font-serif text-5xl font-semibold tracking-tight text-foreground">
          {displayTx.length > 0 ? displayTx.length : '—'}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {displayTx.length > 0
            ? `${displayTx.length} transaction${displayTx.length !== 1 ? 's' : ''} shown`
            : 'No transactions match the current filter'}
        </p>
      </header>

      {/* ── Filters ────────────────────────────────────────── */}
      <div className="mb-6 flex flex-wrap gap-3">
        {/* Account filter */}
        <div className="relative">
          <select
            value={accountFilter}
            onChange={e => setAccountFilter(e.target.value)}
            className="appearance-none rounded border border-input bg-background py-2 pl-3 pr-8 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary transition-all duration-150"
            aria-label="Filter by account"
          >
            <option value="all">All accounts</option>
            {displayAccounts.map(a => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>
          <ChevronDown size={13} className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
        </div>

        {/* Category filter */}
        <div className="relative">
          <select
            value={categoryFilter}
            onChange={e => setCategoryFilter(e.target.value)}
            className="appearance-none rounded border border-input bg-background py-2 pl-3 pr-8 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary transition-all duration-150"
            aria-label="Filter by category"
          >
            {ALL_CATEGORIES.map(c => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          <ChevronDown size={13} className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
        </div>

        {(accountFilter !== 'all' || categoryFilter !== 'All categories') && (
          <button
            onClick={() => { setAccountFilter('all'); setCategoryFilter('All categories') }}
            className="text-xs text-muted-foreground underline underline-offset-4 hover:text-foreground transition-colors duration-150"
          >
            Clear filters
          </button>
        )}
      </div>

      {/* ── API error ──────────────────────────────────────── */}
      {error && !isDemo && (
        <div className="mb-6 rounded border border-border bg-muted px-4 py-3 text-sm text-muted-foreground">
          Could not load transactions. Check that your server is running.
        </div>
      )}

      {/* ── Ledger ─────────────────────────────────────────── */}
      {/* Column headers */}
      <div className="mb-1 flex items-center gap-4 pb-2 border-b border-border">
        <span className="w-14 flex-shrink-0 text-[10px] uppercase tracking-widest text-muted-foreground">Date</span>
        <span className="flex-1 text-[10px] uppercase tracking-widest text-muted-foreground">Merchant</span>
        <span className="hidden text-[10px] uppercase tracking-widest text-muted-foreground sm:block">Category</span>
        <span className="text-[10px] uppercase tracking-widest text-muted-foreground">Amount</span>
      </div>

      <div role="list" aria-label="Transaction ledger">
        {displayTx.map(tx => (
          <div key={tx.id} role="listitem">
            {/* Main row */}
            <button
              className="ledger-divide w-full text-left"
              onClick={() => setEditingId(editingId === tx.id ? null : tx.id)}
              aria-expanded={editingId === tx.id}
              aria-label={`${tx.merchant}, ${tx.category}, ${tx.amount < 0 ? 'debit' : 'credit'}`}
            >
              <div className="flex items-center gap-4 py-3 transition-colors duration-150 hover:bg-muted/50 -mx-2 px-2 rounded">
                <time dateTime={tx.date} className="w-14 flex-shrink-0 text-xs text-muted-foreground">
                  {formatDate(tx.date)}
                </time>
                <span className="flex-1 truncate text-sm text-foreground">
                  {tx.merchant}
                </span>
                <span className="hidden rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase tracking-wider text-muted-foreground sm:block">
                  {tx.category}
                </span>
                <LedgerAmount amount={tx.amount} />
              </div>
            </button>

            {/* Inline category editor */}
            {editingId === tx.id && (
              <EditCategoryRow
                tx={tx}
                onSave={handleSaveCategory}
                onCancel={() => setEditingId(null)}
              />
            )}
          </div>
        ))}

        {displayTx.length === 0 && !error && (
          <p className="py-10 text-center text-sm text-muted-foreground">
            No transactions found. Try changing the filters.
          </p>
        )}
      </div>
    </div>
  )
}
