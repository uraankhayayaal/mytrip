# Фронтенд

## Обзор

Фронтенд MyTrip — React SPA, собранный через Vite. Отображает поездки, точки маршрута на карте, календарь, позволяет управлять данными через REST API бэкенда.

## Стек

| Технология | Версия | Назначение |
|---|---|---|
| React | 18 | UI-фреймворк |
| TypeScript | 5 | Типизация |
| Vite | 5 | Сборка + HMR |
| Tailwind CSS | 3 | Стили |
| React Router | 6 | Роутинг |
| React Query | 5 | Серверный стейт, кэширование |
| Axios | 1 | HTTP-клиент |
| Leaflet + react-leaflet | 4 | Карта с маркерами |
| React Big Calendar | 1 | Календарь поездок |
| React Hook Form + Zod | 7/3 | Валидация форм |
| Vitest + Testing Library | 1/14 | Unit-тесты (jsdom) |
| Playwright | 1.49 | E2E-тесты |
| date-fns | 4 | Работа с датами |

## Структура проекта

```
frontend/
├── src/
│   ├── main.tsx                    — точка входа, <App> с QueryClientProvider + BrowserRouter
│   ├── App.tsx                     — роутинг (Routes), публичные + защищённые страницы
│   ├── api/
│   │   ├── client.ts               — axios-инстанс, interceptors (JWT, refresh)
│   │   ├── errors.ts               — ApiError, toApiError()
│   │   ├── auth.ts                 — register(), login(), logout(), me()
│   │   ├── trips.ts                — list(), get(), create(), update(), delete()
│   │   ├── stops.ts                — list(), create(), update(), delete()
│   │   └── photos.ts               — list(), upload(), delete()
│   ├── components/
│   │   ├── Layout.tsx              — навигация, хедер, футер
│   │   ├── ProtectedLayout.tsx      — layout с проверкой авторизации (useQuery me)
│   │   ├── MapView.tsx             — Leaflet карта с маркерами точек маршрута
│   │   ├── CalendarView.tsx        — обёртка над react-big-calendar
│   │   ├── PhotoUploader.tsx       — загрузка фото через presigned URL
│   │   ├── StopForm.tsx            — форма добавления/редактирования точки
│   │   ├── TripCard.tsx            — карточка поездки в списке
│   │   └── TripForm.tsx            — форма создания/редактирования поездки
│   ├── pages/
│   │   ├── Login.tsx               — страница входа
│   │   ├── Register.tsx            — страница регистрации
│   │   ├── TripsList.tsx           — список поездок пользователя
│   │   ├── TripDetail.tsx          — детализация поездки + точки + карта
│   │   ├── TripEdit.tsx            — редактирование поездки
│   │   └── CalendarPage.tsx        — календарь всех поездок
│   ├── lib/
│   │   ├── types.ts                — единственные клиентские типы (зеркалят ARCH-03)
│   │   ├── schemas.ts              — Zod-схемы для валидации
│   │   ├── queryClient.ts          — QueryClient (retry:1, refetchOnWindowFocus:false)
│   │   ├── calendar.ts             — утилиты календаря (преобразование Trip → Event)
│   │   └── dates.ts                — утилиты работы с датами (date-fns)
│   └── test/
│       └── setup.ts                — тестовый конфиг (Testing Library + jsdom)
├── playwright.config.ts            — конфиг Playwright (testDir: ./e2e)
├── e2e/
│   ├── auth.spec.ts                — e2e: регистрация → логин → список поездок
│   └── trips.spec.ts               — e2e: создание поездки → детали → добавление точки → экспорт
├── vite.config.ts                  — Vite + React plugin + proxy /api → localhost:8080
├── tailwind.config.ts              — Tailwind конфигурация
└── tsconfig.json                   — TypeScript конфиг
```

## Роутинг

```tsx
// App.tsx
<Routes>
  <Route path="/login" element={<Login />} />
  <Route path="/register" element={<Register />} />
  <Route element={<ProtectedLayout />}>
    <Route path="/" element={<TripsList />} />
    <Route path="/trips/:id" element={<TripDetail />} />
    <Route path="/trips/:id/edit" element={<TripEdit />} />
    <Route path="/calendar" element={<CalendarPage />} />
  </Route>
</Routes>
```

