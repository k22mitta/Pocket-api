// ─── Base URL ───────────────────────────────────────────────────────────────
// Change NEXT_PUBLIC_API_URL in your .env to point at a deployed backend.
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

// ─── Error class ────────────────────────────────────────────────────────────
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// ─── Core fetch helper ──────────────────────────────────────────────────────
async function request<T>(
  path: string,
  options: RequestInit & { token?: string } = {},
): Promise<T> {
  const { token, headers: extraHeaders, ...init } = options
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(extraHeaders as Record<string, string>),
  }

  const res = await fetch(`${API_BASE_URL}${path}`, { ...init, headers })

  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new ApiError(res.status, text || `HTTP ${res.status}`)
  }

  return res.json() as Promise<T>
}

// ─── API surface ────────────────────────────────────────────────────────────
export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{ token: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),

    signup: (email: string, password: string) =>
      request<{ token: string }>('/auth/signup', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
  },

  summary: (token: string) =>
    request<Summary>('/summary', { token }),

  accounts: (token: string) =>
    request<Account[]>('/accounts', { token }),

  transactions: (token: string, params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : ''
    return request<Transaction[]>(`/transactions${qs}`, { token })
  },

  budgets: (token: string) =>
    request<Budget[]>('/budgets', { token }),

  chat: (token: string, message: string) =>
    request<ChatResponse>('/chat', {
      method: 'POST',
      token,
      body: JSON.stringify({ message }),
    }),

  plaidLinkToken: (token: string) =>
    request<{ link_token: string }>('/plaid/link-token', {
      method: 'POST',
      token,
    }),

  csvUpload: (token: string, formData: FormData) =>
    request<{ imported: number }>('/csv/upload', {
      method: 'POST',
      token,
      body: formData,
      // Let browser set multipart Content-Type with boundary
      headers: { Authorization: `Bearer ${token}` },
    }),
}

// ─── Types ──────────────────────────────────────────────────────────────────
export interface Summary {
  totalBalance: number
  monthlySpend: number
  lastMonthSpend: number
  transactionCount: number
}

export interface Account {
  id: string
  name: string
  type: 'checking' | 'savings' | 'credit' | 'investment'
  institution: string
  balance: number
  lastSynced?: string
}

export interface Transaction {
  id: string
  accountId: string
  date: string       // ISO date string "YYYY-MM-DD"
  merchant: string
  category: string
  amount: number     // negative = debit, positive = credit/income
  notes?: string
}

export interface Budget {
  id: string
  category: string
  limit: number
  spent: number
}

export interface ChatResponse {
  response: string
  relatedTransactionIds?: string[]
}
