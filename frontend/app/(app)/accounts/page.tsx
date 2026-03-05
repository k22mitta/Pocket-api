'use client'

import { useRef, useState } from 'react'
import useSWR from 'swr'
import { Building2, Upload, Plus, RotateCw } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api, type Account } from '@/lib/api'
import { MOCK_ACCOUNTS } from '@/lib/mock-data'
import { LedgerAmount } from '@/components/amount'

const TYPE_LABELS: Record<Account['type'], string> = {
  checking:   'Checking',
  savings:    'Savings',
  credit:     'Credit',
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

export default function AccountsPage() {
  const { token, isDemo } = useAuth()
  const fileRef = useRef<HTMLInputElement>(null)
  const [uploadStatus, setUploadStatus] = useState<'idle' | 'uploading' | 'done' | 'error'>('idle')
  const [uploadMessage, setUploadMessage] = useState('')
  const [linkingBank, setLinkingBank] = useState(false)

  const { data: accounts, error, mutate } = useSWR(
    !isDemo && token ? ['accounts', token] : null,
    ([, t]) => api.accounts(t),
    { fallbackData: isDemo ? MOCK_ACCOUNTS : undefined },
  )

  const displayAccounts = accounts ?? (isDemo ? MOCK_ACCOUNTS : [])
  const totalBalance = displayAccounts.reduce((sum, a) => sum + a.balance, 0)

  async function handleConnectBank() {
    if (isDemo || !token) {
      alert('Connect your real API to use Plaid bank linking.')
      return
    }
    setLinkingBank(true)
    try {
      const { link_token } = await api.plaidLinkToken(token)
      // In production: open Plaid Link with link_token here.
      // Plaid Link is a client-side flow; integrate @plaid/link-react for the full experience.
      console.log('[v0] Plaid link token received:', link_token)
      alert(`Plaid link token received (${link_token.slice(0, 20)}…). Integrate @plaid/link-react to complete the flow.`)
    } catch {
      alert('Could not start Plaid Link. Check that your API is running.')
    } finally {
      setLinkingBank(false)
    }
  }

  async function handleCsvUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file || !token) return
    if (isDemo) {
      alert('Connect your real API to upload CSV statements.')
      return
    }
    setUploadStatus('uploading')
    setUploadMessage('')
    try {
      const fd = new FormData()
      fd.append('file', file)
      const { imported } = await api.csvUpload(token, fd)
      setUploadStatus('done')
      setUploadMessage(`Imported ${imported} transaction${imported !== 1 ? 's' : ''}.`)
      mutate()
    } catch {
      setUploadStatus('error')
      setUploadMessage('Upload failed. Make sure the file is a valid CSV statement.')
    } finally {
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  return (
    <div className="mx-auto max-w-2xl px-8 py-12">

      {/* ── Header ─────────────────────────────────────────── */}
      <header className="mb-10 border-b border-border pb-8">
        <p className="mb-2 text-xs uppercase tracking-widest text-muted-foreground">
          Accounts
        </p>
        <h1 className="font-serif text-5xl font-semibold tracking-tight text-foreground">
          {displayAccounts.length > 0
            ? (totalBalance < 0 ? '−' : '') +
              new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' })
                .format(Math.abs(totalBalance))
            : '—'}
        </h1>
        {displayAccounts.length > 0 && (
          <p className="mt-2 text-sm text-muted-foreground">
            Across {displayAccounts.length} account{displayAccounts.length !== 1 ? 's' : ''}
          </p>
        )}
      </header>

      {/* ── Account list ───────────────────────────────────── */}
      {error && !isDemo && (
        <div className="mb-6 rounded border border-border bg-muted px-4 py-3 text-sm text-muted-foreground">
          Could not load accounts. Check that your server is running.
        </div>
      )}

      <div className="ledger-divide mb-8" role="list" aria-label="Bank accounts">
        {displayAccounts.map(account => (
          <div
            key={account.id}
            role="listitem"
            className="flex items-center gap-4 py-4"
          >
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
          </div>
        ))}

        {displayAccounts.length === 0 && !error && (
          <p className="py-10 text-center text-sm text-muted-foreground">
            No accounts connected. Connect a bank or upload a statement to get started.
          </p>
        )}
      </div>

      {/* ── Actions ────────────────────────────────────────── */}
      <div className="flex flex-col gap-3 sm:flex-row">
        <button
          onClick={handleConnectBank}
          disabled={linkingBank}
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

      {/* Upload feedback */}
      {uploadStatus !== 'idle' && (
        <div
          role="status"
          className={`mt-3 flex items-center gap-2 rounded border px-3 py-2.5 text-sm ${
            uploadStatus === 'uploading'
              ? 'border-border bg-muted text-muted-foreground'
              : uploadStatus === 'done'
              ? 'border-green-200 bg-green-50 text-green-800'
              : 'border-red-200 bg-red-50 text-destructive'
          }`}
        >
          {uploadStatus === 'uploading' && (
            <RotateCw size={13} className="animate-spin" aria-hidden="true" />
          )}
          {uploadStatus === 'uploading' ? 'Uploading…' : uploadMessage}
        </div>
      )}
    </div>
  )
}