`ProtectedLayout` проверяет авторизацию через `useQuery(['me'])`. Если пользователь не авторизован — редирект на `/login`.

## HTTP-клиент

### Axios-инстанс (`src/api/client.ts`)

```typescript
const client = axios.create({
  baseURL: '/api/v1',       // относительный путь
  withCredentials: true,    // cookies для refresh token
});
```

### Interceptor: отправка access-токена

```typescript
client.interceptors.request.use(config => {
  const token = getToken(); // из localStorage / memory
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### Interceptor: автоматический refresh на 401

```typescript
let isRefreshing = false;
let retryQueue: any[] = [];

client.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401 && !error.config._retry) {
      if (isRefreshing) {
        return new Promise(resolve => retryQueue.push(resolve));
      }
      error.config._retry = true;
      isRefreshing = true;
      try {
        await client.post('/auth/refresh'); // refresh через httpOnly cookie
        // получить новый access-token из ответа
        isRefreshing = false;
        retryQueue.forEach(r => r());
        return client(error.config);
      } catch {
        clearAuth();
        window.location.href = '/login';
        return Promise.reject(error);
      }
    }
    return Promise.reject(toApiError(error));
  }
);
```

**Важно**: фронтенд использует **относительные пути** `/api/v1/...` — никаких абсолютных URL. В dev режиме Vite проксирует `/api` → `localhost:8080` (настроено в `vite.config.ts`).

## Типы

Единственный источник типов — `src/lib/types.ts`. Эти типы зеркалят контракты бэкенда (ARCH-03): менять сигнатуры нельзя.

```typescript
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

export interface ApiErrorBody {
  error: {
    code: 'validation' | 'not_found' | 'unauthorized' | 'conflict' | 'internal';
    message: string;
    details?: unknown;
  };
}
```

## Серверный стейт (React Query)

```typescript
// lib/queryClient.ts
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});
```

Пример запроса:
```typescript
const { data: trips } = useQuery({
  queryKey: ['trips'],
  queryFn: () => tripsApi.list(),
});
```

## Валидация форм

Используется связка **React Hook Form + Zod**:

```typescript
// lib/schemas.ts
export const tripSchema = z.object({
  title: z.string().min(1, 'Обязательное поле'),
  description: z.string(),
  start_date: z.string().datetime(),
  end_date: z.string().datetime().refine(
    (v, ctx) => v >= ctx.parent.start_date,
    'Дата окончания не может быть раньше начала'
  ),
});
```

## Карта (MapView)

`src/components/MapView.tsx` — компонент на базе Leaflet:

- **Маркеры**: divIcon с номером по `stop.order`.
- **Выбранный маркер**: фон `bg-red-600`.
- **Drag & Drop**: `onMoveStop` — обновление координат точки.
- **Клик**: `onSelectStop` — выбор точки для отображения.
- **Полилайн**: соединяет точки в порядке `order`.

Плитки: OpenStreetMap (без API-ключа).

## Тестирование

### Unit-тесты (Vitest + Testing Library)

```bash
make frontend-test   # vitest --run
```

Тесты покрывают:
- `*.test.tsx` — компоненты (рендер, props, user events)
- `*.test.ts` — утилиты, API-клиент, обработка ошибок, календарь

### E2E-тесты (Playwright)

```bash
npm --prefix frontend run e2e
```

Сценарии:
- **auth.spec.ts**: регистрация → логин → проверка списка поездок
- **trips.spec.ts**: создание поездки → просмотр деталей → добавление точки → экспорт CSV

## Команды

| Команда | Описание |
|---|---|
| `make frontend-run` | Vite dev-сервер на :5173 с прокси /api → :8080 |
| `make frontend-build` | `tsc -b && vite build` → `frontend/dist/` |
| `make frontend-test` | `vitest --run` (unit-тесты) |
| `make frontend-lint` | `prettier --check . && eslint .` |
