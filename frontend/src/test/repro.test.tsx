import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, vi } from 'vitest';
import { Login } from '../pages/Login';
import { login } from '../api/auth';

vi.mock('../api/auth', () => ({ login: vi.fn() }));
vi.mock('../api/client', () => ({ setAccessToken: vi.fn() }));

describe('repro', () => {
  it('debug', () => {
    render(<Login />);
    const email = screen.getByLabelText('Email');
    fireEvent.change(email, { target: { value: 'a@b.c' } });
    console.log('EMAIL VALUE AFTER CHANGE:', email.value);
    console.log('EMAIL HTML:', email.outerHTML);
  });
});
