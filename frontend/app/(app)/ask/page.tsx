'use client'

import { useEffect, useRef, useState } from 'react'
import { AlertCircle, ArrowUp, RotateCw, Sparkles } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { api, ApiError } from '@/lib/api'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
}

const SUGGESTIONS = [
  'How much did I spend on dining this month?',
  'Which budget am I closest to exceeding?',
  'What were my biggest expenses recently?',
  "How does this month's spending compare to last month?",
]

function Bubble({ role, content }: Pick<Message, 'role' | 'content'>) {
  if (role === 'user') {
    return (
      <div className="flex justify-end">
        <div className="max-w-sm rounded rounded-br-sm bg-primary px-4 py-2.5 text-sm leading-relaxed text-primary-foreground">
          {content}
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
        {content}
      </div>
    </div>
  )
}

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

export default function AskPage() {
  const { token } = useAuth()
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [status, setStatus] = useState<'idle' | 'loading' | 'error'>('idle')
  const [errorText, setErrorText] = useState<string | null>(null)
  const [lastFailedMessage, setLastFailedMessage] = useState<string | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)

  const isLoading = status === 'loading'

  useEffect(() => {
    if (!token) return
    api.chatHistory(token)
      .then(history => {
        setMessages(history.map(m => ({ id: m.id, role: m.role, content: m.content })))
      })
      .catch(() => {
        // history is a nice-to-have; a fresh conversation is a fine fallback
      })
  }, [token])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, status])

  async function sendToBackend(text: string) {
    if (!token) return
    setStatus('loading')
    setErrorText(null)
    try {
      const res = await api.chat(token, text)
      setMessages(prev => [...prev, { id: `a-${Date.now()}`, role: 'assistant', content: res.response }])
      setStatus('idle')
      setLastFailedMessage(null)
    } catch (err) {
      setStatus('error')
      setLastFailedMessage(text)
      if (err instanceof ApiError && err.status === 429) {
        setErrorText("You're sending messages a bit too fast. Wait a minute and try again.")
      } else {
        setErrorText('Something went wrong reaching Pocket. Check that your API is running.')
      }
    }
  }

  function handleSend() {
    const text = input.trim()
    if (!text || isLoading) return
    setInput('')
    setMessages(prev => [...prev, { id: `u-${Date.now()}`, role: 'user', content: text }])
    sendToBackend(text)
  }

  function handleSuggestion(text: string) {
    if (isLoading) return
    setMessages(prev => [...prev, { id: `u-${Date.now()}`, role: 'user', content: text }])
    sendToBackend(text)
  }

  function handleRetry() {
    if (lastFailedMessage && !isLoading) sendToBackend(lastFailedMessage)
  }

  return (
    <div className="flex h-full flex-col">
      <header className="flex-shrink-0 border-b border-border px-8 py-6">
        <p className="mb-0.5 text-xs uppercase tracking-widest text-muted-foreground">
          Ask Pocket
        </p>
        <h1 className="font-serif text-2xl font-semibold tracking-tight text-foreground">
          Your financial advisor
        </h1>
      </header>

      <div className="flex-1 overflow-y-auto px-8 py-6">
        <div className="mx-auto max-w-2xl space-y-5">
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
                    disabled={isLoading}
                    onClick={() => handleSuggestion(s)}
                    className="rounded border border-border bg-background px-4 py-3 text-left text-sm text-muted-foreground transition-all duration-150 hover:border-primary hover:bg-muted hover:text-foreground disabled:opacity-50"
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          )}

          {messages.map(msg => (
            <Bubble key={msg.id} role={msg.role} content={msg.content} />
          ))}

          {isLoading && <TypingIndicator />}

          {status === 'error' && errorText && (
            <div
              role="alert"
              className="flex items-center justify-between gap-3 rounded border border-red-200 bg-red-50 px-3 py-2.5 text-sm text-destructive"
            >
              <span className="flex items-center gap-2">
                <AlertCircle size={14} className="flex-shrink-0" aria-hidden="true" />
                {errorText}
              </span>
              <button
                onClick={handleRetry}
                className="flex flex-shrink-0 items-center gap-1.5 rounded border border-destructive/30 px-2.5 py-1 text-xs font-medium text-destructive transition-colors duration-150 hover:bg-red-100"
              >
                <RotateCw size={12} aria-hidden="true" /> Retry
              </button>
            </div>
          )}

          <div ref={bottomRef} />
        </div>
      </div>

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
                disabled={isLoading}
                rows={1}
                className="w-full resize-none rounded border border-input bg-background px-4 py-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary disabled:opacity-60"
                style={{ lineHeight: '1.5', minHeight: '44px', maxHeight: '120px' }}
                aria-label="Message input"
              />
            </div>
            <button
              type="submit"
              disabled={!input.trim() || isLoading}
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

      <style>{`
        @keyframes bounce {
          0%, 80%, 100% { transform: translateY(0); opacity: 0.4; }
          40% { transform: translateY(-5px); opacity: 1; }
        }
      `}</style>
    </div>
  )
}
