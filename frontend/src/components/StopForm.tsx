import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { createStop, updateStop } from '../api/stops';
import { toApiError } from '../api/errors';
import type { Stop } from '../lib/types';

export interface StopFormProps {
  tripId: string;
  initial?: Stop; // если есть — режим редактирования
  onSaved: () => void;
  onCancel: () => void;
}

const stopSchema = z.object({
  title: z.string().min(1, 'Название обязательно'),
  description: z.string().default(''),
  lat: z.coerce
    .number({ invalid_type_error: 'Широта должна быть числом' })
    .min(-90, 'Широта должна быть в диапазоне -90..90')
    .max(90, 'Широта должна быть в диапазоне -90..90'),
  lng: z.coerce
    .number({ invalid_type_error: 'Долгота должна быть числом' })
    .min(-180, 'Долгота должна быть в диапазоне -180..180')
    .max(180, 'Долгота должна быть в диапазоне -180..180'),
  visit_date: z.string().min(1, 'Дата посещения обязательна'),
  order: z.coerce
    .number({ invalid_type_error: 'Порядок должен быть числом' })
    .min(0, 'Порядок должен быть >= 0'),
});

type StopInput = z.infer<typeof stopSchema>;

export function StopForm({ tripId, initial, onSaved, onCancel }: StopFormProps): JSX.Element {
  const [apiError, setApiError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<StopInput>({
    resolver: zodResolver(stopSchema),
    defaultValues: initial
      ? {
          title: initial.title,
          description: initial.description,
          lat: initial.location.lat,
          lng: initial.location.lng,
          visit_date: initial.visit_date,
          order: initial.order,
        }
      : undefined,
  });

  const onSubmit = handleSubmit(async (values) => {
    setApiError(null);
    const payload = {
      title: values.title,
      description: values.description,
      location: { lat: Number(values.lat), lng: Number(values.lng) },
      visit_date: values.visit_date,
      order: Number(values.order),
    };
    try {
      if (initial) {
        await updateStop(initial.id, payload);
      } else {
        await createStop(tripId, payload);
      }
      onSaved();
    } catch (err) {
      const e = toApiError(err);
      if (e.code === 'validation') {
        setApiError(e.message);
      } else {
        throw err;
      }
    }
  });

  return (
    <form onSubmit={onSubmit} className="space-y-3" noValidate>
      {apiError && (
        <div role="alert" className="rounded border border-red-300 bg-red-50 p-2 text-sm text-red-700">
          {apiError}
        </div>
      )}
      <div>
        <label htmlFor="stop-title" className="block text-sm font-medium">
          Название
        </label>
        <input
          id="stop-title"
          type="text"
          {...register('title')}
          className="mt-1 block w-full rounded border border-gray-300 px-2 py-1"
        />
        {errors.title && <p className="text-sm text-red-600">{errors.title.message}</p>}
      </div>
      <div>
        <label htmlFor="stop-description" className="block text-sm font-medium">
          Описание
        </label>
        <textarea
          id="stop-description"
          {...register('description')}
          className="mt-1 block w-full rounded border border-gray-300 px-2 py-1"
        />
        {errors.description && <p className="text-sm text-red-600">{errors.description.message}</p>}
      </div>
      <div>
        <label htmlFor="stop-lat" className="block text-sm font-medium">
          Широта
        </label>
        <input
          id="stop-lat"
          type="number"
          step="0.000001"
          {...register('lat')}
          className="mt-1 block w-full rounded border border-gray-300 px-2 py-1"
        />
        {errors.lat && <p className="text-sm text-red-600">{errors.lat.message}</p>}
      </div>
      <div>
        <label htmlFor="stop-lng" className="block text-sm font-medium">
          Долгота
        </label>
        <input
          id="stop-lng"
          type="number"
          step="0.000001"
          {...register('lng')}
          className="mt-1 block w-full rounded border border-gray-300 px-2 py-1"
        />
        {errors.lng && <p className="text-sm text-red-600">{errors.lng.message}</p>}
      </div>
      <div>
        <label htmlFor="stop-visit-date" className="block text-sm font-medium">
          Дата посещения
        </label>
        <input
          id="stop-visit-date"
          type="date"
          {...register('visit_date')}
          className="mt-1 block w-full rounded border border-gray-300 px-2 py-1"
        />
        {errors.visit_date && <p className="text-sm text-red-600">{errors.visit_date.message}</p>}
      </div>
      <div>
        <label htmlFor="stop-order" className="block text-sm font-medium">
          Порядок
        </label>
        <input
          id="stop-order"
          type="number"
          min={0}
          {...register('order')}
          className="mt-1 block w-full rounded border border-gray-300 px-2 py-1"
        />
        {errors.order && <p className="text-sm text-red-600">{errors.order.message}</p>}
      </div>
      <div className="flex gap-2">
        <button type="submit" disabled={isSubmitting} className="rounded bg-blue-600 px-3 py-1 text-white">
          Сохранить
        </button>
        <button type="button" onClick={onCancel} className="rounded border border-gray-300 px-3 py-1">
          Отмена
        </button>
      </div>
    </form>
  );
}
