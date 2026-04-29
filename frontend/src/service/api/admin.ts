import { request } from '../request';

export function fetchGetUsers(params?: Api.Admin.UserSearchParams) {
  return request<Api.Admin.User[]>({ url: '/api/admin/users', params });
}

export function fetchCreateUser(data: Api.Admin.CreateUserParams) {
  return request({ url: '/api/admin/users', method: 'post', data });
}

export function fetchUpdateUser(userId: number, data: Api.Admin.UpdateUserParams) {
  return request({ url: `/api/admin/users/${userId}`, method: 'put', data });
}

export function fetchDeleteUser(userId: number) {
  return request({ url: `/api/admin/users/${userId}`, method: 'delete' });
}

export function fetchGetRoles() {
  return request<Api.Admin.Role[]>({ url: '/api/admin/roles' });
}

export function fetchGetPermissions() {
  return request<Api.Admin.Permission[]>({ url: '/api/admin/permissions' });
}

export function fetchGetRolePermissions(roleCode: string) {
  return request<string[]>({ url: `/api/admin/roles/${roleCode}/permissions` });
}

export function fetchUpdateRolePermissions(roleCode: string, permissions: string[]) {
  return request({ url: `/api/admin/roles/${roleCode}/permissions`, method: 'put', data: { permissions } });
}
