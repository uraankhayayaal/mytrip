import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';
import { Register } from './Register';

// Mock the auth API
const mockRegister = vi.fn();
vi.mock('../api/auth', () => ({
  register: mockRegister,
}));

describe('Register', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock window.location
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: '' },
    });
  });

  it('renders register form correctly', () => {
    render(<Register />);
    
    expect(screen.getByText('Создать аккаунт')).toBeInTheDocument();
    expect(screen.getByLabelText('Имя')).toBeInTheDocument();
    expect(screen.getByLabelText('Email адрес')).toBeInTheDocument();
    expect(screen.getByLabelText('Пароль')).toBeInTheDocument();
    expect(screen.getByText('Зарегистрироваться')).toBeInTheDocument();
    expect(screen.getByText('Уже есть аккаунт? Войти')).toBeInTheDocument();
  });

  it('submits valid form and calls register function', async () => {
    const user = userEvent.setup();
    mockRegister.mockResolvedValue({ id: '1', email: 'test@example.com', name: 'Test User', created_at: '2023-01-01' });
    
    render(<Register />);
    
    await user.type(screen.getByLabelText('Имя'), 'Test User');
    await user.type(screen.getByLabelText('Email адрес'), 'test@example.com');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByText('Зарегистрироваться'));
    
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
        name: 'Test User',
      });
      expect(window.location.href).toBe('/login');
    });
  });

  it('shows validation error for password less than 8 characters', async () => {
    const user = userEvent.setup();
    render(<Register />);
    
    await user.type(screen.getByLabelText('Пароль'), 'pass');
    await user.click(screen.getByText('Зарегистрироваться'));
    
    expect(screen.getByText('Минимум 8 символов')).toBeInTheDocument();
  });

  it('shows error message for conflict (user already exists)', async () => {
    const user = userEvent.setup();
    mockRegister.mockRejectedValue({
      response: {
        data: {
          error: {
            code: 'conflict',
            message: 'Conflict',
          },
        },
      },
    });
    
    render(<Register />);
    
    await user.type(screen.getByLabelText('Имя'), 'Test User');
    await user.type(screen.getByLabelText('Email адрес'), 'test@example.com');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByText('Зарегистрироваться'));
    
    await waitFor(() => {
      expect(screen.getByText('Пользователь с таким email уже существует')).toBeInTheDocument();
    });
  });

  it('shows validation errors from backend for validation error', async () => {
    const user = userEvent.setup();
    mockRegister.mockRejectedValue({
      response: {
        data: {
          error: {
            code: 'validation',
            message: 'Validation failed',
            details: {
              name: ['Name is required'],
            },
          },
        },
      },
    });
    
    render(<Register />);
    
    await user.type(screen.getByLabelText('Имя'), 'Test User');
    await user.type(screen.getByLabelText('Email адрес'), 'test@example.com');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByText('Зарегистрироваться'));
    
    await waitFor(() => {
      expect(screen.getByText('Name is required')).toBeInTheDocument();
    });
  });
});