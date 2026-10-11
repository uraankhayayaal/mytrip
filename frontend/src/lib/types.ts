// ЕДИНЫЙ источник клиентских типов. Зеркалит контракт ARCH-03 (бэкенд): менять сигнатуры нельзя.

export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface Location {
  lat: number;
  lng: number;
}

export interface Trip {
  id: string;
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  cover_photo_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Stop {
  id: string;
  trip_id: string;
  title: string;
  description: string;
  location: Location;
  visit_date: string;
  order: number;
  created_at: string;
  updated_at: string;
}

export interface Photo {
  id: string;
  trip_id: string;
  url: string;
  created_at: string;
}

export interface PhotoUrlResponse {
  url: string;
  expires_at: string;
}

export interface ApiErrorBody {
  error: {
    code: 'validation' | 'not_found' | 'unauthorized' | 'conflict' | 'internal';
    message: string;
    details?: unknown;
  };
}

export interface LoginResponse {
  access_token: string;
  user: User;
}

export interface RegisterPayload {
  email: string;
  password: string;
  name: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface TripPayload {
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  cover_photo_url?: string;
}

export interface StopPayload {
  title: string;
  description: string;
  location: Location;
  visit_date: string;
  order: number;
}
