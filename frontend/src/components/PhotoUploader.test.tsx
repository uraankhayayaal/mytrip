import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { PhotoUploader } from './PhotoUploader';
import { uploadPhoto, deletePhoto } from '../api/photos';
import type { Photo } from '../lib/types';

vi.mock('../api/photos', () => ({
  uploadPhoto: vi.fn(),
  deletePhoto: vi.fn(),
}));

const mockUpload = vi.mocked(uploadPhoto);
const mockDelete = vi.mocked(deletePhoto);

const photo: Photo = {
  id: 'photo-1',
  trip_id: 'trip-1',
  url: 'https://cdn.example.com/photo-1.jpg',
  created_at: '2025-06-01T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUpload.mockResolvedValue(photo);
  mockDelete.mockResolvedValue(undefined);
  vi.spyOn(window, 'confirm').mockReturnValue(true);
});

describe('PhotoUploader', () => {
  it('вызывает uploadPhoto для каждого файла', async () => {
    render(<PhotoUploader tripId="trip-1" photos={[]} onPhotosChange={vi.fn()} />);

    const file1 = new File(['a'], 'a.jpg', { type: 'image/jpeg' });
    const file2 = new File(['b'], 'b.jpg', { type: 'image/jpeg' });
    const input = document.querySelector('input[type="file"]')!;
    fireEvent.change(input, { target: { files: [file1, file2] } });

    await waitFor(() => expect(mockUpload).toHaveBeenCalledTimes(2));
    expect(mockUpload).toHaveBeenNthCalledWith(1, 'trip-1', file1);
    expect(mockUpload).toHaveBeenNthCalledWith(2, 'trip-1', file2);
  });

  it('вызывает deletePhoto и onPhotosChange при удалении', async () => {
    const onPhotosChange = vi.fn();
    render(
      <PhotoUploader tripId="trip-1" photos={[photo]} onPhotosChange={onPhotosChange} />
    );

    fireEvent.click(screen.getByRole('button', { name: /Удалить/ }));

    await waitFor(() => expect(mockDelete).toHaveBeenCalledWith('photo-1'));
    expect(onPhotosChange).toHaveBeenCalled();
  });
});
