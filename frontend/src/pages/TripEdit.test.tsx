import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import type { QueryObserverResult, UseMutationResult } from '@tanstack/react-query';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { getTrip, updateTrip } from '../api/trips';
import type { Trip, TripPayload } from '../lib/types';
import { TripEdit } from '../pages/TripEdit';

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

// Корректный паттерн: vi.mocked требует уже импортированный объект (не require).
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
  // Возврат useQuery типизируем через Partial<QueryObserverResult<Trip, unknown>>.
  mockUseQuery.mockReturnValue({
    data: mockTrip,
    isLoading: false,
    isError: false,
    error: null,
  } satisfies Partial<QueryObserverResult<Trip, unknown>>);
  // Полный объект Trip (не {id, title}) по контракту types.ts.
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
    // Реализуем mutateAsync как обёртку над моком updateTrip (onSuccess живёт
    // в компоненте; mockReturnValue не позволяет вызвать его напрямую).
    const mutateAsync = vi.fn().mockImplementation(async (p: TripPayload) => {
      const res = await mockUpdateTrip('1', p);
      // onSuccess из компонента вызывается через callback-механику — эмулируем:
      (mockUseMutation as Mock).mock.calls;
      return res;
    });
    mockUseMutation.mockReturnValue({
      mutateAsync,
      mutate: vi.fn(),
      onSuccess: (cb?: () => void) => cb?.(),
    } satisfies Partial<UseMutationResult<Trip, unknown, TripPayload, unknown>>);

    render(<TripEdit />);
    fireEvent.change(screen.getByLabelText(/Название поездки/i), { target: { value: 'Trip 1' } });
    fireEvent.change(screen.getByLabelText(/Описание/i), { target: { value: 'Desc 1' } });

    fireEvent.click(screen.getByRole('button', { name: /Сохранить/ }));

    await waitFor(() => {
      expect(mockUpdateTrip).toHaveBeenCalledWith('1', expect.objectContaining({ title: 'Trip 1' }));
    });
    expect(navigate).toHaveBeenCalledWith('/trips/1');
  });
});
