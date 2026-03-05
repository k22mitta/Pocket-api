'use client'

import { useEffect, useRef, useState } from 'react'
import { useChat } from '@ai-sdk/react'
import { DefaultChatTransport } from 'ai'
import { ArrowUp, Sparkles } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { MOCK_SUMMARY, MOCK_ACCOUNTS, MOCK_TRANSACTIONS, MOCK_BUDGETS } from '@/lib/mock-data'

// ─── Build financial context string from mock data ───────────────────────────
function buildFinancialContext() {
  const accountLines = MOCK_ACCOUNTS.map(
    a => `  - ${a.name} (${a.type}): $${a.balance.toLocaleString('en-US', { minimumFractionDigits: 2 })}`,
  ).join('\n')

  const txLines = MOCK_TRANSACTIONS.slice(0, 12)
    .map(t => `  - ${t.date} | ${t.merchant} | ${t.category} | ${t.amount >= 0 ? '+' : ''}$${t.amount.toLocaleString('en-US', { minimumFractionDigits: 2 })}`)
    .join('\n')

  const budgetLines = MOCK_BUDGETS.map(
    b => `  - ${b.category}: $${b.spent} spent / $${b.limit} limit (${Math.round((b.spent / b.limit) * 100)}%)`,
  ).join('\n')

  return `SUMMARY
  Total balance: $${MOCK_SUMMARY.totalBalance.toLocaleString('en-US', { minimumFractionDigits: 2 })}
  This month's spend: $${MOCK_SUMMARY.monthlySpend.toLocaleString('en-US', { minimumFractionDigits: 2 })}
  Last month's spend: $${MOCK_SUMMARY.lastMonthSpend.toLocaleString('en-US', { minimumFractionDigits: 2 })}

ACCOUNTS
${accountLines}

RECENT TRANSACTIONS (latest 12)
${txLines}

BUDGETS
${budgetLines}`
}

// ─── Suggested prompts shown on empty state ──────────────────────────────────
const SUGGESTIONS = [
  'How much did I spend on dining this month?',
  'Which budget am I closest to exceeding?',
  'What were my biggest expenses recently?',
  "How does this month's spending compare to last month?",
]

// ─── Message bubble ──────────────────────────────────────────────────────────
interface BubbleProps {
  role: 'user' | 'assistant'
  parts: Array<{ type: string; text?: string }>
}

function Bubble({ role, parts }: BubbleProps) {
  const text = parts
    .filter((p): p is { type: 'text'; text: string } => p.type === 'text')
    .map(p => p.text)
    .join('')

  if (!text) return null

  if (role === 'user') {
    return (
      <div className="flex justify-end">
        <div className="max-w-sm rounded rounded-br-sm bg-primary px-4 py-2.5 text-sm leading-relaxed text-primary-foreground">
          {text}
        </div>
      </div>
    )
  }

  return (
    <div className="flex items-start gap-3">
      <div
        className="mt-0.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-foreground"
        aria-hidden="true"
      >
        <Sparkles size={13} className="text-background" strokeWidth={1.75} />
      </div>
      <div className="max-w-prose text-sm leading-relaxed text-foreground whitespace-pre-wrap">
        {text}
      </div>
    </div>
  )
}

