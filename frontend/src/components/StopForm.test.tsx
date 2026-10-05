import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { StopForm } from './StopForm';
import { createStop, updateStop } from '../api/stops';
import type { Stop } from '../lib/types';

vi.mock('../api/stops', () => ({
  createStop: vi.fn(),
  updateStop: vi.fn(),
}));

const mockCreate = vi.mocked(createStop);
const mockUpdate = vi.mocked(updateStop);

const stop: Stop = {
  id: 'stop-1',
  trip_id: 'trip-1',
  title: 'Старый стоп',
  description: 'Описание',
  location: { lat: 55.75, lng: 37.61 },
  visit_date: '2025-06-10',
  order: 1,
  created_at: '2025-06-01T00:00:00Z',
  updated_at: '2025-06-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockCreate.mockResolvedValue(stop);
  mockUpdate.mockResolvedValue(stop);
});

describe('StopForm', () => {
  it('вызывает createStop при создании нового стопа', async () => {
    const onSaved = vi.fn();
    render(<StopForm tripId="trip-1" onSaved={onSaved} onCancel={vi.fn()} />);

    fireEvent.change(screen.getByLabelText(/Название/), { target: { value: 'Новый стоп' } });
    fireEvent.change(screen.getByLabelText(/Широта/), { target: { value: '55.75' } });
    fireEvent.change(screen.getByLabelText(/Долгота/), { target: { value: '37.61' } });
    fireEvent.change(screen.getByLabelText(/Дата посещения/), { target: { value: '2025-06-10' } });
    fireEvent.change(screen.getByLabelText(/Порядок/), { target: { value: '0' } });
    fireEvent.click(screen.getByRole('button', { name: /Сохранить/ }));

    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate).toHaveBeenCalledWith('trip-1', expect.objectContaining({
      title: 'Новый стоп',
      location: { lat: 55.75, lng: 37.61 },
    }));
    expect(onSaved).toHaveBeenCalled();
  });

  it('вызывает updateStop при редактировании', async () => {
    render(<StopForm tripId="trip-1" initial={stop} onSaved={vi.fn()} onCancel={vi.fn()} />);

    fireEvent.change(screen.getByLabelText(/Название/), { target: { value: 'Обновлённый' } });
    fireEvent.click(screen.getByRole('button', { name: /Сохранить/ }));

    await waitFor(() => expect(mockUpdate).toHaveBeenCalledTimes(1));
    expect(mockUpdate).toHaveBeenCalledWith('stop-1', expect.objectContaining({
      title: 'Обновлённый',
    }));
    expect(mockCreate).not.toHaveBeenCalled();
  });

  it('показывает ошибки валидации lat/lng', async () => {
    render(<StopForm tripId="trip-1" onSaved={vi.fn()} onCancel={vi.fn()} />);

    fireEvent.change(screen.getByLabelText(/Широта/), { target: { value: '999' } });
    fireEvent.change(screen.getByLabelText(/Долгота/), { target: { value: '-999' } });
    fireEvent.click(screen.getByRole('button', { name: /Сохранить/ }));

    await waitFor(() => {
      expect(screen.getByText(/Широта должна быть в диапазоне/)).toBeInTheDocument();
      expect(screen.getByText(/Долгота должна быть в диапазоне/)).toBeInTheDocument();
    });
    expect(mockCreate).not.toHaveBeenCalled();
  });
});
