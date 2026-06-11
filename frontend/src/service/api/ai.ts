import { request } from "../request";
import { getAuthorization, handleExpiredRequest } from "../request/shared";
import { getServiceBaseURL } from "@/utils/service";
import { request as requestInstance } from "../request";

export function fetchGetAIModels() {
  return request<Api.Admin.AIModel[]>({ url: "/api/ai/models" });
}

export function fetchGetUserPrompt(moduleKey: string) {
  return request<{
    effective_prompt: string;
    memory_search_query: string;
    memory_search_top_k: number;
    default_prompt: string;
    versions: any[];
    is_customized: boolean;
  }>({
    url: `/api/user-prompts/${moduleKey}`,
  });
}

export function fetchSaveUserPrompt(
  moduleKey: string,
  prompt: string,
  remark?: string,
  memorySearchQuery?: string,
  memorySearchTopK?: number,
) {
  return request({
    url: `/api/user-prompts/${moduleKey}`,
    method: "post",
    data: { prompt, remark, memory_search_query: memorySearchQuery, memory_search_top_k: memorySearchTopK },
  });
}

export function fetchSwitchUserPrompt(moduleKey: string, versionId: number) {
  return request({
    url: `/api/user-prompts/${moduleKey}/switch`,
    method: "put",
    data: { version_id: versionId },
  });
}

export function fetchDeleteUserPromptVersion(
  moduleKey: string,
  versionId: number,
) {
  return request({
    url: `/api/user-prompts/${moduleKey}/versions/${versionId}`,
    method: "delete",
  });
}

export function fetchResetUserPrompt(moduleKey: string) {
  return request({
    url: `/api/user-prompts/${moduleKey}`,
    method: "delete",
  });
}

/**
 * Chat streaming API - direct fetch with proper auth interceptors
 * Returns raw Response object with ReadableStream for SSE handling
 */
export async function fetchChatStream(data: {
  history_id: number;
  training_type: string;
  custom_training_id?: number;
  model: string;
  messages: any[];
}): Promise<Response> {
  const isHttpProxy =
    import.meta.env.DEV && import.meta.env.VITE_HTTP_PROXY === "Y";
  const { baseURL } = getServiceBaseURL(import.meta.env, isHttpProxy);

  let Authorization = getAuthorization();

  let response = await fetch(`${baseURL}/api/chat`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: Authorization || "",
    },
    body: JSON.stringify(data),
  });

  // Handle token expiration - check first chunk for error codes
  if (response.ok) {
    try {
      const contentType = response.headers.get("content-type");
      if (contentType?.includes("text/event-stream")) {
        // For SSE streams, we check the response but don't consume the body
        // Just return the response as-is for streaming
        return response;
      }

      if (contentType?.includes("application/json")) {
        const errorData = await response.json();
        const expiredTokenCodes =
          import.meta.env.VITE_SERVICE_EXPIRED_TOKEN_CODES?.split(",") || [];

        if (expiredTokenCodes.includes(String(errorData.code))) {
          // Try to refresh token
          const success = await handleExpiredRequest(requestInstance.state);
          if (success) {
            // Retry with new token
            Authorization = getAuthorization();
            response = await fetch(`${baseURL}/api/chat`, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                Authorization: Authorization || "",
              },
              body: JSON.stringify(data),
            });
          }
        }
      }
    } catch {
      // Continue normally if checking fails
    }
  }

  return response;
}
