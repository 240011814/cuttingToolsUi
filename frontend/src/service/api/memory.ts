import { request } from "../request";

export interface Mem0Memory {
  id: string;
  memory: string;
  user_id: string;
  metadata?: any;
  created_at?: string;
  updated_at?: string;
}

/** 搜索相关记忆 */
export function fetchSearchMemories(query: string, topK?: number) {
  return request<Mem0Memory[]>({
    url: "/api/memories/search",
    method: "post",
    data: { query, top_k: topK || 5 },
  });
}

/** 获取所有记忆列表 */
export function fetchListMemories() {
  return request<Mem0Memory[]>({
    url: "/api/memories",
  });
}

/** 删除指定记忆 */
export function fetchDeleteMemory(memoryId: string) {
  return request({
    url: `/api/memories/${memoryId}`,
    method: "delete",
  });
}
