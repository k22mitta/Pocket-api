// Change NEXT_PUBLIC_API_URL in your .env to point at a deployed backend.
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function monthRange(monthsAgo: number) {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth() - monthsAgo, 1)
  const end = new Date(now.getFullYear(), now.getMonth() - monthsAgo + 1, 0)
  const fmt = (d: Date) => d.toISOString().slice(0, 10)
  return { start: fmt(start), end: fmt(end) }
}

async function request<T>(
  path: string,
  options: RequestInit & { token?: string } = {},
): Promise<T> {
  const { token, headers: extraHeaders, ...init } = options
  const isFormData = init.body instanceof FormData
  const headers: Record<string, string> = {
    ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
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

export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{ token: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),

    signup: async (email: string, password: string, name: string) => {
      await request<{ id: string; email: string; name: string }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, name }),
      })
      return request<{ token: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
    },
  },

  summary: async (token: string): Promise<Summary> => {
    const thisMonth = monthRange(0)
    const lastMonth = monthRange(1)
    const sumValues = (obj: Record<string, number>) => Object.values(obj).reduce((a, b) => a + b, 0)

    const [balanceRes, thisSpendRes, lastSpendRes, cashflowRes] = await Promise.all([
      request<{ balance: number }>('/summary/balance', { token }),
      request<{ spending: Record<string, number> }>(
        `/summary/spending?start=${thisMonth.start}&end=${thisMonth.end}`,
        { token },
      ),
      request<{ spending: Record<string, number> }>(
        `/summary/spending?start=${lastMonth.start}&end=${lastMonth.end}`,
        { token },
      ),
      request<{ cashflow: Array<{ month: string; income: number; expenses: number }> }>('/summary/cashflow', { token }),
    ])

    return {
      balance: balanceRes.balance,
      spending: thisSpendRes.spending,
      cashflow: cashflowRes.cashflow,
      thisMonthSpend: sumValues(thisSpendRes.spending),
      lastMonthSpend: sumValues(lastSpendRes.spending),
    }
  },

  accounts: (token: string) =>
    request<Account[]>('/accounts', { token }),

  unlinkPlaidItem: (token: string, plaidItemId: string) =>
    request<{ success: boolean }>(`/plaid/items/${plaidItemId}`, {
      method: 'DELETE',
      token,
    }),

  transactions: (token: string, params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : ''
    return request<TransactionsPage>(`/transactions${qs}`, { token })
  },

  budgets: (token: string) =>
    request<Budget[]>('/budgets', { token }),

  createBudget: (token: string, category: string, amountLimit: number) =>
    request<Budget>('/budgets', {
      method: 'POST',
      token,
      body: JSON.stringify({ category, amountLimit }),
    }),

  updateBudget: (token: string, id: string, amountLimit: number) =>
    request<Budget>(`/budgets/${id}`, {
      method: 'PUT',
      token,
      body: JSON.stringify({ amountLimit }),
    }),

  deleteBudget: (token: string, id: string) =>
    request<{ success: boolean }>(`/budgets/${id}`, {
      method: 'DELETE',
      token,
    }),

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

  plaidExchange: (
    token: string,
    public_token: string,
    institution_id: string,
    institution_name: string,
  ) =>
    request<{ success: boolean }>('/plaid/exchange', {
      method: 'POST',
      token,
      body: JSON.stringify({ public_token, institution_id, institution_name }),
    }),

  demoLoad: (token: string) =>
    request<{ message?: string }>('/demo/load', {
      method: 'POST',
      token,
    }),

  demoReset: (token: string) =>
    request<{ message?: string }>('/demo/reset', {
      method: 'POST',
      token,
    }),

  csvUpload: async (token: string, formData: FormData) => {
    const res = await request<{ transactions_created: number; errors: string[] | null }>('/import/csv', {
      method: 'POST',
      token,
      body: formData,
    })
    return { imported: res.transactions_created, errors: res.errors ?? [] }
  },

  chatHistory: (token: string) =>
    request<Array<{ id: string; role: 'user' | 'assistant'; content: string; created_at: string }>>(
      '/chat/history',
      { token },
    ),
}

export const CATEGORIES = [
  'Groceries',
  'Dining',
  'Shopping',
  'Transport',
  'Travel',
  'Subscriptions',
  'Entertainment',
  'Health',
  'Housing',
  'Income',
  'Transfers',
  'Other',
] as const

export interface Summary {
  balance: number
  spending: Record<string, number>
  cashflow: Array<{ month: string; income: number; expenses: number }>
  thisMonthSpend: number
  lastMonthSpend: number
}

export interface Account {
  id: string
  name: string
  type: 'checking' | 'savings' | 'credit' | 'loan' | 'investment'
  institution: string
  balance: number
  lastSynced?: string
  plaidItemId: string
  isPlaidLinked: boolean
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

export interface TransactionsPage {
  transactions: Transaction[]
  total: number
  limit: number
  offset: number
}

export interface Budget {
  id: string
  category: string
  amountLimit: number
  period: string
  spent: number
  remaining: number
}

export interface ChatResponse {
  response: string
  relatedTransactionIds?: string[]
}
