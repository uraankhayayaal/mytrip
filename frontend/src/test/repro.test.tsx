import { render, screen } from '@testing-library/react';
import { fireEvent } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { describe, it, expect } from 'vitest';

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8, 'Минимум 8 символов'),
});
type LoginInput = z.infer<typeof loginSchema>;

function DebugForm() {
  const { register, handleSubmit, formState } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  });
  return (
    <form onSubmit={handleSubmit((v) => console.log('SUBMIT', JSON.stringify(v)))} noValidate>
      <label htmlFor="email">Email</label>
      <input id="email" type="email" {...register('email')} />
      <label htmlFor="password">Пароль</label>
      <input id="password" type="password" {...register('password')} />
      <button type="submit">Войти</button>
      <div data-testid="err-email">{formState.errors.email?.message ?? ''}</div>
      <div data-testid="err-pass">{formState.errors.password?.message ?? ''}</div>
    </form>
  );
}

describe('debug', () => {
  it('change works', () => {
    const spy = vi.fn();
    render(<DebugForm />);
    const email = screen.getByLabelText('Email');
    const pass = screen.getByLabelText('Пароль');
    fireEvent.change(email, { target: { value: 'a@b.c' } });
    fireEvent.change(pass, { target: { value: 'password123' } });
    console.log('email value after change:', (email as HTMLInputElement).value);
    console.log('err-email:', screen.getByTestId('err-email').textContent);
    console.log('err-pass:', screen.getByTestId('err-pass').textContent);
    screen.getByRole('button', { name: 'Войти' }).click();
  });
});
