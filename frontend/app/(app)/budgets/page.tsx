'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { ChevronDown, Plus, X, Check, Pencil, Trash2 } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api, CATEGORIES, type Budget } from '@/lib/api'
import { formatMoney } from '@/components/amount'

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

interface EditLimitRowProps {
  budget: Budget
  onSave: (id: string, newLimit: number) => void
  onDelete: (id: string) => void
  onCancel: () => void
}

function EditLimitRow({ budget, onSave, onDelete, onCancel }: EditLimitRowProps) {
  const [value, setValue] = useState(String(budget.amountLimit))
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  function handleSave() {
    const num = parseFloat(value)
    if (!isNaN(num) && num > 0) {
      onSave(budget.id, num)
    }
  }

  if (confirmingDelete) {
    return (
      <div className="bg-muted px-3 py-3">
        <p className="mb-2 text-xs text-muted-foreground">
          Delete the{' '}
          <strong className="text-foreground">{budget.category}</strong> budget?
          This can&apos;t be undone.
        </p>
        <div className="flex gap-2">
          <button
            onClick={() => onDelete(budget.id)}
            className="flex items-center gap-1.5 rounded bg-destructive px-3 py-2 text-xs font-medium text-primary-foreground transition-all duration-150 hover:opacity-90"
          >
            <Trash2 size={12} /> Delete
          </button>
          <button
            onClick={() => setConfirmingDelete(false)}
            className="flex items-center gap-1.5 rounded border border-border px-3 py-2 text-xs text-muted-foreground transition-all duration-150 hover:text-foreground"
          >
            <X size={12} /> Cancel
          </button>
        </div>
      </div>
    )
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
          onClick={() => setConfirmingDelete(true)}
          className="flex items-center gap-1.5 rounded border border-border px-3 py-2 text-xs text-muted-foreground transition-all duration-150 hover:text-destructive"
        >
          <Trash2 size={12} /> Delete
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

interface AddBudgetFormProps {
  onSave: (category: string, limit: number) => void
  onCancel: () => void
  existingCategories: string[]
}

function AddBudgetForm({ onSave, onCancel, existingCategories }: AddBudgetFormProps) {
  const availableCategories = CATEGORIES.filter(c => !existingCategories.includes(c))
  const [category, setCategory] = useState<string>(availableCategories[0] ?? '')
  const [limit, setLimit] = useState('')

  function handleSave() {
    const num = parseFloat(limit)
    if (!category || isNaN(num) || num <= 0) return
    onSave(category, num)
  }

  return (
    <div className="rounded border border-border bg-muted p-4">
      <p className="mb-3 text-xs font-medium uppercase tracking-widest text-muted-foreground">
        New budget
      </p>
      <div className="flex flex-col gap-2 sm:flex-row">
        <div className="relative flex-1">
          <select
            value={category}
            onChange={e => setCategory(e.target.value)}
            className="w-full appearance-none rounded border border-input bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
            aria-label="Category"
            autoFocus
          >
            {availableCategories.map(c => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          <ChevronDown size={13} className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
        </div>
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

export default function BudgetsPage() {
  const { token } = useAuth()

  const [editingId, setEditingId] = useState<string | null>(null)
  const [addingNew, setAddingNew] = useState(false)

  const { data: fetched, error, mutate } = useSWR(
    token ? ['budgets', token] : null,
    ([, t]) => api.budgets(t),
  )

  const budgets: Budget[] = fetched ?? []

  const totalSpent  = budgets.reduce((s, b) => s + b.spent,        0)
  const totalLimits = budgets.reduce((s, b) => s + b.amountLimit, 0)
  const overBudget  = budgets.filter(b => b.spent >= b.amountLimit)

  async function handleSaveLimit(id: string, newLimit: number) {
    if (!token) return
    try {
      await api.updateBudget(token, id, newLimit)
      mutate()
    } catch {
      // keep editing open on error
      return
    }
    setEditingId(null)
  }

  async function handleDeleteBudget(id: string) {
    if (!token) return
    try {
      await api.deleteBudget(token, id)
      mutate()
    } catch {
      return
    }
    setEditingId(null)
  }

  async function handleAddBudget(category: string, limit: number) {
    if (!token) return
    try {
      await api.createBudget(token, category, limit)
      mutate()
    } catch {
      return
    }
    setAddingNew(false)
  }

  return (
    <div className="mx-auto max-w-2xl px-8 py-12">
      <header className="mb-10 border-b border-border pb-8">
        <p className="mb-2 text-xs uppercase tracking-widest text-muted-foreground">
          Budgets
        </p>
        <div className="flex items-end justify-between gap-4">
          <div>
            <h1 className="money text-5xl font-semibold tracking-tight text-foreground">
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

      {error && (
        <div className="mb-6 rounded border border-border bg-muted px-4 py-3 text-sm text-muted-foreground">
          Could not load budgets. Check that your server is running.
        </div>
      )}

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

      <div role="list" aria-label="Budget categories">
        {budgets.map(budget => {
          const remaining = budget.amountLimit - budget.spent
          const isOver    = remaining < 0

          return (
            <div key={budget.id} role="listitem">
              <div className="flex items-center gap-4 py-4 ledger-divide">
                <div className="flex-1 min-w-0">
                  <p className="mb-1.5 text-sm text-foreground">{budget.category}</p>
                  <ProgressBar spent={budget.spent} limit={budget.amountLimit} />
                </div>

                <div className="hidden w-32 flex-shrink-0 text-right sm:block">
                  <p className="money text-sm text-foreground">
                    {formatMoney(budget.spent)}
                  </p>
                  <p className="money text-xs text-muted-foreground">
                    of {formatMoney(budget.amountLimit)}
                  </p>
                </div>

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

              {editingId === budget.id && (
                <EditLimitRow
                  budget={budget}
                  onSave={handleSaveLimit}
                  onDelete={handleDeleteBudget}
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

      <div className="mt-6">
        {addingNew ? (
          <AddBudgetForm
            onSave={handleAddBudget}
            onCancel={() => setAddingNew(false)}
            existingCategories={budgets.map(b => b.category)}
          />
        ) : budgets.length < CATEGORIES.length ? (
          <button
            onClick={() => setAddingNew(true)}
            className="flex items-center gap-2 text-sm text-muted-foreground transition-colors duration-150 hover:text-foreground"
          >
            <Plus size={15} strokeWidth={1.75} aria-hidden="true" />
            Add a budget
          </button>
        ) : (
          <p className="text-xs text-muted-foreground">
            A budget already exists for every category.
          </p>
        )}
      </div>
    </div>
  )
}
