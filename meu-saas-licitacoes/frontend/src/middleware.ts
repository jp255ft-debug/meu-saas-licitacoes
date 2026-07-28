import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('access_token')?.value
  const path = request.nextUrl.pathname

  // Rotas públicas que não precisam de token
  const publicRoutes = ['/login', '/cadastro', '/']
  const isAuthPage = path === '/login' || path === '/cadastro'
  const isPublic = publicRoutes.includes(path)

  // Se não tem token e tenta acessar rota protegida → redireciona para login
  if (!token && !isPublic) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('redirect', path)
    return NextResponse.redirect(loginUrl)
  }

  // Se tem token e tenta acessar login/cadastro → redireciona para painel
  if (token && isAuthPage) {
    return NextResponse.redirect(new URL('/painel', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/painel/:path*', '/login', '/cadastro'],
}