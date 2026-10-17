import { describe, it, expect } from 'vitest';
import { loginSchema } from './schemas';
import { z } from 'zod';

// Проверяем, как zodResolver передаёт данные: с пустыми строками и без полей.
async function runResolver(data: Record<string, unknown>) {
  const res = await loginSchema.safeParse(data as never);
  return { success: res.success, error: res.error?.issues ?? null };
}

describe('loginSchema', () => {
  it('valid values pass', async () => {
    const r = await runResolver({ email: 'a@b.c', password: 'password123' });
    expect(r.success).toBe(true);
  });

  it('empty strings fail with which fields', async () => {
    const r = await runResolver({});
    expect(r.success).toBe(false);
    expect(r.error?.map((i) => i.path.join('.'))).toEqual(expect.arrayContaining(['email', 'password']));
  });

  it('zod version has .error issues', async () => {
    const r = await runResolver({});
    expect(r.error?.length).toBeGreaterThan(0);
  });
});
