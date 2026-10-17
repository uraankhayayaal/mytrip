import { describe, it, expect } from 'vitest';
import { loginSchema } from './schemas';

describe('loginSchema', () => {
  it('valid values pass', async () => {
    const r = await loginSchema.safeParseAsync({
      email: 'a@b.c',
      password: 'password123',
    });
    expect(r.success).toBe(true);
  });

  it('empty payload fails with issues for email and password', async () => {
    const r = await loginSchema.safeParseAsync({});
    expect(r.success).toBe(false);
    if (!r.success) {
      const paths = r.error.issues.map((i) => i.path.join('.'));
      expect(paths).toEqual(expect.arrayContaining(['email', 'password']));
    }
  });

  it('short password fails with min-8 message', async () => {
    const r = await loginSchema.safeParseAsync({
      email: 'a@b.c',
      password: 'short',
    });
    expect(r.success).toBe(false);
    if (!r.success) {
      const pw = r.error.issues.find((i) => i.path[0] === 'password');
      expect(pw?.message).toBe('Минимум 8 символов');
    }
  });
});
