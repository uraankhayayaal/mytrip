import { Routes, Route, Navigate } from 'react-router-dom';
import { ProtectedLayout } from './components/ProtectedLayout';
import { CalendarPage } from './pages/CalendarPage';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { TripDetail } from './pages/TripDetail';
import { TripEdit } from './pages/TripEdit';
import { TripsList } from './pages/TripsList';

export function App(): JSX.Element {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route element={<ProtectedLayout />}>
        <Route path="/" element={<TripsList />} />
        <Route path="/trips/:id" element={<TripDetail />} />
        <Route path="/trips/:id/edit" element={<TripEdit />} />
        <Route path="/calendar" element={<CalendarPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
