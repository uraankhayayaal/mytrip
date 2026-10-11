import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from 'react-router-dom';
import { register as registerUser } from '../api/auth';
import { ApiError } from '../api/errors';
import { registerSchema, type RegisterInput } from '../lib/schemas';

export function Register(): JSX.Element {
  const [serverError, setServerError] = useState<string | null>(null);
  const { register, handleSubmit, formState } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  });

  async function onSubmit(values: RegisterInput): Promise<void> {
    setServerError(null);
    try {
      await registerUser(values);
      window.location.href = '/login';
    } catch (e) {
      if (e instanceof ApiError && e.code === 'conflict') {
        setServerError('Пользователь с таким email уже существует');
      } else if (e instanceof ApiError && e.code === 'validation') {
        setServerError('Проверьте правильность заполнения полей');
      } else {
        setServerError('Не удалось зарегистрироваться. Попробуйте позже.');
      }
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow">
        <h1 className="mb-4 text-2xl font-bold text-gray-900">Регистрация</h1>
        {serverError && (
          <div role="alert" className="mb-4 rounded bg-red-100 px-3 py-2 text-sm text-red-700">
            {serverError}
          </div>
        )}
        <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-4">
          <div>
            <label htmlFor="name" className="mb-1 block text-sm font-medium text-gray-700">
              Имя
            </label>
            <input
              id="name"
              type="text"
              autoComplete="name"
              className="w-full rounded border border-gray-300 px-3 py-2 text-gray-900 focus:border-blue-500 focus:outline-none"
              {...register('name')}
            />
            {formState.errors.name && (
              <p className="mt-1 text-sm text-red-600">{formState.errors.name.message}</p>
            )}
          </div>
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
              autoComplete="new-password"
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
            Зарегистрироваться
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-gray-600">
          Уже есть аккаунт?{' '}
          <Link to="/login" className="text-blue-600 hover:underline">
            Войти
          </Link>
        </p>
      </div>
    </div>
  );
}
