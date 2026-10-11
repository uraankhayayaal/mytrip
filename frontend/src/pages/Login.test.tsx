import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';
import { Login } from './Login';

// Mock the auth API
vi.mock('../api/auth', () => ({
  login: vi.fn(),
}));

const mockLogin = vi.fn();

vi.mock('../api/auth', () => ({
  login: mockLogin,
}));

// Mock the client API
vi.mock('../api/client', () => ({
  setAccessToken: vi.fn(),
}));

const mockSetAccessToken = vi.fn();

vi.mock('../api/client', () => ({
  setAccessToken: mockSetAccessToken,
}));

describe('Login', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock window.location
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { href: '' },
    });
  });

  it('renders login form correctly', () => {
    render(<Login />);
    
    expect(screen.getByText('Войти в аккаунт')).toBeInTheDocument();
    expect(screen.getByLabelText('Email адрес')).toBeInTheDocument();
    expect(screen.getByLabelText('Пароль')).toBeInTheDocument();
    expect(screen.getByText('Войти')).toBeInTheDocument();
    expect(screen.getByText('Нет аккаунта? Регистрация')).toBeInTheDocument();
  });

  it('submits valid form and calls login function', async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValue({ access_token: 'test-token', user: { id: '1', email: 'test@example.com', name: 'Test User', created_at: '2023-01-01' } });
    
    render(<Login />);
    
    await user.type(screen.getByLabelText('Email адрес'), 'test@example.com');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByText('Войти'));
    
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      });
      expect(mockSetAccessToken).toHaveBeenCalledWith('test-token');
      expect(window.location.href).toBe('/');
    });
  });

  it('shows validation errors for empty fields', async () => {
    const user = userEvent.setup();
    render(<Login />);
    
    await user.click(screen.getByText('Войти'));
    
    expect(screen.getByText('Email адрес обязателен')).toBeInTheDocument();
    expect(screen.getByText('Пароль обязателен')).toBeInTheDocument();
  });

  it('shows error message for unauthorized login', async () => {
    const user = userEvent.setup();
    mockLogin.mockRejectedValue({
      response: {
        data: {
          error: {
            code: 'unauthorized',
            message: 'Unauthorized',
          },
        },
      },
    });
    
    render(<Login />);
    
    await user.type(screen.getByLabelText('Email адрес'), 'test@example.com');
    await user.type(screen.getByLabelText('Пароль'), 'password123');
    await user.click(screen.getByText('Войти'));
    
    await waitFor(() => {
      expect(screen.getByText('Неверный email или пароль')).toBeInTheDocument();
    });
  });
});
