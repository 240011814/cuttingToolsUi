import { request } from '../request'

/** 已注册的可调度任务列表 */
export function fetchJobTasks() {
  return request<Api.Job.TaskMeta[]>({
    url: '/api/admin/jobs/tasks',
    method: 'get'
  })
}

/** 定时任务定义列表 */
export function fetchJobs() {
  return request<Api.Job.JobDefinition[]>({
    url: '/api/admin/jobs',
    method: 'get'
  })
}

/** 创建定时任务 */
export function createJob(data: Api.Job.CreateJobParams) {
  return request<Api.Job.JobDefinition>({
    url: '/api/admin/jobs',
    method: 'post',
    data
  })
}

/** 更新定时任务 */
export function updateJob(id: number, data: Api.Job.UpdateJobParams) {
  return request<Api.Job.JobDefinition>({
    url: `/api/admin/jobs/${id}`,
    method: 'put',
    data
  })
}

/** 删除定时任务 */
export function deleteJob(id: number) {
  return request<boolean>({
    url: `/api/admin/jobs/${id}`,
    method: 'delete'
  })
}

/** 手动立即执行 */
export function runJob(id: number) {
  return request<boolean>({
    url: `/api/admin/jobs/${id}/run`,
    method: 'post'
  })
}

/** 执行历史 */
export function fetchJobRuns(id: number) {
  return request<Api.Job.JobRun[]>({
    url: `/api/admin/jobs/${id}/runs`,
    method: 'get'
  })
}
