import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getTrip, updateTrip } from '../api/trips';
import { TripForm } from '../components/TripForm';
import { ApiError } from '../api/errors';

export function TripEdit(): JSX.Element {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: trip, isLoading, error } = useQuery({
    queryKey: ['trip', id],
    queryFn: () => getTrip(id!),
  });

  const mutation = useMutation({
    mutationFn: (p: any) => updateTrip(id!, p),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['trip', id] });
      queryClient.invalidateQueries({ queryKey: ['trips'] });
      navigate(`/trips/${id}`);
    },
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

  if (!trip) return <div>Поездка не найдена</div>;

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-6">Редактирование поездки</h1>
      <TripForm 
        initial={trip} 
        onSubmit={async (p) => {
          try {
            await mutation.mutateAsync(p);
          } catch (e) {
            // Ошибка обрабатывается в mutation.onError или через общий баннер
            throw e;
          }
        }} 
      />
    </div>
  );
}
