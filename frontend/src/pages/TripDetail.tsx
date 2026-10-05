import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getTrip } from '../api/trips';
import { createStop, deleteStop, listStops, updateStop } from '../api/stops';
import { listPhotos } from '../api/photos';
import { MapView } from '../components/MapView';
import { PhotoUploader } from '../components/PhotoUploader';
import { StopForm } from '../components/StopForm';
import { formatDate } from '../lib/dates';
import type { Location, Stop } from '../lib/types';

export function TripDetail(): JSX.Element {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [selectedStopId, setSelectedStopId] = useState<string | null>(null);
  const [editingStop, setEditingStop] = useState<Stop | null>(null);
  const [showStopForm, setShowStopForm] = useState(false);

  const tripQuery = useQuery({ queryKey: ['trip', id], queryFn: () => getTrip(id!), enabled: !!id });
  const stopsQuery = useQuery({ queryKey: ['stops', id], queryFn: () => listStops(id!), enabled: !!id });
  const photosQuery = useQuery({ queryKey: ['photos', id], queryFn: () => listPhotos(id!), enabled: !!id });

  const invalidateStops = () => queryClient.invalidateQueries({ queryKey: ['stops', id] });

  const updateStopMut = useMutation({
    mutationFn: ({ stopId, payload }: { stopId: string; payload: Parameters<typeof updateStop>[1] }) => updateStop(stopId, payload),
    onSuccess: invalidateStops,
  });
  const deleteStopMut = useMutation({
    mutationFn: (stopId: string) => deleteStop(stopId),
    onSuccess: invalidateStops,
  });
  const createStopMut = useMutation({
    mutationFn: (payload: Parameters<typeof createStop>[1]) => createStop(id!, payload),
    onSuccess: invalidateStops,
  });

  if (!id) return <div>Не найдено</div>;

  const trip = tripQuery.data;
  const stops = stopsQuery.data ?? [];
  const photos = photosQuery.data ?? [];
  const selectedStop = stops.find((s) => s.id === selectedStopId) ?? null;

  const onMoveStop = (stopId: string, location: Location) => {
    const s = stops.find((st) => st.id === stopId);
    if (!s) return;
    updateStopMut.mutate({
      stopId,
      payload: { title: s.title, description: s.description, location, visit_date: s.visit_date, order: s.order },
    });
  };

  const openEdit = (stop: Stop) => {
    setEditingStop(stop);
    setShowStopForm(true);
  };

  const openCreate = () => {
    setEditingStop(null);
    setShowStopForm(true);
  };

  const closeForm = () => {
    setShowStopForm(false);
    setEditingStop(null);
  };

  const handleDeleteStop = (stop: Stop) => {
    if (!window.confirm(`Удалить стоп «${stop.title}»?`)) return;
    deleteStopMut.mutate(stop.id);
  };

  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{trip?.title ?? '…'}</h1>
          {trip && (
            <p className="text-sm text-gray-500">
              {formatDate(trip.start_date)} — {formatDate(trip.end_date)}
            </p>
          )}
        </div>
        <button type="button" onClick={() => navigate(`/trips/${id}/edit`)} className="rounded border border-gray-300 px-3 py-1">
          Редактировать
        </button>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <div className="space-y-4">
          <MapView stops={stops} onMoveStop={onMoveStop} selectedStopId={selectedStopId} onSelectStop={setSelectedStopId} />
          {selectedStop && (
            <div className="rounded border border-gray-200 p-3 text-sm">
              <p className="font-semibold">{selectedStop.title}</p>
              <p>{selectedStop.description}</p>
              <p>Дата: {formatDate(selectedStop.visit_date)}</p>
              <p>Координаты: {selectedStop.location.lat} / {selectedStop.location.lng}</p>
            </div>
          )}
        </div>

        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">Стопы</h2>
              <button type="button" onClick={openCreate} className="rounded bg-blue-600 px-3 py-1 text-white">
                + Стоп
              </button>
            </div>
            {showStopForm && (
              <div className="rounded border border-gray-200 p-3">
                <StopForm
                  tripId={id}
                  initial={editingStop ?? undefined}
                  onSaved={() => {
                    createStopMut.reset();
                    closeForm();
                  }}
                  onCancel={closeForm}
                />
              </div>
            )}
            <ul className="space-y-2">
              {stops.map((stop) => (
                <li key={stop.id} className="rounded border border-gray-200 p-2">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">{stop.title}</p>
                      <p className="text-xs text-gray-500">
                        {formatDate(stop.visit_date)} · порядок {stop.order}
                      </p>
                    </div>
                    <div className="flex gap-1">
                      <button type="button" onClick={() => openEdit(stop)} className="rounded border border-gray-300 px-2 py-0.5 text-xs">
                        Изменить
                      </button>
                      <button type="button" onClick={() => handleDeleteStop(stop)} className="rounded border border-red-300 px-2 py-0.5 text-xs text-red-600">
                        Удалить
                      </button>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h2 className="mb-2 text-lg font-semibold">Фото</h2>
            <PhotoUploader tripId={id} photos={photos} onPhotosChange={() => queryClient.invalidateQueries({ queryKey: ['photos', id] })} />
          </div>

          <div className="flex gap-2">
            <button type="button" onClick={() => { window.location.href = `/api/v1/trips/${id}/export?format=pdf`; }} className="rounded border border-gray-300 px-3 py-1">
              Экспорт PDF
            </button>
            <button type="button" onClick={() => { window.location.href = `/api/v1/trips/${id}/export?format=csv`; }} className="rounded border border-gray-300 px-3 py-1">
              Экспорт CSV
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
