import { request } from '../request';

export interface ModelScenarioItem {
  id: number;
  type: string;
  name: string;
  summary: string;
  description: string;
  detail: string;
  category: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface ModelScenarioCreateRequest {
  type: string;
  name: string;
  summary: string;
  description: string;
  detail: string;
  category: string;
  sortOrder: number;
}

export interface ModelScenarioUpdateRequest {
  name?: string;
  summary?: string;
  description?: string;
  detail?: string;
  category?: string;
  sortOrder?: number;
}

export function fetchModelScenarios(params?: { type?: string }) {
  return request<ModelScenarioItem[]>({ url: '/api/model-scenario', params });
}

export function fetchCreateModelScenario(data: ModelScenarioCreateRequest) {
  return request<ModelScenarioItem>({ url: '/api/model-scenario', method: 'post', data });
}

export function fetchUpdateModelScenario(id: number, data: ModelScenarioUpdateRequest) {
  return request<null>({ url: `/api/model-scenario/${id}`, method: 'put', data });
}

export function fetchDeleteModelScenario(id: number) {
  return request<null>({ url: `/api/model-scenario/${id}`, method: 'delete' });
}
