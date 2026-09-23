import { request } from '../request';

export interface AISkillItem {
  id: number;
  name: string;
  description: string;
  context: string;
  agent: string;
  model: string;
  content: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface AISkillCreateRequest {
  name: string;
  description: string;
  context: string;
  agent: string;
  model: string;
  content: string;
  enabled: boolean;
}

export function fetchSkills() {
  return request<AISkillItem[]>({ url: '/api/skills' });
}

export function fetchCreateSkill(data: AISkillCreateRequest) {
  return request<AISkillItem>({ url: '/api/skills', method: 'post', data });
}

export function fetchUpdateSkill(id: number, data: Partial<AISkillCreateRequest>) {
  return request<null>({ url: `/api/skills/${id}`, method: 'put', data });
}

export function fetchDeleteSkill(id: number) {
  return request<null>({ url: `/api/skills/${id}`, method: 'delete' });
}