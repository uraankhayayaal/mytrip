import React from 'react'
import { Outlet, Navigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { me } from '../api/auth'
import { getAccessToken } from '../api/client'

export function ProtectedLayout(): JSX.Element {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['me'],
    queryFn: me,
    enabled: getAccessToken() !== null
  })

  // Loading state
  if (isLoading) {
    return <div className="flex justify-center items-center h-screen">Загрузка...</div>
  }

  // Error handling
  if (isError) {
    // Check if it's an unauthorized error
    if (error && typeof error === 'object' && 'response' in error && error.response && 
        (error.response as any).status === 401) {
      return <Navigate to="/login" replace />
    }
    // For other errors, we might want to show an error message
    return <Navigate to="/login" replace />
  }

  // Render layout with children
  return (
    <div>
      <Layout>
        <Outlet />
      </Layout>
    </div>
  )
}

// We need to import Layout here
import { Layout } from './Layout'