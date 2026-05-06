import { request } from '../request';

export interface TrainingHistory {
  id: number;
  user_id: number;
  training_type: string;
  title: string;
  record_type: string;
  messages: string;
  created_at: string;
  updated_at: string;
}

export interface ListHistoryParams {
  page: number;
  pageSize: number;
  title?: string;
  record_type?: string;
}

export interface ListHistoryResponse {
  total: number;
  items: TrainingHistory[];
}

export function fetchHistoryList(params: ListHistoryParams) {
  return request<ListHistoryResponse>({
    url: '/api/histories',
    method: 'get',
    params
  });
}

export function fetchArchiveHistory(data: { training_type: string; messages: string; title?: string }) {
  return request<{ id: number }>({
    url: '/api/histories/archive',
    method: 'post',
    data
  });
}
