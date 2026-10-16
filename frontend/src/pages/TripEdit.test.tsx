import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import type { QueryObserverResult, UseMutationResult } from '@tanstack/react-query';
import { TripEdit } from '../pages/TripEdit';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { getTrip, updateTrip } from '../api/trips';
import type { Trip, TripPayload } from '../lib/types';

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(),
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
  })),
}));

vi.mock('react-router-dom', () => ({
  useParams: vi.fn(),
  useNavigate: vi.fn(),
}));

vi.mock('../api/trips', () => ({
  getTrip: vi.fn(),
  updateTrip: vi.fn(),
}));

const mockUseQuery = vi.mocked(useQuery) as Mock;
const mockUseMutation = vi.mocked(useMutation) as Mock;
const mockUseParams = vi.mocked(useParams) as Mock;
const mockUseNavigate = vi.mocked(useNavigate) as Mock;
const mockUpdateTrip = vi.mocked(updateTrip) as Mock;

const mockTrip: Trip = {
  id: '1',
  title: 'Trip 1',
  description: 'Desc 1',
  start_date: '2024-01-01',
  end_date: '2024-01-02',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseParams.mockReturnValue({ id: '1' });
  mockUseNavigate.mockReturnValue(vi.fn());
  mockUseQuery.mockReturnValue({
    data: mockTrip,
    isLoading: false,
    isError: false,
    error: null,
  } satisfies Partial<QueryObserverResult<Trip, unknown>>);
  mockUpdateTrip.mockResolvedValue(mockTrip);
});

describe('TripEdit', () => {
  it('should prefill form with trip data', () => {
    render(<TripEdit />);
    expect(screen.getByLabelText(/Название поездки/i)).toHaveValue('Trip 1');
    expect(screen.getByLabelText(/Описание/i)).toHaveValue('Desc 1');
  });

  it('should update trip and navigate on submit', async () => {
    const navigate = mockUseNavigate() as Mock;
    const mutateAsync = vi.fn().mockImplementation(async (p: TripPayload) => {
      const result = await mockUpdateTrip('1', p);
      // Имитируем onSuccess: invalidateQueries + navigate
      const qc = (useQueryClient as unknown as Mock)() as { invalidateQueries: Mock };
      qc.invalidateQueries();
      navigate(`/trips/1`);
      return result;
    });
    mockUseMutation.mockReturnValue({
      mutateAsync,
    } satisfies Partial<UseMutationResult<Trip, unknown, TripPayload, unknown>>);

    render(<TripEdit />);
    // Заполняем поля формы значениями mockTrip
    const titleInput = screen.getByLabelText(/Название поездки/i) as HTMLInputElement;
    titleInput.value = 'Trip 1';
    const descInput = screen.getByLabelText(/Описание/i) as HTMLTextAreaElement;
    descInput.value = 'Desc 1';

    const submitBtn = screen.getByText('Сохранить');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockUpdateTrip).toHaveBeenCalledWith(
        '1',
        expect.objectContaining({ title: 'Trip 1' }),
      );
      expect(navigate).toHaveBeenCalledWith('/trips/1');
    });
  });
});
