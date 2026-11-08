import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { login } from '../api/auth';
import { setAccessToken } from '../api/client';
import { ApiError } from '../api/errors';
import { Login } from './Login';

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}));

vi.mock('../api/client', () => ({
  setAccessToken: vi.fn(),
}));

function renderLogin() {
  return render(
    <MemoryRouter>
      <Login />
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

describe('Login', () => {
  it('валидная форма вызывает login и setAccessToken, редирект на /', async () => {
    vi.mocked(login).mockResolvedValue({
      access_token: 'tok-123',
      user: { id: 'u1', email: 'a@b.c', name: 'A', created_at: '' },
    });
    renderLogin();
    
    // Use fireEvent directly on form fields + dispatch submit event manually
    const emailInput = screen.getByLabelText('Email');
    const passwordInput = screen.getByLabelText('Пароль');
    
    // Set values and dispatch both input and change events
    emailInput.value = 'a@b.c';
    fireEvent.input(emailInput, { target: { value: 'a@b.c' } });
    fireEvent.change(emailInput, { target: { value: 'a@b.c' } });
    
    passwordInput.value = 'password123';
    fireEvent.input(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    
    // Dispatch submit event with cancelable flag to prevent native behavior
    const form = document.querySelector('form')!;
    const submitEvent = new SubmitEvent('submit', { cancelable: true, bubbles: true });
    form.dispatchEvent(submitEvent);
    
    await waitFor(() => {
      expect(login).toHaveBeenCalledWith({ email: 'a@b.c', password: 'password123' });
    });
    expect(setAccessToken).toHaveBeenCalledWith('tok-123');
    expect(window.location.href).toBe('/');
  });

  it('пустые поля: ошибки zod, запрос не уходит', async () => {
    renderLogin();
    const form = document.querySelector('form')!;
    const submitEvent = new SubmitEvent('submit', { cancelable: true, bubbles: true });
    form.dispatchEvent(submitEvent);
    expect(screen.getByText(/Invalid email/i)).toBeInTheDocument();
    expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();
  });

  it('ApiError unauthorized: баннер «Неверный email или пароль»', async () => {
    vi.mocked(login).mockRejectedValue(new ApiError('unauthorized', 'bad credentials'));
    renderLogin();
    
    const emailInput = screen.getByLabelText('Email');
    const passwordInput = screen.getByLabelText('Пароль');
    
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
      expect(screen.getByRole('alert')).toHaveTextContent('Неверный email или пароль');
    });
  });
});
