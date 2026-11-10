// The views of the admin dashboard, which are not pages of their own: they are chosen with ?view=
export type DashboardView = 'events' | 'inbox' | 'ingestion';

export function isDashboardView(value: string | null): value is DashboardView {
  return value === 'events' || value === 'inbox' || value === 'ingestion';
}
