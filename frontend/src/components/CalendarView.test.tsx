import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';
import { CalendarView } from './CalendarView';
import { CalendarEvent } from '../lib/calendar';

// Мокаем react-big-calendar
vi.mock('react-big-calendar', () => ({
  Calendar: ({ events, onSelectEvent }: any) => (
    <div data-testid="calendar" className="h-[600px]">
      {events.map((event: CalendarEvent) => (
        <div 
          key={event.id} 
          data-testid="event"
          onClick={() => onSelectEvent(event)}
        >
          {event.title}
        </div>
      ))}
    </div>
  ),
  dateFnsLocalizer: vi.fn(),
}));

describe('CalendarView', () => {
  const mockEvents: CalendarEvent[] = [
    {
      id: 'trip-1',
      title: 'Trip 1',
      start: new Date('2023-01-01'),
      end: new Date('2023-01-05'),
      tripId: '1',
      kind: 'trip',
    },
    {
      id: 'stop-1',
      title: 'Stop 1',
      start: new Date('2023-01-02'),
      end: new Date('2023-01-02'),
      tripId: '1',
      kind: 'stop',
    },
  ];

  const mockOnSelectEvent = vi.fn();

  it('should render events', () => {
    const { getByTestId } = render(
      <CalendarView events={mockEvents} onSelectEvent={mockOnSelectEvent} />
    );

    const calendar = getByTestId('calendar');
    expect(calendar).toBeInTheDocument();
    
    const events = calendar.querySelectorAll('[data-testid="event"]');
    expect(events).toHaveLength(2);
  });

  it('should call onSelectEvent when event is selected', () => {
    const { getByTestId } = render(
      <CalendarView events={mockEvents} onSelectEvent={mockOnSelectEvent} />
    );

    const eventElement = getByTestId('event');
    eventElement.click();
    
    expect(mockOnSelectEvent).toHaveBeenCalled();
  });
});