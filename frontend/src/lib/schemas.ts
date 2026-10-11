import { z } from 'zod';

export const emailSchema = z.string().email();
export const passwordSchema = z.string().min(8, 'Минимум 8 символов');
export const nameSchema = z.string().min(1).max(100);

export const loginSchema = z.object({
  email: emailSchema,
  password: passwordSchema,
});

export const registerSchema = z.object({
  email: emailSchema,
  password: passwordSchema,
  name: nameSchema,
});

export const tripSchema = z
  .object({
    title: z.string().min(1).max(200),
    description: z.string().max(5000).default(''),
    start_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
    end_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
    cover_photo_url: z.string().url().optional(),
  })
  .refine((d) => d.end_date >= d.start_date, {
    message: 'end_date >= start_date',
    path: ['end_date'],
  });

export const stopSchema = z.object({
  title: z.string().min(1).max(200),
  description: z.string().max(5000).default(''),
  location: z.object({
    lat: z.number().min(-90).max(90),
    lng: z.number().min(-180).max(180),
  }),
  visit_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  order: z.number().int().min(0),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type TripInput = z.infer<typeof tripSchema>;
export type StopInput = z.infer<typeof stopSchema>;
