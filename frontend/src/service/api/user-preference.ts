import { request } from '../request';

/** 获取用户主题配置 */
export function fetchGetThemePreference() {
  return request<App.Theme.ThemeSetting>({ url: '/api/user/preferences/theme' });
}

/** 保存用户主题配置 */
export function fetchSaveThemePreference(data: App.Theme.ThemeSetting) {
  return request<null>({ url: '/api/user/preferences/theme', method: 'put', data });
}

/** 获取用户通知渠道配置 */
export function fetchGetNotificationPreference() {
  return request<string[]>({ url: '/api/user/preferences/notification' });
}

/** 保存用户通知渠道配置 */
export function fetchSaveNotificationPreference(data: string[]) {
  return request<null>({ url: '/api/user/preferences/notification', method: 'put', data });
}
