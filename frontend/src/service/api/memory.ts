import { request } from "../request";

export interface Mem0Memory {
  id: string;
  memory: string;
  score?: number;
  user_id?: string;
  metadata?: any;
  categories?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface Mem0AddV3Response {
  message: string;
  status: "PENDING" | "SUCCEEDED" | "FAILED";
  event_id: string;
}

export interface Mem0Message {
  role: "user" | "assistant";
  content: string;
}

export interface Mem0PaginatedResponse {
  count: number;
  next: string | null;
  previous: string | null;
  results: Mem0Memory[];
}

/** 添加记忆 (v3 异步接口) */
export function fetchAddMemory(messages: Mem0Message[], metadata?: Record<string, any>) {
  return request<Mem0AddV3Response>({
    url: "/api/memories",
    method: "post",
    data: { messages, metadata },
  });
}

/** 搜索相关记忆 (v3) */
export function fetchSearchMemories(query: string, topK?: number) {
  return request<Mem0Memory[]>({
    url: "/api/memories/search",
    method: "post",
    data: { query, top_k: topK || 5 },
  });
}

/** 获取所有记忆列表 (v3 分页) */
export function fetchListMemories(page?: number, pageSize?: number) {
  return request<Mem0PaginatedResponse>({
    url: "/api/memories",
    method: "get",
    params: { page: page || 1, page_size: pageSize || 100 },
  });
}

/** 删除指定记忆 (v1) */
export function fetchDeleteMemory(memoryId: string) {
  return request({
    url: `/api/memories/${memoryId}`,
    method: "delete",
  });
}
