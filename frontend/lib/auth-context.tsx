'use client'

import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from 'react'
import { useRouter } from 'next/navigation'
import { api } from './api'

interface AuthState {
  token: string | null
  email: string | null
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  signup: (email: string, password: string) => Promise<void>
  loginDemo: () => void
  logout: () => void
  isAuthenticated: boolean
  isDemo: boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ token: null, email: null })
  const router = useRouter()

  const login = useCallback(
    async (email: string, password: string) => {
      const { token } = await api.auth.login(email, password)
      setState({ token, email })
      router.push('/dashboard')
    },
    [router],
  )

  const signup = useCallback(
    async (email: string, password: string) => {
      const { token } = await api.auth.signup(email, password)
      setState({ token, email })
      router.push('/dashboard')
    },
    [router],
  )

  const loginDemo = useCallback(() => {
    setState({ token: '__demo__', email: 'demo@example.com' })
    router.push('/dashboard')
  }, [router])

  const logout = useCallback(() => {
    setState({ token: null, email: null })
    router.push('/login')
  }, [router])

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        signup,
        loginDemo,
        logout,
        isAuthenticated: !!state.token,
        isDemo: state.token === '__demo__',
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
