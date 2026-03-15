import { Analytics } from '@vercel/analytics/next'
import type { Metadata, Viewport } from 'next'
import { Inter, Fraunces } from 'next/font/google'
import Script from 'next/script'
import { AuthProvider } from '@/lib/auth-context'
import './globals.css'

const inter = Inter({
  subsets: ['latin'],
  variable: '--inter',
})

const fraunces = Fraunces({
  subsets: ['latin'],
  variable: '--fraunces',
})

export const metadata: Metadata = {
  title: 'Pocket — Personal Finance',
  description: 'Track your accounts, transactions, and budgets in one place.',
  generator: 'v0.app',
}

export const viewport: Viewport = {
  colorScheme: 'light',
  themeColor: '#F7F5F0',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" className={`${inter.variable} ${fraunces.variable} bg-background`}>
      <body className="font-sans antialiased">
        {/*
          react-plaid-link's usePlaidLink loads this script itself on mount via
          an internal useScript hook that has no way to abort an in-flight
          fetch. React 18 StrictMode's dev-only mount -> cleanup -> remount
          cycle removes and re-adds the tag while the first request is still
          in flight, so the script executes twice and logs Plaid's "embedded
          more than once" warning — even with zero clicks, on a plain mount.
          Loading it here with beforeInteractive means it's already a
          <script> tag in the initial HTML, executed once before React (and
          therefore StrictMode) ever runs; usePlaidLink's own
          checkForExisting finds this tag, treats it as already loaded, and
          never creates a second one, regardless of how many times its
          effect fires.
        */}
        <Script src="https://cdn.plaid.com/link/v2/stable/link-initialize.js" strategy="beforeInteractive" />
        <AuthProvider>{children}</AuthProvider>
        {process.env.NODE_ENV === 'production' && <Analytics />}
      </body>
    </html>
  )
}
