import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { TripEdit } from '../pages/TripEdit';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { getTrip, updateTrip } from '../api/trips';

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

describe('TripEdit', () => {
  const mockTrip = {
    id: '1',
    title: 'Trip 1',
    description: 'Desc 1',
    start_date: '2024-01-01',
    end_date: '2024-01-02',
    created_at: '',
    updated_at: '',
  };

  it('should prefill form with trip data', async () => {
    vi.mocked(useParams).mockReturnValue({ id: '1' });
    vi.mocked(useQuery).mockReturnValue({
      data: mockTrip,
      isLoading: false,
      error: null,
    });
    vi.mocked(useMutation).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({}),
    });

    render(<TripEdit />);
    expect(screen.getByLabelText(/Название поездки/i)).toHaveValue('Trip 1');
    expect(screen.getByLabelText(/Описание/i)).toHaveValue('Desc 1');
  });

  it('should update trip and navigate on submit', async () => {
    const navigate = vi.mocked(useNavigate);
    vi.mocked(useParams).mockReturnValue({ id: '1' });
    vi.mocked(useQuery).mockReturnValue({
      data: mockTrip,
      isLoading: false,
      error: null,
    });
    const mutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useMutation).mockReturnValue({
      mutateAsync,
    });

    render(<TripEdit />);
    
    // Find the submit button
    const submitBtn = screen.getByText('Сохранить');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalled();
      expect(navigate).toHaveBeenCalledWith('/trips/1');
    });
  });
});
