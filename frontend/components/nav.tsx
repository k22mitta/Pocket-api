'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  Landmark,
  ArrowUpDown,
  PieChart,
  Sparkles,
  LogOut,
} from 'lucide-react'
import { useAuth } from '@/lib/auth-context'
import { cn } from '@/lib/utils'

const NAV_ITEMS = [
  { href: '/dashboard',    label: 'Dashboard',    icon: LayoutDashboard },
  { href: '/accounts',     label: 'Accounts',     icon: Landmark       },
  { href: '/transactions', label: 'Transactions', icon: ArrowUpDown    },
  { href: '/budgets',      label: 'Budgets',      icon: PieChart       },
  { href: '/ask',          label: 'Ask Pocket',   icon: Sparkles       },
]

export default function Nav() {
  const pathname = usePathname()
  const { logout, email, isDemo } = useAuth()

  return (
    <nav
      className="flex w-52 flex-shrink-0 flex-col bg-foreground text-background"
      aria-label="Main navigation"
    >
      <div className="px-6 pb-6 pt-8">
        <span className="font-serif text-xl font-semibold tracking-tight text-background">
          Pocket
        </span>
        {isDemo && (
          <span className="ml-2 rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider"
            style={{ background: '#C9A961', color: '#0B1210' }}>
            Demo
          </span>
        )}
      </div>

      <ul className="flex flex-col gap-0.5 px-3" role="list">
        {NAV_ITEMS.map(({ href, label, icon: Icon }) => {
          const active = pathname === href
          return (
            <li key={href}>
              <Link
                href={href}
                className={cn(
                  'flex items-center gap-3 rounded px-3 py-2.5 text-sm transition-all duration-150',
                  active
                    ? 'bg-white/10 text-background'
                    : 'text-background/60 hover:bg-white/6 hover:text-background',
                )}
                aria-current={active ? 'page' : undefined}
              >
                <Icon size={16} strokeWidth={1.75} aria-hidden="true" />
                {label}
              </Link>
            </li>
          )
        })}
      </ul>

      <div className="mt-auto border-t border-white/8 px-3 py-4">
        {email && (
          <p className="truncate px-3 pb-2 text-[11px] text-background/40">
            {email}
          </p>
        )}
        <button
          onClick={logout}
          className="flex w-full items-center gap-3 rounded px-3 py-2.5 text-sm text-background/60 transition-all duration-150 hover:bg-white/6 hover:text-background"
        >
          <LogOut size={16} strokeWidth={1.75} aria-hidden="true" />
          Sign out
        </button>
      </div>
    </nav>
  )
}
