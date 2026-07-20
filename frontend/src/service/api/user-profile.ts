import { request } from '../request';

export function fetchGetUserProfile() {
  return request<Api.Admin.UserProfile>({ url: '/api/user/profile' });
}

export function fetchUpdateProfile(data: Api.Admin.UpdateProfileParams) {
  return request({ url: '/api/user/profile', method: 'put', data });
}

export function fetchChangePassword(data: Api.Admin.ChangePasswordParams) {
  return request({ url: '/api/user/password', method: 'put', data });
}

// Telegram Binding APIs
export function fetchGetTelegramConfig() {
  return request<Api.Telegram.ConfigResponse>({ url: '/api/telegram/config' });
}

export function fetchGetTelegramStatus() {
  return request<Api.Telegram.StatusResponse>({ url: '/api/telegram/status' });
}

export function fetchGenerateTelegramBindCode() {
  return request<Api.Telegram.BindCodeResponse>({ url: '/api/telegram/bind-code', method: 'post' });
}

export function fetchUnbindTelegram() {
  return request({ url: '/api/telegram/unbind', method: 'post' });
}
