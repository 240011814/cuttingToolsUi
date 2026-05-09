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
