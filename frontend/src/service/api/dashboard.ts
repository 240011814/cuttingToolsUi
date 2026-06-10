import { request } from '../request';

export interface DashboardStats {
  today_messages: number;
  total_messages: number;
  total_vocabulary: number;
  total_notes: number;
  total_favorites: number;
  training_trend: TrendItem[];
  training_type_stats: TypeStatItem[];
}

export interface TrendItem {
  date: string;
  count: number;
}

export interface TypeStatItem {
  type: string;
  count: number;
}

export function fetchDashboardStats() {
  return request<DashboardStats>({
    url: '/api/dashboard/stats',
    method: 'get'
  });
}
