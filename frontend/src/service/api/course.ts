import { request } from '../request';

export interface Course {
  id: number;
  user_id: number;
  title: string;
  description: string;
  tags: string;
  is_public: boolean;
  item_count: number;
  created_at: string;
  updated_at: string;
}

export interface CourseItem {
  id: number;
  course_id: number;
  english_sentence: string;
  chinese_translation: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface CourseDetail extends Course {
  items: CourseItem[];
}

export interface CreateCourseParams {
  title: string;
  description?: string;
  tags?: string[];
  is_public?: boolean;
}

export interface UpdateCourseParams {
  title?: string;
  description?: string;
  tags?: string[];
  is_public?: boolean;
}

export interface CreateCourseItemParams {
  english_sentence: string;
  chinese_translation?: string;
  sort_order?: number;
}

export interface UpdateCourseItemParams {
  english_sentence?: string;
  chinese_translation?: string;
  sort_order?: number;
}

export interface CourseListParams {
  show_all?: boolean;
  keyword?: string;
  is_public?: boolean;
  tag?: string;
  page?: number;
  page_size?: number;
}

export interface CourseListResult {
  list: Course[];
  total: number;
  page: number;
  page_size: number;
}

export function fetchCourseList(params: CourseListParams = {}) {
  return request<CourseListResult>({
    url: '/api/courses',
    method: 'get',
    params
  });
}

export function fetchCourseDetail(id: number) {
  return request<CourseDetail>({
    url: `/api/courses/${id}`,
    method: 'get'
  });
}

export function fetchCreateCourse(data: CreateCourseParams) {
  return request<Course>({
    url: '/api/courses',
    method: 'post',
    data
  });
}

export function fetchUpdateCourse(id: number, data: UpdateCourseParams) {
  return request<Course>({
    url: `/api/courses/${id}`,
    method: 'put',
    data
  });
}

export function fetchDeleteCourse(id: number) {
  return request({
    url: `/api/courses/${id}`,
    method: 'delete'
  });
}

export function fetchCourseItems(courseId: number) {
  return request<CourseItem[]>({
    url: `/api/courses/${courseId}/items`,
    method: 'get'
  });
}

export function fetchCreateCourseItem(courseId: number, data: CreateCourseItemParams) {
  return request<CourseItem>({
    url: `/api/courses/${courseId}/items`,
    method: 'post',
    data
  });
}

export function fetchBatchCreateCourseItems(courseId: number, items: CreateCourseItemParams[]) {
  return request<CourseItem[]>({
    url: `/api/courses/${courseId}/items/batch`,
    method: 'post',
    data: { items }
  });
}

export function fetchUpdateCourseItem(courseId: number, itemId: number, data: UpdateCourseItemParams) {
  return request<CourseItem>({
    url: `/api/courses/${courseId}/items/${itemId}`,
    method: 'put',
    data
  });
}

export function fetchDeleteCourseItem(courseId: number, itemId: number) {
  return request({
    url: `/api/courses/${courseId}/items/${itemId}`,
    method: 'delete'
  });
}

export function fetchBatchDeleteCourseItems(courseId: number, ids: number[]) {
  return request({
    url: `/api/courses/${courseId}/items/batch`,
    method: 'delete',
    data: { ids }
  });
}
