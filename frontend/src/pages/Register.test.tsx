import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { register } from '../api/auth';
import { Register } from './Register';
import type { User } from '../lib/types';

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

const mockRegister = vi.mocked(register);

const user: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Анна',
  created_at: '2024-01-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { href: window.location.href },
    writable: true,
  });
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    mockRegister.mockResolvedValue(user);
    render(<Register />);
    const inputs = screen.getAllByRole('textbox');
    inputs[0].value = 'Анна';
    inputs[1].value = 'a@b.c';
    inputs[2].value = 'password123';
    const submit = screen.getByRole('button', { name: 'Зарегистрироваться' });
    submit.click();
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
    expect(window.location.href).toBe('/login');
  });
});
