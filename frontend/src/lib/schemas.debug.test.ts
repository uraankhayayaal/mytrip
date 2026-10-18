import { describe, it, expect } from 'vitest';
import { loginSchema, tripSchema, stopSchema } from './schemas';

describe('schemas (debug)', () => {
  it('loginSchema: valid values pass', () => {
    const result = loginSchema.safeParse({ email: 'a@b.c', password: '12345678' });
    expect(result.success).toBe(true);
  });

  it('loginSchema: invalid email fails', () => {
    const result = loginSchema.safeParse({ email: 'not-an-email', password: '12345678' });
    expect(result.success).toBe(false);
  });

  it('tripSchema: valid trip passes', () => {
    const result = tripSchema.safeParse({
      title: 'My Trip',
      description: 'desc',
      start_date: '2024-01-01',
      end_date: '2024-01-10',
    });
    expect(result.success).toBe(true);
  });

  it('tripSchema: missing title fails', () => {
    const result = tripSchema.safeParse({
      description: 'desc',
      start_date: '2024-01-01',
      end_date: '2024-01-10',
    });
    expect(result.success).toBe(false);
  });

  it('stopSchema: valid stop passes', () => {
    const result = stopSchema.safeParse({
      title: 'Stop',
      description: 'd',
      location: { lat: 55.75, lng: 37.61 },
      visit_date: '2024-01-02',
      order: 1,
    });
    expect(result.success).toBe(true);
  });
});
