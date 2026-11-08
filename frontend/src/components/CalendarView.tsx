import { Calendar, dateFnsLocalizer } from 'react-big-calendar';
import format from 'date-fns/format';
import parse from 'date-fns/parse';
import locale from 'date-fns/locale';
import { CalendarEvent } from '../lib/calendar';
import 'react-big-calendar/lib/css/react-big-calendar.css';

const localizer = dateFnsLocalizer({
  format,
  parse,
  locale,
});

export interface CalendarViewProps {
  events: CalendarEvent[];
  onSelectEvent: (e: CalendarEvent) => void;
}

const EventComponent = ({ event }: { event: CalendarEvent }) => (
  <span className="px-1 truncate">{event.title}</span>
);

export function CalendarView({ events, onSelectEvent }: CalendarViewProps): JSX.Element {
  return (
    <div className="h-[600px]">
      <Calendar
        localizer={localizer}
        events={events.map(e => ({ ...e, start: e.start, end: e.end }))}
        views={['month', 'agenda']}
        selectable={false}
        popup
        eventPropGetter={(event) => {
          const calendarEvent = events.find(e => e.id === event.id);
          return {
            className: calendarEvent?.kind === 'trip' 
              ? 'bg-blue-100 border-blue-500 text-blue-700' 
              : 'bg-green-100 border-green-500 text-green-700',
            style: { cursor: 'pointer' },
          };
        }}
        onSelectEvent={(event) => {
          const found = events.find(e => e.id === event.id);
          if (found) onSelectEvent(found);
        }}
        components={{
          event: EventComponent,
        }}
      />
    </div>
  );
}
