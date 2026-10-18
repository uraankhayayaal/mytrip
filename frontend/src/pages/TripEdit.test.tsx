import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { updateTrip } from '../api/trips';
import type { Trip } from '../lib/types';
import { TripEdit } from './TripEdit';

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>(
    '@tanstack/react-query',
  );
  return {
    ...actual,
    useQuery: vi.fn(),
    useQueryClient: vi.fn(() => new QueryClient()),
  };
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: vi.fn(),
    useParams: vi.fn(() => ({ id: 'trip-1' })),
  };
});

vi.mock('../api/trips', () => ({
  getTrip: vi.fn(),
  updateTrip: vi.fn(),
  listTrips: vi.fn(),
  createTrip: vi.fn(),
  deleteTrip: vi.fn(),
}));

const mockUseQuery = vi.mocked(useQuery);
const mockUseNavigate = vi.mocked(useNavigate);
const mockUpdateTrip = vi.mocked(updateTrip);

const fullTrip: Trip = {
  id: 'trip-1',
  title: 'Original Title',
  description: 'Original desc',
  start_date: '2024-05-01',
  end_date: '2024-05-05',
  created_at: '2024-05-01T00:00:00Z',
  updated_at: '2024-05-01T00:00:00Z',
};

function renderWithProvider(ui: ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn());
  mockUseQuery.mockReturnValue({ data: fullTrip, isLoading: false });
});

describe('TripEdit', () => {
  it('should prefill form with trip data', () => {
    mockUseQuery.mockReturnValue({ data: fullTrip, isLoading: false });
    renderWithProvider(<TripEdit />);

    expect(screen.getByDisplayValue('Original Title')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Original desc')).toBeInTheDocument();
  });

  it('should update trip and navigate on submit', async () => {
    const updated: Trip = { ...fullTrip, title: 'Updated Title', updated_at: '2024-05-02T00:00:00Z' };
    mockUpdateTrip.mockResolvedValue(updated);
    const navigate = vi.fn();
    mockUseNavigate.mockReturnValue(navigate);
    mockUseQuery.mockReturnValue({ data: fullTrip, isLoading: false });

    renderWithProvider(<TripEdit />);

    const titleInput = screen.getByDisplayValue('Original Title') as HTMLInputElement;
    fireEvent.change(titleInput, { target: { value: 'Updated Title' } });
    fireEvent.click(screen.getByRole('button', { name: /сохранить/i }));

    await waitFor(() => {
      expect(mockUpdateTrip).toHaveBeenCalledWith('trip-1', {
        title: 'Updated Title',
        description: 'Original desc',
        start_date: '2024-05-01',
        end_date: '2024-05-05',
      });
    });
    expect(navigate).toHaveBeenCalledWith('/trips');
  });
});
