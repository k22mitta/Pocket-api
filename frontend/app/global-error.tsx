'use client'

import { useEffect } from 'react'

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <html lang="en">
      <body className="font-sans antialiased" style={{ background: '#F7F5F0' }}>
        <div className="flex min-h-screen items-center justify-center px-4">
          <div className="w-full max-w-sm text-center">
            <h1 className="mb-2 text-xl font-semibold" style={{ color: '#0B1210' }}>
              Something went wrong
            </h1>
            <p className="mb-6 text-sm" style={{ color: '#8A8478' }}>
              Pocket failed to load. Try again, or reload the page.
            </p>
            <button
              onClick={reset}
              className="rounded px-4 py-2.5 text-sm font-medium transition-opacity hover:opacity-90"
              style={{ background: '#1C3D2E', color: '#F7F5F0' }}
            >
              Try again
            </button>
          </div>
        </div>
      </body>
    </html>
  )
}
