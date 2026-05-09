import { request } from '../request';

export interface TrainingHistory {
  id: number;
  user_id: number;
  training_type: string;
  title: string;
  is_favorite: boolean;
  messages: string;
  created_at: string;
  updated_at: string;
}

export interface ListHistoryParams {
  page: number;
  pageSize: number;
  title?: string;
  is_favorite?: boolean;
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

export function fetchHistoryDetail(id: number) {
  return request<TrainingHistory>({
    url: `/api/histories/${id}`,
    method: 'get'
  });
}

export function fetchUpdateFavorite(id: number, is_favorite: boolean) {
  return request({
    url: `/api/histories/${id}/favorite`,
    method: 'put',
    data: { is_favorite }
  });
}

export function fetchUpdateHistoryTitle(id: number, title: string) {
  return request({
    url: `/api/histories/${id}/title`,
    method: 'put',
    data: { title }
  });
}
