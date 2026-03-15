'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth-context'
import Nav from '@/components/nav'

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, hydrated } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (hydrated && !isAuthenticated) {
      router.replace('/login')
    }
  }, [isAuthenticated, hydrated, router])

  // Wait for localStorage to be read before showing content or redirecting
  if (!hydrated || !isAuthenticated) return null

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Nav />
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  )
}
