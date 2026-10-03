import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { TripsList } from '../pages/TripsList';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { createTrip } from '../api/trips';

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
}));

vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
}));

vi.mock('../api/trips', () => ({
  createTrip: vi.fn(),
}));

describe('TripsList', () => {
  it('should render a list of trips', async () => {
    const mockTrips = [
      { id: '1', title: 'Trip 1', description: 'Desc 1', start_date: '2024-01-01', end_date: '2024-01-02', created_at: '', updated_at: '' },
      { id: '2', title: 'Trip 2', description: 'Desc 2', start_date: '2024-01-03', end_date: '2024-01-04', created_at: '', updated_at: '' },
    ];
    vi.mocked(useQuery).mockReturnValue({
      data: mockTrips,
      isLoading: false,
      error: null,
    });

    render(<TripsList />);
    expect(screen.getByText('Trip 1')).toBeInTheDocument();
    expect(screen.getByText('Trip 2')).toBeInTheDocument();
  });

  it('should render placeholder when list is empty', async () => {
    vi.mocked(useQuery).mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    });

    render(<TripsList />);
    expect(screen.getByText('Пока нет поездок')).toBeInTheDocument();
  });

  it('should create a draft trip and navigate to edit when clicking + New Trip', async () => {
    const navigate = vi.mocked(useNavigate());
    vi.mocked(createTrip).mockResolvedValue({ id: 'new-id', title: 'Новая поездка' });

    render(<TripsList />);
    fireEvent.click(screen.getByText('+ Новая поездка'));

    await waitFor(() => {
      expect(createTrip).toHaveBeenCalled();
      expect(navigate).toHaveBeenCalledWith('/trips/new-id/edit');
    });
  });
});
