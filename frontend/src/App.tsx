import { Routes, Route, Navigate } from 'react-router-dom'

// Lazy-loaded components
const Login = React.lazy(() => import('./pages/Login'))
const Register = React.lazy(() => import('./pages/Register'))
const TripsList = React.lazy(() => import('./pages/TripsList'))
const TripDetail = React.lazy(() => import('./pages/TripDetail'))
const TripEdit = React.lazy(() => import('./pages/TripEdit'))
const CalendarPage = React.lazy(() => import('./pages/CalendarPage'))

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login/>}/>
      <Route path="/register" element={<Register/>}/>
      <Route element={<ProtectedLayout/>}>
        <Route path="/" element={<TripsList/>}/>
        <Route path="/trips/:id" element={<TripDetail/>}/>
        <Route path="/trips/:id/edit" element={<TripEdit/>}/>
        <Route path="/calendar" element={<CalendarPage/>}/>
      </Route>
      <Route path="*" element={<Navigate to="/" replace/>}/>
    </Routes>
  )
}

export default App