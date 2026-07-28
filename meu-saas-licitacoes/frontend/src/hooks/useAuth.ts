'use client'

import { useRouter } from 'next/navigation'
import { api } from '../lib/api'

export function useAuth() {
  const router = useRouter()

  const login = async (username: string, password: string) => {
    const response = await api.post('/api/v1/auth/token', {
      username,
      password,
    })
    // Token é automaticamente armazenado em cookie httpOnly pelo backend
    // Nada a fazer no frontend — apenas redirecionar
    return response.data
  }

  const logout = async () => {
    try {
      await api.post('/api/v1/auth/logout')
    } finally {
      router.push('/login')
    }
  }

  return { login, logout }
}