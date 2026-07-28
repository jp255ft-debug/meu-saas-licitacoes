"use client"

import {
  CalendarCheck,
  FileText,
  LayoutDashboard,
  LogOut,
  Settings,
  Star,
} from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { Button } from "../../components/ui/button"
import { useAuth } from "../../hooks/useAuth"

const menuItems = [
  { label: "Dashboard", href: "/painel", icon: LayoutDashboard },
  { label: "Editais", href: "/painel/editais", icon: FileText },
  { label: "Favoritos", href: "/painel/favoritos", icon: Star },
  { label: "PCA", href: "/painel/pca", icon: CalendarCheck },
  { label: "Configurações", href: "/painel/configuracoes", icon: Settings },
]

export default function PainelLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const { logout } = useAuth()

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 bg-gray-900 text-white flex flex-col">
        <div className="p-6 border-b border-gray-700">
          <h1 className="text-xl font-bold">Licitação SaaS</h1>
          <p className="text-sm text-gray-400 mt-1">Painel de Controle</p>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {menuItems.map((item) => {
            const Icon = item.icon
            const isActive = pathname === item.href
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${isActive
                    ? "bg-blue-600 text-white"
                    : "text-gray-300 hover:bg-gray-800 hover:text-white"
                  }`}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="p-4 border-t border-gray-700">
          <Button
            variant="ghost"
            className="w-full justify-start text-gray-300 hover:text-white hover:bg-gray-800"
            onClick={logout}
          >
            <LogOut className="h-4 w-4 mr-2" />
            Sair
          </Button>
        </div>
      </aside>

      {/* Conteúdo principal */}
      <main className="flex-1 bg-gray-50 overflow-auto">
        {children}
      </main>
    </div>
  )
}