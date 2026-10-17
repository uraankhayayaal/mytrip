import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
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

const mockRegister = vi.mocked(register);

const user: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Анна',
  created_at: '2024-01-01T00:00:00Z',
};

// QueryClientProvider как в App.tsx — единый паттерн для всех страниц.
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
  const nameInput = screen.getByLabelText('Имя') as HTMLInputElement;
  const emailInput = screen.getByLabelText('Email') as HTMLInputElement;
  const passwordInput = screen.getByLabelText('Пароль') as HTMLInputElement;
  nameInput.value = 'Анна';
  nameInput.dispatchEvent(new Event('input', { bubbles: true }));
  emailInput.value = 'a@b.c';
  emailInput.dispatchEvent(new Event('input', { bubbles: true }));
  passwordInput.value = password;
  passwordInput.dispatchEvent(new Event('input', { bubbles: true }));
  return screen.getByRole('button', { name: 'Зарегистрироваться' });
}

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { href: window.location.href },
    writable: true,
    configurable: true,
  });
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    mockRegister.mockResolvedValue(user);
    renderRegister();
    fillAndSubmit('password123').click();

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
    expect(window.location.href).toBe('/login');
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
      expect(screen.getByRole('alert')).toHaveTextContent('Пользователь с таким email уже существует');
    });
  });
});
