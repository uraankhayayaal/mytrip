import { useQuery, useQueries } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { CalendarView } from '../components/CalendarView';
import { listTrips } from '../api/trips';
import { listStops } from '../api/stops';
import { tripsToEvents, stopsToEvents } from '../lib/calendar';

export function CalendarPage(): JSX.Element {
  const navigate = useNavigate();

  const { data: trips = [], isLoading: tripsLoading, isError: tripsError } = useQuery({
    queryKey: ['trips'],
    queryFn: listTrips,
  });

  const stopsQueries = useQueries({
    queries: (trips ?? []).map(trip => ({
      queryKey: ['stops', trip.id],
      queryFn: () => listStops(trip.id),
      enabled: !!trips && trips.length > 0,
    })),
  });

  const allStops = stopsQueries.flatMap(q => q.data ?? []);
  const events = [...tripsToEvents(trips), ...stopsToEvents(allStops)];

  const handleSelectEvent = (e: any) => {
    navigate(`/trips/${e.tripId}`);
  };

  if (tripsLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (tripsError) {
    return (
      <div className="bg-red-100 text-red-700 p-4 rounded-md m-4">
        Ошибка при загрузке данных календаря.
      </div>
    );
  }

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Календарь поездок</h1>
      <CalendarView events={events} onSelectEvent={handleSelectEvent} />
    </div>
  );
}
