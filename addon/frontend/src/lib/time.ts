const SECOND = 1000;
const MINUTE = 60 * SECOND;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

function age(value: string | number | Date, nowMs: number): number {
  const ts = value instanceof Date ? value.getTime() : new Date(value).getTime();
  return Math.max(0, nowMs - ts);
}

export function formatUpdatedAgo(updatedAt: number | undefined, nowMs: number) {
  if (!updatedAt) {
    return "Updated just now";
  }

  const delta = age(updatedAt, nowMs);
  if (delta < MINUTE) {
    const seconds = Math.max(1, Math.floor(delta / SECOND));
    return `Updated ${seconds}s ago`;
  }
  if (delta < HOUR) {
    const minutes = Math.floor(delta / MINUTE);
    return `Updated ${minutes}m ago`;
  }
  if (delta < DAY) {
    const hours = Math.floor(delta / HOUR);
    return `Updated ${hours}h ago`;
  }
  const days = Math.floor(delta / DAY);
  return `Updated ${days}d ago`;
}

export function formatLastSeenLabel(
  online: boolean,
  lastSeenAt: string | null | undefined,
  nowMs: number
) {
  if (online) {
    return "Active now";
  }
  if (!lastSeenAt) {
    return "Never seen";
  }

  const delta = age(lastSeenAt, nowMs);
  if (delta < HOUR) {
    const minutes = Math.max(1, Math.floor(delta / MINUTE));
    return `${minutes}m ago`;
  }
  if (delta < DAY) {
    const hours = Math.floor(delta / HOUR);
    return `${hours}h ago`;
  }
  if (delta < 2 * DAY) {
    return "Yesterday";
  }
  const days = Math.floor(delta / DAY);
  return `${days}d ago`;
}

export function formatExactTimestamp(value: string | null | undefined) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "medium"
  }).format(date);
}
