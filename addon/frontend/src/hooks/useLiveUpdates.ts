import { useState } from "react";

export function useLiveUpdates() {
  const [isPaused, setIsPaused] = useState(false);

  return {
    isPaused,
    pause: () => setIsPaused(true),
    resume: () => setIsPaused(false),
    toggle: () => setIsPaused((current) => !current)
  };
}
