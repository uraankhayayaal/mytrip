import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { listTrips, createTrip } from '../api/trips';
import { TripCard } from '../components/TripCard';
import { ApiError } from '../api/errors';
import { todayISO } from '../lib/dates';

export function TripsList(): JSX.Element {
  return <div>TODO: TripsList</div>;
}
