import { request } from '../request';

export interface CustomTraining {
  id: number;
  user_id: number;
  title: string;
  description: string;
  system_prompt: string;
  icon: string;
  color: string;
  initial_message: string;
  input_placeholder: string;
  speech_lang: string;
  speech_rate: number;
  is_favorite: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateCustomTrainingParams {
  title: string;
  description?: string;
  system_prompt: string;
  icon?: string;
  color?: string;
  initial_message?: string;
  input_placeholder?: string;
  speech_lang?: string;
  speech_rate?: number;
}

export interface UpdateCustomTrainingParams {
  title?: string;
  description?: string;
  system_prompt?: string;
  icon?: string;
  color?: string;
  initial_message?: string;
  input_placeholder?: string;
  speech_lang?: string;
  speech_rate?: number;
}

export function fetchCustomTrainingList() {
  return request<CustomTraining[]>({
    url: '/api/custom-trainings',
    method: 'get'
  });
}

export function fetchCustomTrainingDetail(id: number) {
  return request<CustomTraining>({
    url: `/api/custom-trainings/${id}`,
    method: 'get'
  });
}

export function fetchCreateCustomTraining(data: CreateCustomTrainingParams) {
  return request<CustomTraining>({
    url: '/api/custom-trainings',
    method: 'post',
    data
  });
}

export function fetchUpdateCustomTraining(id: number, data: UpdateCustomTrainingParams) {
  return request({
    url: `/api/custom-trainings/${id}`,
    method: 'put',
    data
  });
}

export function fetchDeleteCustomTraining(id: number) {
  return request({
    url: `/api/custom-trainings/${id}`,
    method: 'delete'
  });
}
