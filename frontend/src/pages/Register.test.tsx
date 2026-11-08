import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { register } from '../api/auth';
import { Register } from './Register';

vi.mock('../api/auth', () => ({
  register: vi.fn(),
}));

function renderRegister() {
  return render(
    <MemoryRouter>
      <Register />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { href: 'http://localhost:3000/' },
    writable: true,
    configurable: true,
  });
});

describe('Register', () => {
  it('валидная форма вызывает register, редирект на /login', async () => {
    vi.mocked(register).mockResolvedValue({
      id: 'u1',
      email: 'a@b.c',
      name: 'A',
      created_at: '',
    });
    renderRegister();
    
    const nameInput = screen.getByLabelText('Имя');
    const emailInput = screen.getByLabelText('Email');
    const passwordInput = screen.getByLabelText('Пароль');
    
    nameInput.value = 'Анна';
    fireEvent.input(nameInput, { target: { value: 'Анна' } });
    fireEvent.change(nameInput, { target: { value: 'Анна' } });
    
    emailInput.value = 'a@b.c';
    fireEvent.input(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    
    passwordInput.value = 'password123';
    fireEvent.input(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    
    const form = document.querySelector('form')!;
    const submitEvent = new SubmitEvent('submit', { cancelable: true, bubbles: true });
    form.dispatchEvent(submitEvent);
    
    await waitFor(() => {
      expect(register).toHaveBeenCalledWith({
        email: 'a@b.c',
        password: 'password123',
        name: 'Анна',
      });
    });
    expect(window.location.href).toBe('/login');
  });

  it('password < 8 символов: ошибка zod, запрос не уходит', async () => {
    renderRegister();
    
    const form = document.querySelector('form')!;
    const submitEvent = new SubmitEvent('submit', { cancelable: true, bubbles: true });
    form.dispatchEvent(submitEvent);
    
    expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    expect(register).not.toHaveBeenCalled();
  });
});
