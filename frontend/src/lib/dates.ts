export function formatDate(iso: string): string {
  if (!iso) return '';
  const date = new Date(iso);
  const d = date.getDate().toString().padStart(2, '0');
  const m = (date.getMonth() + 1).toString().padStart(2, '0');
  const y = date.getFullYear();
  return `${d}.${m}.${y}`;
}

export function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}
