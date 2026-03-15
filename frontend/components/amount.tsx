'use client'

import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'

export function formatMoney(
  amount: number,
  opts?: { compact?: boolean; showSign?: boolean },
) {
  const abs = Math.abs(amount)
  const formatted = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    notation: opts?.compact && abs >= 10_000 ? 'compact' : 'standard',
    maximumFractionDigits: 2,
    minimumFractionDigits: 2,
  }).format(abs)

  if (opts?.showSign) {
    return amount >= 0 ? `+${formatted}` : `−${formatted}`
  }
  return amount < 0 ? `−${formatted}` : formatted
}

export function useCountUp(target: number, duration = 1_100) {
  const [display, setDisplay] = useState(0)

  useEffect(() => {
    let rafId: number
    let startTime: number | null = null

    const step = (timestamp: number) => {
      if (startTime === null) startTime = timestamp
      const elapsed = timestamp - startTime
      const progress = Math.min(elapsed / duration, 1)
      // ease-out cubic
      const eased = 1 - Math.pow(1 - progress, 3)
      setDisplay(target * eased)
      if (progress < 1) {
        rafId = requestAnimationFrame(step)
      } else {
        setDisplay(target)
      }
    }

    rafId = requestAnimationFrame(step)
    return () => cancelAnimationFrame(rafId)
  }, [target, duration])

  return display
}

interface HeroAmountProps {
  amount: number
  className?: string
  animate?: boolean
}

export function HeroAmount({ amount, className, animate = true }: HeroAmountProps) {
  const display = useCountUp(animate ? amount : 0, 1_100)
  const value = animate ? display : amount

  return (
    <span
      className={cn(
        'money font-semibold tracking-tight text-foreground',
        className,
      )}
    >
      {formatMoney(value)}
    </span>
  )
}

interface LedgerAmountProps {
  amount: number
  className?: string
}

export function LedgerAmount({ amount, className }: LedgerAmountProps) {
  const isPositive = amount >= 0
  return (
    <span
      className={cn(
        'money whitespace-nowrap text-sm tabular-nums',
        isPositive ? 'text-foreground' : 'text-foreground',
        className,
      )}
    >
      {amount >= 0 ? '+' : '−'}
      {new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
      }).format(Math.abs(amount))}
    </span>
  )
}
