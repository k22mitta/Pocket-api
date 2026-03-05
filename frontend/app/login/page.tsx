'use client'

import { useState, type FormEvent } from 'react'
import { useAuth } from '@/lib/auth-context'
import { ApiError } from '@/lib/api'

type Mode = 'signin' | 'signup'

export default function LoginPage() {
  const { login, signup, loginDemo, isAuthenticated } = useAuth()
  const [mode, setMode] = useState<Mode>('signin')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // If somehow already authenticated, this handles the redirect edge case
  if (isAuthenticated) return null

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      if (mode === 'signin') {
        await login(email, password)
      } else {
        await signup(email, password)
      }
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 401) {
          setError('Email or password is incorrect.')
        } else if (err.status === 409) {
          setError('An account with that email already exists.')
        } else {
          setError(`Server error (${err.status}). Check that your API is running.`)
        }
      } else if (err instanceof TypeError) {
        setError('Could not reach the server. Check that your API is running at the configured URL.')
      } else {
        setError('Something unexpected happened. Try again.')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <div className="w-full max-w-sm">

        {/* Wordmark */}
        <h1 className="mb-10 font-serif text-4xl font-semibold tracking-tight text-foreground">
          Pocket
        </h1>

        {/* Mode toggle */}
        <div className="mb-8 flex gap-0 border-b border-border">
          <button
            type="button"
            onClick={() => { setMode('signin'); setError(null) }}
            className={`pb-3 pr-6 text-sm transition-all duration-150 ${
              mode === 'signin'
                ? 'border-b-2 border-foreground font-medium text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            style={mode === 'signin' ? { marginBottom: '-1px' } : undefined}
          >
            Sign in
          </button>
          <button
            type="button"
            onClick={() => { setMode('signup'); setError(null) }}
            className={`pb-3 pr-6 text-sm transition-all duration-150 ${
              mode === 'signup'
                ? 'border-b-2 border-foreground font-medium text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            style={mode === 'signup' ? { marginBottom: '-1px' } : undefined}
          >
            Create account
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex flex-col gap-4" noValidate>
          <div className="flex flex-col gap-1.5">
            <label htmlFor="email" className="text-sm text-foreground">
              Email address
            </label>
            <input
              id="email"
              type="email"
              required
              autoComplete="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="you@example.com"
              className="rounded border border-input bg-background px-3 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary transition-all duration-150"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="password" className="text-sm text-foreground">
              Password
            </label>
            <input
              id="password"
              type="password"
              required
              autoComplete={mode === 'signin' ? 'current-password' : 'new-password'}
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder={mode === 'signup' ? 'At least 8 characters' : '••••••••'}
              className="rounded border border-input bg-background px-3 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary transition-all duration-150"
            />
          </div>

          {/* Error */}
          {error && (
            <p role="alert" className="rounded bg-red-50 px-3 py-2.5 text-sm text-destructive border border-red-200">
              {error}
            </p>
          )}

          {/* Submit */}
          <button
            type="submit"
            disabled={loading}
            className="mt-1 rounded bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition-all duration-150 hover:opacity-90 disabled:opacity-50"
          >
            {loading
              ? mode === 'signin' ? 'Signing in…' : 'Creating account…'
              : mode === 'signin' ? 'Sign in' : 'Create account'}
          </button>
        </form>

        {/* Demo link */}
        <div className="mt-8 border-t border-border pt-6">
          <p className="text-sm text-muted-foreground">
            No backend set up yet?{' '}
            <button
              type="button"
              onClick={loginDemo}
              className="text-foreground underline underline-offset-4 hover:text-primary transition-colors duration-150"
            >
              View demo
            </button>
          </p>
        </div>
      </div>
    </div>
  )
}
