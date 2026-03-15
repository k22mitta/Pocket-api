'use client'

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import { useRouter } from 'next/navigation'
import { api } from './api'

// Fixed credentials for the shared demo account.
// The backend requires a real JWT, so loginDemo registers/logs-in this account
// on first use and obtains a proper token. isDemo is detected by email, not by
// a token sentinel, so the backend treats demo users like any other user.
const DEMO_EMAIL = 'demo@example.com'
const DEMO_PASSWORD = 'demo-pocket'

interface AuthState {
  token: string | null
  email: string | null
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  signup: (email: string, password: string, name: string) => Promise<void>
  loginDemo: () => Promise<void>
  logout: () => void
  isAuthenticated: boolean
  isDemo: boolean
  hydrated: boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ token: null, email: null })
  const [hydrated, setHydrated] = useState(false)
  const router = useRouter()

  // Read persisted auth from localStorage after mount to avoid SSR/hydration mismatch
  useEffect(() => {
    const token = localStorage.getItem('pocket_token')
    const email = localStorage.getItem('pocket_email')
    if (token) setState({ token, email })
    setHydrated(true)
  }, [])

  const persistState = useCallback((newState: AuthState) => {
    if (newState.token) {
      localStorage.setItem('pocket_token', newState.token)
      if (newState.email) localStorage.setItem('pocket_email', newState.email)
    } else {
      localStorage.removeItem('pocket_token')
      localStorage.removeItem('pocket_email')
    }
    setState(newState)
  }, [])

  const login = useCallback(
    async (email: string, password: string) => {
      const { token } = await api.auth.login(email, password)
      persistState({ token, email })
      router.push('/dashboard')
    },
    [router, persistState],
  )

  const signup = useCallback(
    async (email: string, password: string, name: string) => {
      const { token } = await api.auth.signup(email, password, name)
      persistState({ token, email })
      router.push('/dashboard')
    },
    [router, persistState],
  )

  const loginDemo = useCallback(async () => {
    let token: string
    try {
      const res = await api.auth.login(DEMO_EMAIL, DEMO_PASSWORD)
      token = res.token
    } catch {
      const res = await api.auth.signup(DEMO_EMAIL, DEMO_PASSWORD, 'Demo User')
      token = res.token
    }
    persistState({ token, email: DEMO_EMAIL })
    router.push('/dashboard')
  }, [router, persistState])

  const logout = useCallback(() => {
    persistState({ token: null, email: null })
    router.push('/login')
  }, [router, persistState])

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        signup,
        loginDemo,
        logout,
        isAuthenticated: !!state.token,
        isDemo: state.email === DEMO_EMAIL,
        hydrated,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be called inside <AuthProvider>')
  return ctx
}
