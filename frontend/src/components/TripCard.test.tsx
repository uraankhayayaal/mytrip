import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { TripCard } from '../components/TripCard';
import type { Trip } from '../lib/types';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { deleteTrip } from '../api/trips';

vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
}));

vi.mock('../api/trips', () => ({
  deleteTrip: vi.fn(),
}));

vi.mock('@tanstack/react-query', () => ({
  useQueryClient: vi.fn(() => ({
    invalidateQueries: vi.fn(),
  })),
}));

const mockUseNavigate = vi.mocked(useNavigate) as Mock;
const mockUseQueryClient = vi.mocked(useQueryClient) as Mock;
const mockDeleteTrip = vi.mocked(deleteTrip) as Mock;

const mockTrip: Trip = {
  id: '1',
  title: 'Поездка в горы',
  description: 'Крутой отдых в Альпах',
  start_date: '2024-01-01',
  end_date: '2024-01-10',
  cover_photo_url: 'http://example.com/photo.jpg',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  const invalidateQueries = vi.fn();
  mockUseQueryClient.mockReturnValue({ invalidateQueries });
  mockUseNavigate.mockReturnValue(vi.fn());
  mockDeleteTrip.mockResolvedValue(undefined);
});

describe('TripCard', () => {
  it('should render title and dates', () => {
    render(<TripCard trip={mockTrip} />);
    expect(screen.getByText('Поездка в горы')).toBeInTheDocument();
    expect(screen.getByText('01.01.2024 — 10.01.2024')).toBeInTheDocument();
  });

  it('should navigate to detail page on click', () => {
    const navigate = mockUseNavigate() as Mock;
    render(<TripCard trip={mockTrip} />);
    fireEvent.click(screen.getByText('Поездка в горы').closest('div')!);
    expect(navigate).toHaveBeenCalledWith('/trips/1');
  });

  it('should navigate to edit page when clicking Edit', () => {
    const navigate = mockUseNavigate() as Mock;
    render(<TripCard trip={mockTrip} />);
    fireEvent.click(screen.getByText('Изменить'));
    expect(navigate).toHaveBeenCalledWith('/trips/1/edit');
  });

  it('should call deleteTrip and invalidate queries when clicking Delete', async () => {
    const invalidateQueries = vi.mocked(mockUseQueryClient().invalidateQueries);
    vi.spyOn(window, 'confirm').mockImplementation(() => true);

    render(<TripCard trip={mockTrip} />);
    fireEvent.click(screen.getByText('Удалить'));

    await waitFor(() => {
      expect(mockDeleteTrip).toHaveBeenCalledWith('1');
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['trips'] });
    });
  });
});
