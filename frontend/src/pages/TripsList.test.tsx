import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useQuery } from '@tanstack/react-query';
import type { QueryObserverResult } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import type { NavigateFunction } from 'react-router-dom';
import { createTrip } from '../api/trips';
import type { Trip } from '../lib/types';
import { TripsList } from './TripsList';

// Тестируемый компонент использует реальный QueryClientProvider (см. App.tsx).
// useQueryClient() в этом провайдере возвращает реальный клиент — его мокать не нужно.
vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>(
    '@tanstack/react-query',
  );
  return {
    ...actual,
    useQuery: vi.fn(),
  };
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: vi.fn(),
  };
});

vi.mock('../api/trips', () => ({
  createTrip: vi.fn(),
  listTrips: vi.fn(),
}));

const mockUseQuery = vi.mocked(useQuery);
const mockUseNavigate = vi.mocked(useNavigate);
const mockCreateTrip = vi.mocked(createTrip);

// Полный объект Trip по контракту frontend/src/lib/types.ts.
function makeTrip(id: string, title: string, description: string): Trip {
  return {
    id,
    title,
    description,
    start_date: '2024-01-01',
    end_date: '2024-01-02',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };
}

// Возврат useQuery: типизируем через Partial<QueryObserverResult<Trip[], unknown>> —
// в mockReturnValue не нужен полный объект (остальные поля компоненты не читают).
function queryResult(data: Trip[] | undefined): Partial<QueryObserverResult<Trip[], unknown>> {
  return {
    data,
    isLoading: false,
    isError: false,
    error: null,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockUseNavigate.mockReturnValue(vi.fn() as unknown as NavigateFunction);
  mockUseQuery.mockReturnValue(queryResult(undefined));
});

describe('TripsList', () => {
  it('should render a list of trips', () => {
    const mockTrips: Trip[] = [makeTrip('1', 'Trip 1', 'Desc 1'), makeTrip('2', 'Trip 2', 'Desc 2')];
    mockUseQuery.mockReturnValue(queryResult(mockTrips));

    render(<TripsList />);
    expect(screen.getByText('Trip 1')).toBeInTheDocument();
    expect(screen.getByText('Trip 2')).toBeInTheDocument();
  });

  it('should render placeholder when list is empty', () => {
    mockUseQuery.mockReturnValue(queryResult([]));

    render(<TripsList />);
    expect(screen.getByText('Пока нет поездок')).toBeInTheDocument();
  });

  it('should render loading spinner while query is loading', () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    });

    const { container } = render(<TripsList />);
    expect(container.querySelector('.animate-spin')).not.toBeNull();
  });

  it('should create a draft trip and navigate to edit when clicking + New Trip', async () => {
    mockUseQuery.mockReturnValue(queryResult([]));
    const navigate = mockUseNavigate() as unknown as ReturnType<typeof vi.fn>;
    const newTrip: Trip = makeTrip('new-id', 'Новая поездка', '');
    mockCreateTrip.mockResolvedValue(newTrip);

    render(<TripsList />);
    fireEvent.click(screen.getByText('+ Новая поездка'));

    await waitFor(() => {
      expect(mockCreateTrip).toHaveBeenCalledWith({
        title: 'Новая поездка',
        description: '',
        start_date: expect.any(String),
        end_date: expect.any(String),
      });
    });
    expect(navigate).toHaveBeenCalledWith('/trips/new-id/edit');
  });

  it('shows an alert when createTrip rejects', async () => {
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => undefined);
    mockUseQuery.mockReturnValue(queryResult([]));
    mockCreateTrip.mockRejectedValue(new Error('boom'));

    render(<TripsList />);
    fireEvent.click(screen.getByText('+ Новая поездка'));

    await waitFor(() => {
      expect(alertSpy).toHaveBeenCalledWith('Не удалось создать черновик поездки');
    });
    alertSpy.mockRestore();
  });
});
