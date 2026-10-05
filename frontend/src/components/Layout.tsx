import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { logout } from '../api/auth';

export function Layout({ children }: { children: ReactNode }): JSX.Element {
  async function handleLogout() {
    try {
      await logout();
    } catch {
      // не критично: редирект происходит в любом случае
    } finally {
      window.location.href = '/login';
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b border-gray-200 bg-white">
        <nav className="mx-auto flex max-w-6xl flex-wrap items-center gap-2 px-4 py-3 sm:gap-4">
          <Link to="/" className="text-lg font-bold text-blue-600">
            MyTrip
          </Link>
          <div className="ml-auto flex items-center gap-2 sm:gap-3">
            <Link to="/" className="rounded px-2 py-1 text-sm text-gray-700 hover:bg-gray-100">
              Поездки
            </Link>
            <Link
              to="/calendar"
              className="rounded px-2 py-1 text-sm text-gray-700 hover:bg-gray-100"
            >
              Календарь
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              className="rounded bg-gray-100 px-3 py-1 text-sm font-medium text-gray-800 hover:bg-gray-200"
            >
              Выйти
            </button>
          </div>
        </nav>
      </header>
      <main className="mx-auto max-w-6xl px-4">{children}</main>
    </div>
  );
}
