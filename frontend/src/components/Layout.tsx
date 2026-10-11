import React from 'react';
import { Link } from 'react-router-dom';
import { logout } from '../api/auth';

export function Layout({ children }: { children: React.ReactNode }): JSX.Element {
  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow">
        <div className="mx-auto max-w-6xl px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <Link to="/" className="text-xl font-bold text-indigo-600">
                MyTrip
              </Link>
            </div>
            <nav className="hidden md:flex space-x-8">
              <Link to="/" className="text-gray-700 hover:text-indigo-600">
                Поездки
              </Link>
              <Link to="/calendar" className="text-gray-700 hover:text-indigo-600">
                Календарь
              </Link>
            </nav>
            <div className="flex items-center">
              <button onClick={handleLogout} className="text-gray-700 hover:text-indigo-600">
                Выйти
              </button>
            </div>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-6">{children}</main>
    </div>
  );
}
