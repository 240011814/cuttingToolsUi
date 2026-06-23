import { request } from '../request';

export namespace Api.ModelScenario {
  export interface Item {
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

  export interface CreateRequest {
    type: string;
    name: string;
    summary: string;
    description: string;
    detail: string;
    category: string;
    sortOrder: number;
  }

  export interface UpdateRequest {
    name?: string;
    summary?: string;
    description?: string;
    detail?: string;
    category?: string;
    sortOrder?: number;
  }
}

export function fetchModelScenarios(params?: { type?: string }) {
  return request<Api.ModelScenario.Item[]>({ url: '/api/model-scenario', params });
}

export function fetchCreateModelScenario(data: Api.ModelScenario.CreateRequest) {
  return request<Api.ModelScenario.Item>({ url: '/api/model-scenario', method: 'post', data });
}

export function fetchUpdateModelScenario(id: number, data: Api.ModelScenario.UpdateRequest) {
  return request<null>({ url: `/api/model-scenario/${id}`, method: 'put', data });
}

export function fetchDeleteModelScenario(id: number) {
  return request<null>({ url: `/api/model-scenario/${id}`, method: 'delete' });
}
