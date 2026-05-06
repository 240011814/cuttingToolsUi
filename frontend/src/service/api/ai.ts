import { request } from '../request';

export function fetchGetAIModels() {
  return request<Api.Admin.AIModel[]>({ url: '/api/ai/models' });
}
