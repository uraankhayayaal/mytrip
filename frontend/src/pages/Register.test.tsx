import { render, screen, waitFor, fireEvent } from '@testing-library/react';
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

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: vi.fn() };
});

// Корректный паттерн: vi.mocked требует уже импортированный объект (не require).
const mockRegister = vi.mocked(register);
const mockUseNavigate = vi.mocked(
  (await import('react-router-dom')).useNavigate,
) as unknown as ReturnType<typeof vi.fn>;

const user: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Анна',
  created_at: '2024-01-01T00:00:00Z',
};

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

function fillAndSubmit(password: string) {
  fireEvent.change(screen.getByLabelText('Имя'), { target: { value: 'Анна' } });
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'a@b.c' } });
  fireEvent.change(screen.getByLabelText('Пароль'), { target: { value: password } });
  return screen.getByRole('button', { name: 'Зарегистрироваться' });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    mockRegister.mockResolvedValue(user);
    const navigate = mockUseNavigate() as ReturnType<typeof vi.fn>;
    renderRegister();
    fillAndSubmit('password123').click();

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
    expect(navigate).toHaveBeenCalledWith('/login');
  });

  it('password < 8 символов: ошибка zod, запрос не уходит', async () => {
    renderRegister();
    fillAndSubmit('short').click();

    await waitFor(() => {
      expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    });
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('ApiError conflict: баннер о существующем пользователе', async () => {
    mockRegister.mockRejectedValue(new ApiError('conflict', 'exists'));
    renderRegister();
    fillAndSubmit('password123').click();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Пользователь с таким email уже существует',
      );
    });
  });
});
