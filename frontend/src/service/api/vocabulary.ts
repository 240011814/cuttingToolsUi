import { request } from '../request';

/**
 * 添加生词
 * @param data 生词信息
 */
export function fetchAddVocabulary(data: {
  word: string;
  phonetic?: string;
  definition?: string;
  example?: string;
  sourceContext?: string;
}) {
  return request<any>({
    url: '/api/vocabulary',
    method: 'post',
    data
  });
}

/**
 * 获取生词列表
 * @param params 查询参数
 */
export function fetchGetVocabularyList(params?: { keyword?: string; isMastered?: boolean }) {
  return request<any[]>({
    url: '/api/vocabulary',
    method: 'get',
    params
  });
}

/**
 * 删除生词
 * @param id 生词 ID
 */
export function fetchDeleteVocabulary(id: number) {
  return request({
    url: `/api/vocabulary/${id}`,
    method: 'delete'
  });
}

/**
 * 更新生词
 * @param id 生词 ID
 * @param data 更新内容
 */
export function fetchUpdateVocabulary(id: number, data: {
  phonetic?: string;
  definition?: string;
  example?: string;
  isMastered?: boolean;
}) {
  return request({
    url: `/api/vocabulary/${id}`,
    method: 'put',
    data
  });
}
