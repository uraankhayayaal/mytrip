import { useRef, useState } from 'react';
import { deletePhoto, uploadPhoto } from '../api/photos';
import type { Photo } from '../lib/types';

export interface PhotoUploaderProps {
  tripId: string;
  photos: Photo[];
  onPhotosChange: () => void;
}

export function PhotoUploader({ tripId, photos, onPhotosChange }: PhotoUploaderProps): JSX.Element {
  const [uploading, setUploading] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFiles = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? []);
    if (files.length === 0) return;
    setUploading(true);
    try {
      // Последовательно, не параллельно — чтобы не перегружать бэкенд.
      for (const file of files) {
        await uploadPhoto(tripId, file);
      }
      onPhotosChange();
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = '';
    }
  };

  const handleDelete = async (photo: Photo) => {
    if (!window.confirm('Удалить фото?')) return;
    await deletePhoto(photo.id);
    onPhotosChange();
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <input ref={inputRef} type="file" accept="image/*" multiple onChange={handleFiles} />
        {uploading && <span data-testid="photo-uploader-spinner" className="text-sm text-gray-500">Загрузка…</span>}
      </div>
      {photos.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {photos.map((photo) => (
            <div key={photo.id} className="relative">
              <img src={photo.url} alt={photo.id} className="h-24 w-full rounded object-cover" />
              <button
                type="button"
                onClick={() => handleDelete(photo)}
                className="absolute right-1 top-1 rounded bg-black/60 px-2 py-0.5 text-xs text-white"
              >
                Удалить
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
