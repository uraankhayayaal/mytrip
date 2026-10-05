import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from 'react-router-dom';
import { login } from '../api/auth';
import { setAccessToken } from '../api/client';
import { ApiError } from '../api/errors';
import { loginSchema, type LoginInput } from '../lib/schemas';

export function Login(): JSX.Element {
  const [serverError, setServerError] = useState<string | null>(null);
  const { register, handleSubmit, formState } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  });

  async function onSubmit(values: LoginInput): Promise<void> {
    setServerError(null);
    try {
      const res = await login(values);
      setAccessToken(res.access_token);
      window.location.href = '/';
    } catch (e) {
      if (e instanceof ApiError && e.code === 'unauthorized') {
        setServerError('Неверный email или пароль');
      } else if (e instanceof ApiError && e.code === 'validation') {
        setServerError('Проверьте правильность заполнения полей');
      } else {
        setServerError('Не удалось войти. Попробуйте позже.');
      }
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow">
        <h1 className="mb-4 text-2xl font-bold text-gray-900">Вход</h1>
        {serverError && (
          <div role="alert" className="mb-4 rounded bg-red-100 px-3 py-2 text-sm text-red-700">
            {serverError}
          </div>
        )}
        <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-4">
          <div>
            <label htmlFor="email" className="mb-1 block text-sm font-medium text-gray-700">
              Email
            </label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none"
              {...register('email')}
            />
            {formState.errors.email && (
              <p className="mt-1 text-sm text-red-600">{formState.errors.email.message}</p>
            )}
          </div>
          <div>
            <label htmlFor="password" className="mb-1 block text-sm font-medium text-gray-700">
              Пароль
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none"
              {...register('password')}
            />
            {formState.errors.password && (
              <p className="mt-1 text-sm text-red-600">{formState.errors.password.message}</p>
            )}
          </div>
          <button
            type="submit"
            className="w-full rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700"
          >
            Войти
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-gray-600">
          Нет аккаунта?{' '}
          <Link to="/register" className="text-blue-600 hover:underline">
            Регистрация
          </Link>
        </p>
      </div>
    </div>
  );
}
