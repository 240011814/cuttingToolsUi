import { request } from '../request';

export function fetchGetAIModels() {
  return request<Api.Admin.AIModel[]>({ url: '/api/ai/models' });
}

export function fetchGetUserPrompt(moduleKey: string) {
  return request<{
    effective_prompt: string;
    default_prompt: string;
    versions: any[];
    is_customized: boolean;
  }>({
    url: `/api/user-prompts/${moduleKey}`
  });
}

export function fetchSaveUserPrompt(moduleKey: string, prompt: string, remark?: string) {
  return request({
    url: `/api/user-prompts/${moduleKey}`,
    method: 'post',
    data: { prompt, remark }
  });
}

export function fetchSwitchUserPrompt(moduleKey: string, versionId: number) {
  return request({
    url: `/api/user-prompts/${moduleKey}/switch`,
    method: 'put',
    data: { version_id: versionId }
  });
}

export function fetchDeleteUserPromptVersion(moduleKey: string, versionId: number) {
  return request({
    url: `/api/user-prompts/${moduleKey}/versions/${versionId}`,
    method: 'delete'
  });
}

export function fetchResetUserPrompt(moduleKey: string) {
  return request({
    url: `/api/user-prompts/${moduleKey}`,
    method: 'delete'
  });
}
