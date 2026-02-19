import { withAppBase } from "@/lib/ingress";

export class ApiError extends Error {
  public code: string;

  constructor(message: string, code = "unknown_error") {
    super(message);
    this.name = "ApiError";
    this.code = code;
  }
}

async function parseError(response: Response): Promise<never> {
  let message = `Request failed with status ${response.status}`;
  let code = "unknown_error";
  try {
    const payload = await response.json();
    if (payload?.error?.message) {
      message = payload.error.message;
    }
    if (payload?.error?.code) {
      code = payload.error.code;
    }
  } catch {
    // noop
  }
  throw new ApiError(message, code);
}

export async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(withAppBase(path), {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    }
  });

  if (!response.ok) {
    await parseError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}
