'use client'

import { useState } from 'react'
import useSWR, { useSWRConfig } from 'swr'
import { Plus, X, Check, Pencil } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api, type Budget } from '@/lib/api'
import { MOCK_BUDGETS } from '@/lib/mock-data'
import { formatMoney } from '@/components/amount'

// ─── Budget progress bar ─────────────────────────────────────────────────────
interface ProgressBarProps {
  spent: number
  limit: number
}

function ProgressBar({ spent, limit }: ProgressBarProps) {
  const pct = Math.min((spent / limit) * 100, 100)
  const isOver = spent >= limit
  const isWarning = pct >= 80 && !isOver

  let barColor = 'bg-primary'
  if (isOver) barColor = 'bg-destructive'
  else if (isWarning) barColor = 'bg-accent'

  return (
    <div
      className="h-1 w-full overflow-hidden rounded-full bg-muted"
      role="progressbar"
      aria-valuenow={Math.round(pct)}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={`${Math.round(pct)}% of budget used`}
    >
      <div
        className={`h-full rounded-full transition-all duration-700 ${barColor}`}
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}

// ─── Inline edit row ─────────────────────────────────────────────────────────
interface EditLimitRowProps {
  budget: Budget
  onSave: (id: string, newLimit: number) => void
  onCancel: () => void
}

function EditLimitRow({ budget, onSave, onCancel }: EditLimitRowProps) {
  const [value, setValue] = useState(String(budget.limit))

  function handleSave() {
    const num = parseFloat(value)
    if (!isNaN(num) && num > 0) {
      onSave(budget.id, num)
    }
  }

  return (
    <div className="bg-muted px-3 py-3">
      <p className="mb-2 text-xs text-muted-foreground">
        New monthly limit for{' '}
        <strong className="text-foreground">{budget.category}</strong>
      </p>
      <div className="flex gap-2">
        <div className="relative flex-1">
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
            $
          </span>
          <input
            type="number"
            min="1"
            step="10"
            value={value}
            onChange={e => setValue(e.target.value)}
            onKeyDown={e => {
              if (e.key === 'Enter' && !e.nativeEvent.isComposing) handleSave()
              if (e.key === 'Escape') onCancel()
            }}
            className="w-full rounded border border-input bg-background py-2 pl-7 pr-3 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
            autoFocus
          />
        </div>
        <button
          onClick={handleSave}
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

// ─── Add budget form ──────────────────────────────────────────────────────────
interface AddBudgetFormProps {
  onSave: (category: string, limit: number) => void
  onCancel: () => void
  existingCategories: string[]
}

function AddBudgetForm({ onSave, onCancel, existingCategories }: AddBudgetFormProps) {
  const [category, setCategory] = useState('')
  const [limit, setLimit] = useState('')

  function handleSave() {
    const num = parseFloat(limit)
    const cat = category.trim()
    if (!cat || isNaN(num) || num <= 0) return
    onSave(cat, num)
  }

  return (
    <div className="rounded border border-border bg-muted p-4">
      <p className="mb-3 text-xs font-medium uppercase tracking-widest text-muted-foreground">
        New budget
      </p>
      <div className="flex flex-col gap-2 sm:flex-row">
        <input
          type="text"
          placeholder="Category"
          value={category}
          onChange={e => setCategory(e.target.value)}
          className="flex-1 rounded border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
          autoFocus
        />
        <div className="relative">
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
            $
          </span>
          <input
            type="number"
            min="1"
            step="10"
            placeholder="Monthly limit"
            value={limit}
            onChange={e => setLimit(e.target.value)}
            onKeyDown={e => {
              if (e.key === 'Enter' && !e.nativeEvent.isComposing) handleSave()
              if (e.key === 'Escape') onCancel()
            }}
            className="w-full rounded border border-input bg-background py-2 pl-7 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary sm:w-36"
          />
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleSave}
            className="flex flex-1 items-center justify-center gap-1.5 rounded bg-primary px-4 py-2 text-xs font-medium text-primary-foreground transition-all duration-150 hover:opacity-90 sm:flex-initial"
          >
            <Check size={12} /> Add
          </button>
          <button
            onClick={onCancel}
            className="flex flex-1 items-center justify-center gap-1.5 rounded border border-border px-4 py-2 text-xs text-muted-foreground transition-all duration-150 hover:text-foreground sm:flex-initial"
          >
            <X size={12} /> Cancel
          </button>
        </div>
      </div>
    </div>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────
export default function BudgetsPage() {
  const { token, isDemo } = useAuth()
  const { mutate } = useSWRConfig()

  const [editingId, setEditingId] = useState<string | null>(null)
  const [addingNew, setAddingNew] = useState(false)
  const [localBudgets, setLocalBudgets] = useState<Budget[]>([])

  const { data: fetched, error } = useSWR(
    !isDemo && token ? ['budgets', token] : null,
    ([, t]) => api.budgets(t),
    { fallbackData: isDemo ? MOCK_BUDGETS : undefined },
  )

  // Merge server/demo budgets with any locally added ones
  const baseBudgets = fetched ?? (isDemo ? MOCK_BUDGETS : [])
  const budgets: Budget[] = [...baseBudgets, ...localBudgets]

  const totalSpent  = budgets.reduce((s, b) => s + b.spent,  0)
  const totalLimits = budgets.reduce((s, b) => s + b.limit, 0)
  const overBudget  = budgets.filter(b => b.spent >= b.limit)

  function handleSaveLimit(id: string, newLimit: number) {
    if (isDemo) {
      // mutate the local-override list
      setLocalBudgets(prev =>
        prev.map(b => (b.id === id ? { ...b, limit: newLimit } : b)),
      )
      // also mutate base if it's in base
    } else {
      // PATCH /budgets/:id { limit: newLimit }
      mutate(['budgets', token])
    }
    setEditingId(null)
  }

  function handleAddBudget(category: string, limit: number) {
    const newBudget: Budget = {
      id: `bud_${Date.now()}`,
      category,
      limit,
      spent: 0,
    }
    if (isDemo) {
      setLocalBudgets(prev => [...prev, newBudget])
    } else {
      // POST /budgets { category, limit }
      mutate(['budgets', token])
    }
    setAddingNew(false)
  }

  return (
    <div className="mx-auto max-w-2xl px-8 py-12">

      {/* ── Header ─────────────────────────────────────────── */}
      <header className="mb-10 border-b border-border pb-8">
        <p className="mb-2 text-xs uppercase tracking-widest text-muted-foreground">
          Budgets
        </p>
        <div className="flex items-end justify-between gap-4">
          <div>
            <h1 className="font-serif text-5xl font-semibold tracking-tight text-foreground">
              {totalLimits > 0
                ? formatMoney(totalSpent)
                : '—'}
            </h1>
            {totalLimits > 0 && (
              <p className="mt-2 text-sm text-muted-foreground">
                spent of{' '}
                <span className="money text-foreground">{formatMoney(totalLimits)}</span>{' '}
                budgeted
                {overBudget.length > 0 && (
                  <span className="ml-2 text-destructive">
                    · {overBudget.length} over limit
                  </span>
                )}
              </p>
            )}
          </div>

          {/* Overall bar */}
          {totalLimits > 0 && (
            <div className="hidden w-36 sm:block">
              <ProgressBar spent={totalSpent} limit={totalLimits} />
              <p className="mt-1 text-right text-[10px] text-muted-foreground">
                {Math.round((totalSpent / totalLimits) * 100)}% used
              </p>
            </div>
          )}
        </div>
      </header>

      {/* ── API error ──────────────────────────────────────── */}
      {error && !isDemo && (
        <div className="mb-6 rounded border border-border bg-muted px-4 py-3 text-sm text-muted-foreground">
          Could not load budgets. Check that your server is running.
        </div>
      )}

      {/* ── Column headers ─────────────────────────────────── */}
      {budgets.length > 0 && (
        <div className="mb-1 flex items-center gap-4 border-b border-border pb-2">
          <span className="flex-1 text-[10px] uppercase tracking-widest text-muted-foreground">
            Category
          </span>
          <span className="hidden w-32 text-right text-[10px] uppercase tracking-widest text-muted-foreground sm:block">
            Spent / Limit
          </span>
          <span className="w-20 text-right text-[10px] uppercase tracking-widest text-muted-foreground">
            Remaining
          </span>
          <span className="w-6" aria-hidden="true" />
        </div>
      )}

      {/* ── Budget rows ────────────────────────────────────── */}
      <div role="list" aria-label="Budget categories">
        {budgets.map(budget => {
          const remaining = budget.limit - budget.spent
          const isOver    = remaining < 0

          return (
            <div key={budget.id} role="listitem">
              <div className="flex items-center gap-4 py-4 ledger-divide">
                {/* Category + progress bar */}
                <div className="flex-1 min-w-0">
                  <p className="mb-1.5 text-sm text-foreground">{budget.category}</p>
                  <ProgressBar spent={budget.spent} limit={budget.limit} />
                </div>

                {/* Spent / Limit */}
                <div className="hidden w-32 flex-shrink-0 text-right sm:block">
                  <p className="money text-sm text-foreground">
                    {formatMoney(budget.spent)}
                  </p>
                  <p className="money text-xs text-muted-foreground">
                    of {formatMoney(budget.limit)}
                  </p>
                </div>

                {/* Remaining */}
                <div className="w-20 flex-shrink-0 text-right">
                  <p
                    className={`money text-sm font-medium ${
                      isOver ? 'text-destructive' : 'text-foreground'
                    }`}
                  >
                    {isOver
                      ? `−${formatMoney(Math.abs(remaining))}`
                      : formatMoney(remaining)}
                  </p>
                  <p className="text-[10px] text-muted-foreground">
                    {isOver ? 'over' : 'left'}
                  </p>
                </div>

                {/* Edit button */}
                <button
                  onClick={() => setEditingId(editingId === budget.id ? null : budget.id)}
                  className="w-6 flex-shrink-0 text-muted-foreground transition-colors duration-150 hover:text-foreground"
                  aria-label={`Edit ${budget.category} budget`}
                  aria-expanded={editingId === budget.id}
                >
                  {editingId === budget.id ? (
                    <X size={14} strokeWidth={1.75} />
                  ) : (
                    <Pencil size={14} strokeWidth={1.75} />
                  )}
                </button>
              </div>

              {/* Inline limit editor */}
              {editingId === budget.id && (
                <EditLimitRow
                  budget={budget}
                  onSave={handleSaveLimit}
                  onCancel={() => setEditingId(null)}
                />
              )}
            </div>
          )
        })}

        {budgets.length === 0 && !error && (
          <p className="py-10 text-center text-sm text-muted-foreground">
            No budgets yet. Add one to start tracking your spending.
          </p>
        )}
      </div>

      {/* ── Add new budget ─────────────────────────────────── */}
      <div className="mt-6">
        {addingNew ? (
          <AddBudgetForm
            onSave={handleAddBudget}
            onCancel={() => setAddingNew(false)}
            existingCategories={budgets.map(b => b.category)}
          />
        ) : (
          <button
            onClick={() => setAddingNew(true)}
            className="flex items-center gap-2 text-sm text-muted-foreground transition-colors duration-150 hover:text-foreground"
          >
            <Plus size={15} strokeWidth={1.75} aria-hidden="true" />
            Add a budget
          </button>
        )}
      </div>
    </div>
  )
}
