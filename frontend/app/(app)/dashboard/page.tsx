'use client'

import Link from 'next/link'
import useSWR from 'swr'
import { ArrowRight } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api } from '@/lib/api'
import { MOCK_SUMMARY, MOCK_TRANSACTIONS } from '@/lib/mock-data'
import { HeroAmount, LedgerAmount, formatMoney } from '@/components/amount'

function formatDate(iso: string) {
  return new Date(iso + 'T00:00:00').toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })
}

function spendDiff(current: number, previous: number) {
  const diff = current - previous
  const sign = diff >= 0 ? 'up' : 'down'
  return `${sign} ${formatMoney(Math.abs(diff))} from last month`
}

export default function DashboardPage() {
  const { token, isDemo } = useAuth()

  const { data: summary, error: summaryError } = useSWR(
    !isDemo && token ? ['summary', token] : null,
    ([, t]) => api.summary(t),
    { fallbackData: isDemo ? MOCK_SUMMARY : undefined },
  )

  const { data: transactions, error: txError } = useSWR(
    !isDemo && token ? ['transactions', token, 'limit=5'] : null,
    ([, t]) => api.transactions(t, { limit: '5' }),
    { fallbackData: isDemo ? MOCK_TRANSACTIONS.slice(0, 5) : undefined },
  )

  const displaySummary = summary ?? (isDemo ? MOCK_SUMMARY : null)
  const displayTx = transactions ?? (isDemo ? MOCK_TRANSACTIONS.slice(0, 5) : [])

  const apiError = !isDemo && (summaryError || txError)

  const now = new Date()
  const period = now.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })

  return (
    <div className="mx-auto max-w-2xl px-8 py-12">

      {/* ── Statement-style header ─────────────────────────── */}
      <header className="mb-10 border-b border-border pb-8">
        <p className="mb-2 text-xs uppercase tracking-widest text-muted-foreground">
          Net position · {period}
        </p>

        {displaySummary ? (
          <HeroAmount
            amount={displaySummary.totalBalance}
            className="text-6xl leading-none"
          />
        ) : (
          <div className="h-14 w-48 animate-pulse rounded bg-muted" />
        )}

        {displaySummary && (
          <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
            You&apos;ve spent{' '}
            <span className="money text-foreground">
              {formatMoney(displaySummary.monthlySpend)}
            </span>{' '}
            this month —{' '}
            {spendDiff(displaySummary.monthlySpend, displaySummary.lastMonthSpend)}.
          </p>
        )}
      </header>

      {/* ── API unavailable notice ─────────────────────────── */}
      {apiError && (
        <div className="mb-6 rounded border border-border bg-muted px-4 py-3 text-sm text-muted-foreground">
          Could not reach the API. Check that your server is running at the configured URL.
        </div>
      )}

      {/* ── Recent transactions ────────────────────────────── */}
      <section aria-labelledby="recent-heading">
        <div className="mb-4 flex items-baseline justify-between">
          <h2
            id="recent-heading"
            className="text-xs font-medium uppercase tracking-widest text-muted-foreground"
          >
            Recent activity
          </h2>
          <Link
            href="/transactions"
            className="flex items-center gap-1 text-xs text-muted-foreground transition-colors duration-150 hover:text-foreground"
          >
            View all <ArrowRight size={12} />
          </Link>
        </div>

        {displayTx.length === 0 && !apiError && (
          <p className="py-8 text-center text-sm text-muted-foreground">
            No transactions yet. Connect a bank account to get started.
          </p>
        )}

        <div className="ledger-divide" role="list">
          {displayTx.map(tx => (
            <div
              key={tx.id}
              role="listitem"
              className="flex items-center gap-4 py-3"
            >
              {/* Date */}
              <time
                dateTime={tx.date}
                className="w-14 flex-shrink-0 text-xs text-muted-foreground"
              >
                {formatDate(tx.date)}
              </time>

              {/* Merchant */}
              <span className="flex-1 truncate text-sm text-foreground">
                {tx.merchant}
              </span>

              {/* Category tag */}
              <span className="hidden rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase tracking-wider text-muted-foreground sm:block">
                {tx.category}
              </span>

              {/* Amount */}
              <LedgerAmount amount={tx.amount} className="text-right" />
            </div>
          ))}
        </div>
      </section>

      {/* ── Quick links ────────────────────────────────────── */}
      <section className="mt-12 grid grid-cols-2 gap-3" aria-label="Quick actions">
        <Link
          href="/accounts"
          className="rounded border border-border bg-background px-4 py-4 text-sm transition-all duration-150 hover:border-primary hover:bg-muted"
        >
          <p className="font-medium text-foreground">Accounts</p>
          <p className="mt-0.5 text-xs text-muted-foreground">Manage connected banks</p>
        </Link>
        <Link
          href="/budgets"
          className="rounded border border-border bg-background px-4 py-4 text-sm transition-all duration-150 hover:border-primary hover:bg-muted"
        >
          <p className="font-medium text-foreground">Budgets</p>
          <p className="mt-0.5 text-xs text-muted-foreground">Check your spending limits</p>
        </Link>
      </section>
    </div>
  )
}
