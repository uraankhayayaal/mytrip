import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { me } from '../api/auth';
import { getAccessToken } from '../api/client';
import { Layout } from './Layout';
import { Outlet } from 'react-router-dom';

export function ProtectedLayout(): JSX.Element {
  const navigate = useNavigate();

  const { isLoading, error } = useQuery({
    queryKey: ['me'],
    queryFn: me,
    enabled: getAccessToken() !== null,
  });

  React.useEffect(() => {
    if (error && (error as any).response?.status === 401) {
      navigate('/login', { replace: true });
    }
  }, [error, navigate]);

  if (isLoading) {
    return <div>Загрузка...</div>;
  }

  if (error) {
    return <div>Ошибка: {(error as any).message}</div>;
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  );
}
