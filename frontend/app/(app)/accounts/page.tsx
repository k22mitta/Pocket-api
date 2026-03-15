'use client'

import { useEffect, useRef, useState } from 'react'
import useSWR from 'swr'
import { usePlaidLink } from 'react-plaid-link'
import { Building2, Upload, Plus, RotateCw, Database, Trash2, Unlink, X } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api, ApiError, type Account } from '@/lib/api'
import { HeroAmount, LedgerAmount } from '@/components/amount'

const TYPE_LABELS: Record<Account['type'], string> = {
  checking:   'Checking',
  savings:    'Savings',
  credit:     'Credit',
  loan:       'Loan',
  investment: 'Investment',
}

function lastSyncedLabel(iso?: string) {
  if (!iso) return 'Never synced'
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

interface UnlinkConfirmRowProps {
  account: Account
  siblingCount: number
  loading: boolean
  onConfirm: () => void
  onCancel: () => void
}

function UnlinkConfirmRow({ account, siblingCount, loading, onConfirm, onCancel }: UnlinkConfirmRowProps) {
  return (
    <div className="bg-muted px-3 py-3">
      <p className="mb-2 text-xs text-muted-foreground">
        Unlink <strong className="text-foreground">{account.institution}</strong>?{' '}
        {siblingCount > 1
          ? `This removes all ${siblingCount} accounts from this connection and their transactions.`
          : 'This removes the account and its transactions.'}{' '}
        This can&apos;t be undone.
      </p>
      <div className="flex gap-2">
        <button
          onClick={onConfirm}
          disabled={loading}
          className="flex items-center gap-1.5 rounded bg-destructive px-3 py-2 text-xs font-medium text-primary-foreground transition-all duration-150 hover:opacity-90 disabled:opacity-50"
        >
          {loading ? <RotateCw size={12} className="animate-spin" /> : <Unlink size={12} />}
          {loading ? 'Unlinking…' : 'Unlink'}
        </button>
        <button
          onClick={onCancel}
          disabled={loading}
          className="flex items-center gap-1.5 rounded border border-border px-3 py-2 text-xs text-muted-foreground transition-all duration-150 hover:text-foreground disabled:opacity-50"
        >
          <X size={12} /> Cancel
        </button>
      </div>
    </div>
  )
}

export default function AccountsPage() {
  const { token, isDemo } = useAuth()
  const fileRef = useRef<HTMLInputElement>(null)
  const [uploadStatus, setUploadStatus] = useState<'idle' | 'uploading' | 'done' | 'partial' | 'error'>('idle')
  const [uploadMessage, setUploadMessage] = useState('')
  const [uploadErrors, setUploadErrors] = useState<string[]>([])
  const [linkToken, setLinkToken] = useState<string | null>(null)
  const [linkingBank, setLinkingBank] = useState(false)
  const [demoOp, setDemoOp] = useState<'idle' | 'load' | 'reset' | 'error'>('idle')
  const [demoError, setDemoError] = useState<string | null>(null)
  const [linkError, setLinkError] = useState<string | null>(null)
  const [unlinkingAccountId, setUnlinkingAccountId] = useState<string | null>(null)
  const [unlinkLoading, setUnlinkLoading] = useState(false)
  const [unlinkError, setUnlinkError] = useState<string | null>(null)

  const { data: accounts, error, mutate } = useSWR(
    token ? ['accounts', token] : null,
    ([, t]) => api.accounts(t),
  )

  // usePlaidLink lives here (not in a child component) so it is never
  // unmounted/remounted when linkToken changes. This prevents React StrictMode
  // from double-invoking window.Plaid.create() — StrictMode only double-invokes
  // on mount, and AccountsPage's mount always has linkToken = null (empty token),
  // which makes usePlaidLink's effect return early before calling Plaid.create().
  // When linkToken is later set by a user click (a state update, not a mount),
  // the effect re-runs exactly once.
  const { open: openPlaid, ready: plaidReady } = usePlaidLink({
    token: linkToken ?? '',
    onSuccess: async (publicToken, metadata) => {
      if (!token) return
      await api.plaidExchange(
        token,
        publicToken,
        metadata.institution?.institution_id ?? '',
        metadata.institution?.name ?? '',
      )
      setLinkToken(null)
      setLinkingBank(false)
      mutate()
    },
    onExit: () => {
      setLinkToken(null)
      setLinkingBank(false)
    },
  })

  useEffect(() => {
    if (plaidReady && linkToken) openPlaid()
  }, [plaidReady, linkToken, openPlaid])

  const displayAccounts = accounts ?? []
  const totalBalance = displayAccounts.reduce((sum, a) => sum + a.balance, 0)
  const demoLoading = demoOp === 'load' || demoOp === 'reset'

  async function handleDemoLoad() {
    if (!token || demoLoading) return
    setDemoOp('load')
    setDemoError(null)
    try {
      await api.demoLoad(token)
      await mutate()
      setDemoOp('idle')
    } catch {
      setDemoOp('error')
      setDemoError('Failed to load demo data. Check that your API is running.')
    }
  }

  async function handleDemoReset() {
    if (!token || demoLoading) return
    setDemoOp('reset')
    setDemoError(null)
    try {
      await api.demoReset(token)
      await mutate()
      setDemoOp('idle')
    } catch {
      setDemoOp('error')
      setDemoError('Failed to reset demo data. Check that your API is running.')
    }
  }

  async function handleConnectBank() {
    if (linkingBank || linkToken !== null || !token) return
    setLinkingBank(true)
    setLinkError(null)
    try {
      const { link_token } = await api.plaidLinkToken(token)
      setLinkToken(link_token)
    } catch {
      setLinkError('Could not start Plaid Link. Check that your API is running.')
      setLinkingBank(false)
    }
  }

  async function handleUnlink(account: Account) {
    if (!token) return
    setUnlinkLoading(true)
    setUnlinkError(null)
    try {
      await api.unlinkPlaidItem(token, account.plaidItemId)
      setUnlinkingAccountId(null)
      await mutate()
    } catch {
      setUnlinkError(`Failed to unlink ${account.institution}. Check that your API is running.`)
    } finally {
      setUnlinkLoading(false)
    }
  }

  async function handleCsvUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file || !token) return
    setUploadStatus('uploading')
    setUploadMessage('')
    setUploadErrors([])
    try {
      const fd = new FormData()
      fd.append('file', file)
      const { imported, errors: rowErrors } = await api.csvUpload(token, fd)
      setUploadStatus(rowErrors.length > 0 ? 'partial' : 'done')
      setUploadMessage(`Imported ${imported} transaction${imported !== 1 ? 's' : ''}.`)
      setUploadErrors(rowErrors)
      mutate()
    } catch (err) {
      setUploadStatus('error')
      setUploadMessage(
        err instanceof ApiError
          ? err.message
          : 'Upload failed. Make sure the file is a valid CSV statement.',
      )
    } finally {
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  return (
    <div className="mx-auto max-w-2xl px-8 py-12">
      <header className="mb-10 border-b border-border pb-8">
        <p className="mb-2 text-xs uppercase tracking-widest text-muted-foreground">
          Accounts
        </p>
        {displayAccounts.length > 0 ? (
          <HeroAmount amount={totalBalance} className="text-5xl leading-none" animate={false} />
        ) : (
          <h1 className="money text-5xl font-semibold tracking-tight text-foreground">—</h1>
        )}
        {displayAccounts.length > 0 && (
          <p className="mt-2 text-sm text-muted-foreground">
            Across {displayAccounts.length} account{displayAccounts.length !== 1 ? 's' : ''}
          </p>
        )}
      </header>

      {error && (
        <div className="mb-6 rounded border border-border bg-muted px-4 py-3 text-sm text-muted-foreground">
          Could not load accounts. Check that your server is running.
        </div>
      )}

      <div className="ledger-divide mb-8" role="list" aria-label="Bank accounts">
        {displayAccounts.map(account => (
          <div key={account.id} role="listitem">
            <div className="flex items-center gap-4 py-4">
              <div
                className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded bg-muted"
                aria-hidden="true"
              >
                <Building2 size={14} className="text-muted-foreground" strokeWidth={1.75} />
              </div>

              <div className="flex-1 min-w-0">
                <p className="truncate text-sm font-medium text-foreground">
                  {account.name}
                </p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {account.institution} · {TYPE_LABELS[account.type]}
                  {account.lastSynced && (
                    <> · Synced {lastSyncedLabel(account.lastSynced)}</>
                  )}
                </p>
              </div>

              <LedgerAmount
                amount={account.balance}
                className="font-medium text-right"
              />

              {account.isPlaidLinked && (
                <button
                  onClick={() => setUnlinkingAccountId(unlinkingAccountId === account.id ? null : account.id)}
                  className="flex-shrink-0 text-muted-foreground transition-colors duration-150 hover:text-destructive"
                  aria-label={`Unlink ${account.institution}`}
                  aria-expanded={unlinkingAccountId === account.id}
                >
                  <Unlink size={14} strokeWidth={1.75} />
                </button>
              )}
            </div>

            {unlinkingAccountId === account.id && (
              <UnlinkConfirmRow
                account={account}
                siblingCount={displayAccounts.filter(a => a.plaidItemId === account.plaidItemId).length}
                loading={unlinkLoading}
                onConfirm={() => handleUnlink(account)}
                onCancel={() => setUnlinkingAccountId(null)}
              />
            )}
          </div>
        ))}

        {displayAccounts.length === 0 && !error && (
          <p className="py-10 text-center text-sm text-muted-foreground">
            {isDemo
              ? 'No demo data loaded. Click "Load demo data" to seed sample accounts.'
              : 'No accounts connected. Connect a bank or upload a statement to get started.'}
          </p>
        )}
      </div>

      {isDemo ? (
        <div className="flex flex-col gap-3 sm:flex-row">
          <button
            onClick={handleDemoLoad}
            disabled={demoLoading}
            className="flex flex-1 items-center justify-center gap-2 rounded bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition-all duration-150 hover:opacity-90 disabled:opacity-50"
          >
            {demoOp === 'load'
              ? <RotateCw size={15} className="animate-spin" aria-hidden="true" />
              : <Database size={15} strokeWidth={1.75} aria-hidden="true" />}
            {demoOp === 'load' ? 'Loading…' : 'Load demo data'}
          </button>

          <button
            onClick={handleDemoReset}
            disabled={demoLoading}
            className="flex flex-1 items-center justify-center gap-2 rounded border border-border px-4 py-2.5 text-sm text-foreground transition-all duration-150 hover:bg-muted disabled:opacity-50"
          >
            {demoOp === 'reset'
              ? <RotateCw size={15} className="animate-spin" aria-hidden="true" />
              : <Trash2 size={15} strokeWidth={1.75} aria-hidden="true" />}
            {demoOp === 'reset' ? 'Resetting…' : 'Reset demo'}
          </button>
        </div>
      ) : (
        <div className="flex flex-col gap-3 sm:flex-row">
          <button
            onClick={handleConnectBank}
            disabled={linkingBank || linkToken !== null}
            className="flex flex-1 items-center justify-center gap-2 rounded bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition-all duration-150 hover:opacity-90 disabled:opacity-50"
          >
            <Plus size={15} strokeWidth={2} aria-hidden="true" />
            {linkingBank ? 'Opening Plaid…' : 'Connect a bank'}
          </button>

          <label className="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded border border-border px-4 py-2.5 text-sm text-foreground transition-all duration-150 hover:bg-muted">
            <Upload size={15} strokeWidth={1.75} aria-hidden="true" />
            Upload a statement (CSV)
            <input
              ref={fileRef}
              type="file"
              accept=".csv"
              className="sr-only"
              onChange={handleCsvUpload}
            />
          </label>
        </div>
      )}

      {!isDemo && (
        <p className="mt-3 text-xs text-muted-foreground">
          CSV needs a date, description/merchant, and either an amount column
          (negative = money out, positive = money in) or separate debit/credit
          columns. Dates: YYYY-MM-DD or MM/DD/YYYY.
        </p>
      )}

      {isDemo && demoOp === 'error' && demoError && (
        <div
          role="alert"
          className="mt-3 rounded border border-red-200 bg-red-50 px-3 py-2.5 text-sm text-destructive"
        >
          {demoError}
        </div>
      )}

      {!isDemo && linkError && (
        <div
          role="alert"
          className="mt-3 rounded border border-red-200 bg-red-50 px-3 py-2.5 text-sm text-destructive"
        >
          {linkError}
        </div>
      )}

      {unlinkError && (
        <div
          role="alert"
          className="mt-3 rounded border border-red-200 bg-red-50 px-3 py-2.5 text-sm text-destructive"
        >
          {unlinkError}
        </div>
      )}

      {uploadStatus !== 'idle' && (
        <div
          role="status"
          className={`mt-3 rounded border px-3 py-2.5 text-sm ${
            uploadStatus === 'uploading'
              ? 'border-border bg-muted text-muted-foreground'
              : uploadStatus === 'done'
              ? 'border-green-200 bg-green-50 text-green-800'
              : uploadStatus === 'partial'
              ? 'border-amber-200 bg-amber-50 text-amber-900'
              : 'border-red-200 bg-red-50 text-destructive'
          }`}
        >
          <div className="flex items-center gap-2">
            {uploadStatus === 'uploading' && (
              <RotateCw size={13} className="animate-spin" aria-hidden="true" />
            )}
            {uploadStatus === 'uploading' ? 'Uploading…' : uploadMessage}
            {uploadStatus === 'partial' &&
              ` ${uploadErrors.length} row${uploadErrors.length !== 1 ? 's' : ''} skipped.`}
          </div>
          {uploadStatus === 'partial' && uploadErrors.length > 0 && (
            <ul className="mt-1.5 list-disc pl-4 text-xs text-amber-800">
              {uploadErrors.map((e, i) => (
                <li key={i}>{e}</li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
