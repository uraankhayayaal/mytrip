import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { TripPayload, TripInput } from '../lib/types';
import { tripSchema } from '../lib/schemas';
import { useNavigate } from 'react-router-dom';

export interface TripFormProps {
  initial?: TripPayload;
  onSubmit: (p: TripPayload) => Promise<void>;
}

export function TripForm({ initial, onSubmit }: TripFormProps): JSX.Element {
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TripInput>({
    resolver: zodResolver(tripSchema),
    defaultValues: initial,
  });

  const handleSubmit = async (data: TripInput) => {
    try {
      await onSubmit(data as TripPayload);
    } catch (err) {
      // Ошибка перехватывается в вызывающем компоненте’s mutation.onSettled или через try-catch
      throw err;
    }
  };

  return (
    <form onSubmit={handleSubmit(handleSubmit)} className="space-y-4 max-w-2xl bg-white p-6 rounded-lg border shadow-sm">
      <div className="grid grid-cols-1 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Название поездки</label>
          <input 
            {...register('title')} 
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          />
          {errors.title && <p className="text-red-500 text-xs mt-1">{errors.title.message}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Описание</label>
          <textarea 
            {...register('description')} 
            rows={4}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          />
          {errors.description && <p className="text-red-500 text-xs mt-1">{errors.description.message}</p>}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Дата начала</label>
            <input 
              type="date" 
              {...register('start_date')} 
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            />
            {errors.start_date && <p className="text-red-500 text-xs mt-1">{errors.start_date.message}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">Дата окончания</label>
            <input 
              type="date" 
              {...register('end_date')} 
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            />
            {errors.end_date && <p className="text-red-500 text-xs mt-1">{errors.end_date.message}</p>}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">URL обложки (опционально)</label>
          <input 
            {...register('cover_photo_url')} 
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          />
          {errors.cover_photo_url && <p className="text-red-500 text-xs mt-1">{errors.cover_photo_url.message}</p>}
        </div>
      </div>

      <div className="flex justify-end gap-3 pt-4">
        <button 
          type="button" 
          onClick={() => navigate(-1)}
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
        >
          Отмена
        </button>
        <button 
          type="submit" 
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
        >
          Сохранить
        </button>
      </div>
    </form>
  );
}
