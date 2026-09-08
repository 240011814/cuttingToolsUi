import { request } from '../request';

export interface ReminderParams {
  title: string;
  content: string;
}

export interface Reminder {
  id: number;
  jobId: number;
  userId: number;
  scheduledAt: string;
  status: string;
  params: ReminderParams;
  repeatType: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
  repeatInterval: number;
  repeatEndAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateReminderParams {
  title: string;
  content?: string;
  remindAt: string;
  repeatType?: string;
  repeatInterval?: number;
  repeatEndAt?: string | null;
}

export interface UpdateReminderParams {
  title?: string;
  content?: string;
  remindAt?: string;
  repeatType?: string;
  repeatInterval?: number;
  repeatEndAt?: string | null;
}

export function fetchGetReminders(params: { year: number; month: number }) {
  return request<Reminder[]>({ url: '/api/reminders', method: 'get', params });
}

export function fetchCreateReminder(data: CreateReminderParams) {
  return request<Reminder>({ url: '/api/reminders', method: 'post', data });
}

export function fetchUpdateReminder(id: number, data: UpdateReminderParams) {
  return request<Reminder>({ url: `/api/reminders/${id}`, method: 'put', data });
}

export function fetchDeleteReminder(id: number) {
  return request({ url: `/api/reminders/${id}`, method: 'delete' });
}