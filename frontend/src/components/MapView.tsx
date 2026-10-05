import L from 'leaflet';
import { MapContainer, TileLayer, Marker, Polyline } from 'react-leaflet';
import type { Stop, Location } from '../lib/types';

export interface MapViewProps {
  stops: Stop[];
  onMoveStop?: (stopId: string, location: Location) => void; // drag-маркера
  selectedStopId?: string | null;
  onSelectStop?: (stopId: string) => void; // клик по маркеру
}

export function MapView({ stops, onMoveStop, selectedStopId, onSelectStop }: MapViewProps): JSX.Element {
  const stopsSorted = [...stops].sort((a, b) => a.order - b.order);

  return (
    <MapContainer center={[55.75, 37.6]} zoom={4} className="h-96 w-full">
      <TileLayer attribution='&copy; OpenStreetMap contributors' url="https://tile.openstreetmap.org/{z}/{x}/{y}.png" />
      {stopsSorted.map((stop) => {
        const isSelected = stop.id === selectedStopId;
        const icon = L.divIcon({
          html: `<div class="flex h-6 w-6 items-center justify-center rounded-full ${
            isSelected ? 'bg-red-600' : 'bg-blue-600'
          } text-white text-xs">${stop.order + 1}</div>`,
          className: '',
        });
        return (
          <Marker
            key={stop.id}
            position={[stop.location.lat, stop.location.lng]}
            icon={icon}
            draggable={!!onMoveStop}
            eventHandlers={{
              click: () => onSelectStop?.(stop.id),
              dragend: (e) => {
                const ll = e.target.getLatLng();
                onMoveStop?.(stop.id, { lat: ll.lat, lng: ll.lng });
              },
            }}
          />
        );
      })}
      {stopsSorted.length > 0 && (
        <Polyline
          positions={stopsSorted.map((s) => [s.location.lat, s.location.lng] as [number, number])}
          pathOptions={{ color: '#2563eb', weight: 2 }}
        />
      )}
    </MapContainer>
  );
}
