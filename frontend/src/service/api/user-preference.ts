import { request } from '../request';

/** 获取用户主题配置 */
export function fetchGetThemePreference() {
  return request<App.Theme.ThemeSetting>({ url: '/api/user/preferences/theme' });
}

/** 保存用户主题配置 */
export function fetchSaveThemePreference(data: App.Theme.ThemeSetting) {
  return request<null>({ url: '/api/user/preferences/theme', method: 'put', data });
}
