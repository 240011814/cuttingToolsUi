declare namespace Api {
  namespace Job {
    interface TaskMeta {
      name: string
      description: string
      paramsExample: string | null
    }

    interface JobDefinition {
      id: number
      name: string
      taskName: string
      cronExpr: string
      params: string | null
      enabled: boolean
      maxRetries: number
      remark: string | null
      createdBy: number | null
      createdAt: string
      updatedAt: string
      nextRunAt: string | null
    }

    interface JobRun {
      id: number
      definitionId: number
      attempt: number
      status: 'running' | 'success' | 'failed' | 'skipped'
      triggerType: 'scheduler' | 'manual'
      startedAt: string | null
      finishedAt: string | null
      error: string
      createdAt: string
    }

    interface CreateJobParams {
      name: string
      taskName: string
      cronExpr: string
      params?: string | null
      enabled?: boolean
      maxRetries?: number
      remark?: string
    }

    interface UpdateJobParams {
      name?: string
      taskName?: string
      cronExpr?: string
      params?: string | null
      enabled?: boolean
      maxRetries?: number
      remark?: string
    }
  }
}
