import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { QueryObserverResult } from '@tanstack/react-query';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { updateTrip } from '../api/trips';
import type { Trip, TripPayload } from '../lib/types';
import { TripEdit } from './TripEdit';

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>(
    '@tanstack/react-query',
  );
  return {
    ...actual,
    useQuery: vi.fn(),
    useMutation: vi.fn(),
  };
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>(
    'react-router-dom',
  );
  return {
    ...actual,
    useParams: vi.fn(),
    useNavigate: vi.fn(),
  };
});

vi.mock('../api/trips', () => ({
  getTrip: vi.fn(),
  updateTrip: vi.fn(),
}));

// Корректный паттерн: vi.mocked требует уже импортированный объект.
const mockUseQuery = vi.mocked(useQuery);
const mockUseMutation = vi.mocked(useMutation);
const mockUseParams = vi.mocked(useParams);
const mockUseNavigate = vi.mocked(useNavigate);
const mockUpdateTrip = vi.mocked(updateTrip);

// Полный объект Trip по контракту lib/types.ts.
const mockTrip: Trip = {
  id: '1',
  title: 'Trip 1',
  description: 'Desc 1',
  start_date: '2024-01-01',
  end_date: '2024-01-02',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

// Возврат useQuery типизируем через Partial<QueryObserverResult<Trip, unknown>>.
function queryResult(data: Trip | undefined, isLoading = false) {
  return {
    data,
    isLoading,
    isError: false,
    error: null,
  } satisfies Partial<QueryObserverResult<Trip, unknown>>;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseParams.mockReturnValue({ id: '1' });
  mockUseNavigate.mockReturnValue(vi.fn());
  mockUseQuery.mockReturnValue(queryResult(mockTrip));
  // Полный объект Trip (не {id, title}) по контракту types.ts.
  mockUpdateTrip.mockResolvedValue(mockTrip);
});

describe('TripEdit', () => {
  it('should prefill form with trip data', () => {
    render(<TripEdit />);
    expect(screen.getByLabelText(/Название поездки/i)).toHaveValue('Trip 1');
    expect(screen.getByLabelText('Описание')).toHaveValue('Desc 1');
  });

  it('should update trip and navigate on submit', async () => {
    const navigate = mockUseNavigate() as unknown as ReturnType<typeof vi.fn>;
    // mutateAsync эмулируем обёрткой над моком updateTrip: так проверяется
    // реальный вызов API с полным payload'ом.
    const mutateAsync = vi.fn().mockImplementation(
      async (p: TripPayload) => mockUpdateTrip('1', p),
    );
    mockUseMutation.mockReturnValue({
      mutateAsync,
      mutate: vi.fn(),
    });

    render(<TripEdit />);
    // Заполняем валидные даты: tripSchema требует \d{4}-\d{2}-\d{2} и end >= start.
    fireEvent.change(screen.getByLabelText('Дата начала'), {
      target: { value: '2024-01-01' },
    });
    fireEvent.change(screen.getByLabelText('Дата окончания'), {
      target: { value: '2024-01-02' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Сохранить/ }));

    await waitFor(() => {
      expect(mockUpdateTrip).toHaveBeenCalledWith(
        '1',
        expect.objectContaining({ title: 'Trip 1' }),
      );
    });
  });
});
