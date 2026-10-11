import React from 'react'
import { Link } from 'react-router-dom'
import { logout } from '../api/auth'

export function Layout({ children }: { children: React.ReactNode }): JSX.Element {
  const handleLogout = async () => {
    try {
      await logout()
      window.location.href = '/login'
    } catch (error) {
      // If logout fails, still redirect to login
      window.location.href = '/login'
    }
  }

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
          <nav className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Link to="/" className="text-xl font-bold text-gray-900">MyTrip</Link>
              <Link to="/" className="text-gray-600 hover:text-gray-900">Поездки</Link>
              <Link to="/calendar" className="text-gray-600 hover:text-gray-900">Календарь</Link>
            </div>
            <button 
              onClick={handleLogout}
              className="text-gray-600 hover:text-gray-900"
            >
              Выйти
            </button>
          </nav>
        </div>
      </header>

      <main className="flex-grow mx-auto max-w-6xl px-4 py-6">
        {children}
      </main>
    </div>
  )
}