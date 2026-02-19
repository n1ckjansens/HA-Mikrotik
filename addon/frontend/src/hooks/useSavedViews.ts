import { useEffect, useMemo, useState } from "react";

import type { OnlineScope, RegistrationScope } from "@/lib/device-semantics";

const STORAGE_KEY = "mikrotik-presence.saved-views.v1";

export type SavedView = {
  id: string;
  name: string;
  registrationScope: RegistrationScope;
  onlineScope: OnlineScope;
  search: string;
  vendors: string[];
  sources: string[];
  subnets: string[];
};

function loadSavedViews(): SavedView[] {
  if (typeof window === "undefined") {
    return [];
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return [];
    }
    const parsed = JSON.parse(raw) as SavedView[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function persistSavedViews(items: SavedView[]) {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
}

export function useSavedViews() {
  const [savedViews, setSavedViews] = useState<SavedView[]>([]);

  useEffect(() => {
    setSavedViews(loadSavedViews());
  }, []);

  const byId = useMemo(
    () => new Map(savedViews.map((view) => [view.id, view])),
    [savedViews]
  );

  const saveView = (view: Omit<SavedView, "id" | "name">) => {
    const timestamp = new Date();
    const name = `View ${timestamp.toLocaleTimeString()}`;
    const next: SavedView = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      name,
      ...view
    };
    const updated = [next, ...savedViews].slice(0, 12);
    setSavedViews(updated);
    persistSavedViews(updated);
    return next;
  };

  const removeView = (id: string) => {
    const updated = savedViews.filter((view) => view.id !== id);
    setSavedViews(updated);
    persistSavedViews(updated);
  };

  return {
    savedViews,
    getView: (id: string) => byId.get(id),
    saveView,
    removeView
  };
}
