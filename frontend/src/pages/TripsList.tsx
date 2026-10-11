import React from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { listTrips, createTrip } from '../api/trips';
import { TripCard } from '../components/TripCard';
import { ApiError } from '../api/errors';
import { todayISO } from '../lib/dates';

export function TripsList(): JSX.Element {
  const navigate = useNavigate();
  const { data: trips, isLoading, error } = useQuery({
    queryKey: ['trips'],
    queryFn: listTrips,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-10">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    const apiError = error as ApiError;
    return (
      <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
        <span className="font-bold">Ошибка: </span>
        {apiError.message}
      </div>
    );
  }

  const handleCreateNew = async () => {
    try {
      const res = await createTrip({
        title: 'Новая поездка',
        description: '',
        start_date: todayISO(),
        end_date: todayISO(),
      });
      navigate(`/trips/${res.id}/edit`);
    } catch (err) {
      alert('Не удалось создать черновик поездки');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Мои поездки</h1>
        <button 
          onClick={handleCreateNew}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          + Новая поездка
        </button>
      </div>

      {trips && trips.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {trips.map(trip => (
            <TripCard key={trip.id} trip={trip} />
          ))}
        </div>
      ) : (
        <div className="text-center py-20 text-gray-500">
          Пока нет поездок
        </div>
      )}
    </div>
  );
}
