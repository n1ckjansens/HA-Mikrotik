function normalizePath(value: string): string {
  const trimmed = value.trim();
  if (trimmed === "" || trimmed === "/") {
    return "/";
  }
  return trimmed.replace(/\/+$/, "");
}

// Detect add-on base path in Home Assistant ingress and local dev.
export function detectAppBasePath(pathname: string): string {
  const normalized = normalizePath(pathname);
  if (normalized === "/") {
    return "/";
  }

  const automationIndex = normalized.indexOf("/automation");
  if (automationIndex === 0) {
    return "/";
  }
  if (automationIndex > 0) {
    const candidate = normalizePath(normalized.slice(0, automationIndex));
    return candidate === "" ? "/" : candidate;
  }

  return normalized;
}

export const APP_BASE_PATH =
  typeof window === "undefined"
    ? "/"
    : detectAppBasePath(window.location.pathname);

export function withAppBase(path: string): string {
  const trimmed = path.trim();
  if (trimmed === "") {
    return APP_BASE_PATH;
  }

  // Absolute URLs should pass through untouched.
  if (/^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(trimmed) || trimmed.startsWith("//")) {
    return trimmed;
  }

  const normalizedPath = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
  if (APP_BASE_PATH === "/") {
    return normalizedPath;
  }
  if (normalizedPath === APP_BASE_PATH || normalizedPath.startsWith(`${APP_BASE_PATH}/`)) {
    return normalizedPath;
  }
  return `${APP_BASE_PATH}${normalizedPath}`;
}