// ─── Typing indicator ────────────────────────────────────────────────────────
function TypingIndicator() {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-0.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-foreground" aria-hidden="true">
        <Sparkles size={13} className="text-background" strokeWidth={1.75} />
      </div>
      <div className="flex items-center gap-1 py-2" aria-label="Pocket is thinking">
        {[0, 1, 2].map(i => (
          <span
            key={i}
            className="h-1.5 w-1.5 rounded-full bg-muted-foreground"
            style={{ animation: `bounce 1.2s ease-in-out ${i * 0.2}s infinite` }}
          />
        ))}
      </div>
    </div>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────
export default function AskPage() {
  const { isDemo } = useAuth()
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  const financialContext = buildFinancialContext()

  const { messages, sendMessage, status } = useChat({
    transport: new DefaultChatTransport({
      api: '/api/ask',
      prepareSendMessagesRequest: ({ id, messages: msgs }) => ({
        body: {
          id,
          messages: msgs,
          financialContext: isDemo ? financialContext : undefined,
        },
      }),
    }),
  })

  const isStreaming = status === 'streaming' || status === 'submitted'

  // Auto-scroll to bottom when messages update
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, status])

  function handleSend() {
    const text = input.trim()
    if (!text || isStreaming) return
    sendMessage({ text })
    setInput('')
  }

  return (
    <div className="flex h-full flex-col">

      {/* ── Header ─────────────────────────────────────────── */}
      <header className="flex-shrink-0 border-b border-border px-8 py-6">
        <p className="mb-0.5 text-xs uppercase tracking-widest text-muted-foreground">
          Ask Pocket
        </p>
        <h1 className="font-serif text-2xl font-semibold tracking-tight text-foreground">
          Your financial advisor
        </h1>
      </header>

      {/* ── Messages ───────────────────────────────────────── */}
      <div className="flex-1 overflow-y-auto px-8 py-6">
        <div className="mx-auto max-w-2xl space-y-5">

          {/* Empty state */}
          {messages.length === 0 && (
            <div>
              <p className="mb-6 text-sm text-muted-foreground">
                Ask anything about your finances. Pocket has access to your accounts,
                recent transactions, and budget progress.
              </p>
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                {SUGGESTIONS.map(s => (
                  <button
                    key={s}
                    onClick={() => {
                      setInput(s)
                    }}
                    className="rounded border border-border bg-background px-4 py-3 text-left text-sm text-muted-foreground transition-all duration-150 hover:border-primary hover:bg-muted hover:text-foreground"
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Conversation */}
          {messages.map(msg => (
            <Bubble key={msg.id} role={msg.role as 'user' | 'assistant'} parts={msg.parts as Array<{ type: string; text?: string }>} />
          ))}

          {/* Typing indicator — show when submitted but not yet streaming */}
          {status === 'submitted' && <TypingIndicator />}

          <div ref={bottomRef} />
        </div>
      </div>

      {/* ── Input ──────────────────────────────────────────── */}
      <div className="flex-shrink-0 border-t border-border px-8 py-5">
        <div className="mx-auto max-w-2xl">
          <form
            onSubmit={e => {
              e.preventDefault()
              handleSend()
            }}
            className="flex items-end gap-3"
          >
            <div className="relative flex-1">
              <textarea
                value={input}
                onChange={e => {
                  setInput(e.target.value)
                  // Auto-resize up to 5 lines
                  e.target.style.height = 'auto'
                  e.target.style.height = `${Math.min(e.target.scrollHeight, 120)}px`
                }}
                onKeyDown={e => {
                  if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing && !(e.keyCode === 229)) {
                    e.preventDefault()
                    handleSend()
                  }
                }}
                placeholder="Ask about your spending, budgets, or accounts…"
                disabled={isStreaming}
                rows={1}
                className="w-full resize-none rounded border border-input bg-background px-4 py-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary disabled:opacity-60"
                style={{ lineHeight: '1.5', minHeight: '44px', maxHeight: '120px' }}
                aria-label="Message input"
              />
            </div>
            <button
              type="submit"
              disabled={!input.trim() || isStreaming}
              className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded bg-primary text-primary-foreground transition-all duration-150 hover:opacity-90 disabled:opacity-40"
              aria-label="Send message"
            >
              <ArrowUp size={16} strokeWidth={2} />
            </button>
          </form>
          <p className="mt-2 text-center text-[10px] text-muted-foreground">
            Pocket may make mistakes. Always verify important financial decisions.
          </p>
        </div>
      </div>

      {/* Bounce keyframe */}
      <style>{`
        @keyframes bounce {
          0%, 80%, 100% { transform: translateY(0); opacity: 0.4; }
          40% { transform: translateY(-5px); opacity: 1; }
        }
      `}</style>
    </div>
  )
}
