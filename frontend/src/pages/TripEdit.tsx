import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getTrip, updateTrip } from '../api/trips';
import { TripForm } from '../components/TripForm';
import { ApiError } from '../api/errors';

export function TripEdit(): JSX.Element {
  return <div>TODO: TripEdit</div>;
}
