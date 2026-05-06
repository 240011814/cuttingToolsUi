import { request } from '../request';

export function fetchAddNote(data: {
  category: string;
  content: string;
}) {
  return request<any>({
    url: '/api/notes',
    method: 'post',
    data
  });
}

export function fetchGetNoteList(params?: { page?: number; pageSize?: number; category?: string; content?: string }) {
  return request<{ items: any[]; total: number }>({
    url: '/api/notes',
    method: 'get',
    params
  });
}

export function fetchDeleteNote(id: number) {
  return request({
    url: `/api/notes/${id}`,
    method: 'delete'
  });
}

export function fetchUpdateNote(id: number, data: {
  category?: string;
  content?: string;
}) {
  return request({
    url: `/api/notes/${id}`,
    method: 'put',
    data
  });
}
