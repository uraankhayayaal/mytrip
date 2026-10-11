import { useQuery } from '@tanstack/react-query';
import { Navigate, Outlet } from 'react-router-dom';
import { me } from '../api/auth';
import { getAccessToken } from '../api/client';
import { ApiError } from '../api/errors';
import { Layout } from './Layout';

export function ProtectedLayout(): JSX.Element {
  const { isPending, error } = useQuery({
    queryKey: ['me'],
    queryFn: me,
    enabled: getAccessToken() !== null,
  });

  if (isPending) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" />
      </div>
    );
  }

  // 401 / ApiError code 'unauthorized' -> редирект на /login.
  if (error instanceof ApiError && error.code === 'unauthorized') {
    return <Navigate to="/login" replace />;
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  );
}
