import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { Trip } from '../lib/types';
import { deleteTrip } from '../api/trips';
import { formatDate } from '../lib/dates';

export interface TripCardProps {
  trip: Trip;
}

export function TripCard({ trip }: TripCardProps): JSX.Element {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const handleCardClick = () => {
    navigate(`/trips/${trip.id}`);
  };

  const handleEditClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/trips/${trip.id}/edit`);
  };

  const handleDeleteClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm('Вы уверены, что хотите удалить эту поездку?')) {
      try {
        await deleteTrip(trip.id);
        await queryClient.invalidateQueries({ queryKey: ['trips'] });
      } catch (err) {
        alert('Ошибка при удалении поездки');
      }
    }
  };

  return (
    <div 
      onClick={handleCardClick}
      className="cursor-pointer bg-white border border-gray-200 rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow"
    >
      {trip.cover_photo_url && (
        <img 
          src={trip.cover_photo_url} 
          alt={trip.title} 
          className="w-full h-48 object-cover"
        />
      )}
      <div className="p-4">
        <h3 className="text-xl font-bold text-gray-900 truncate">{trip.title}</h3>
        <p className="text-sm text-gray-500 mb-2">
          {formatDate(trip.start_date)} — {formatDate(trip.end_date)}
        </p>
        <p className="text-gray-600 text-sm line-clamp-2 mb-4">
          {trip.description}
        </p>
        <div className="flex gap-2">
          <button 
            onClick={handleEditClick}
            className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Изменить
          </button>
          <button 
            onClick={handleDeleteClick}
            className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
          >
            Удалить
          </button>
        </div>
      </div>
    </div>
  );
}
