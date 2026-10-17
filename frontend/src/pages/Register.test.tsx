import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { register } from '../api/auth';
import { ApiError } from '../api/errors';
import type { User } from '../lib/types';
import { Register } from './Register';

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

// Корректный паттерн: vi.mocked требует уже импортированный объект.
const mockRegister = vi.mocked(register);

// Полный объект User по контракту lib/types.ts.
const user: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Анна',
  created_at: '2024-01-01T00:00:00Z',
};

// Рендерим внутри QueryClientProvider + MemoryRouter: QueryClient используется
// приложением (App.tsx), а Register содержит Link (нужен Router context).
function renderRegister() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <Register />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

// react-hook-form читает значения из onChange-обработчиков, поэтому заполняем
// поля fireEvent.change.
function fillAndSubmit(password: string) {
  fireEvent.change(screen.getByLabelText('Имя'), {
    target: { value: 'Анна' },
  });
  fireEvent.change(screen.getByLabelText('Email'), {
    target: { value: 'a@b.c' },
  });
  fireEvent.change(screen.getByLabelText('Пароль'), {
    target: { value: password },
  });
  screen.getByRole('button', { name: 'Зарегистрироваться' }).click();
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('Register', () => {
  it('валидная форма вызывает register с полным payload', async () => {
    mockRegister.mockResolvedValue(user);
    renderRegister();
    fillAndSubmit('password123');

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
  });

  it('password < 8 символов: ошибка zod, запрос не уходит', async () => {
    renderRegister();
    fillAndSubmit('short');

    await waitFor(() => {
      expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    });
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('ApiError conflict: баннер о существующем пользователе', async () => {
    // Ошибка создаётся в тесте, поэтому instanceof в компоненте сработает.
    mockRegister.mockRejectedValue(new ApiError('conflict', 'exists'));
    renderRegister();
    fillAndSubmit('password123');

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Пользователь с таким email уже существует',
      );
    });
  });
});
