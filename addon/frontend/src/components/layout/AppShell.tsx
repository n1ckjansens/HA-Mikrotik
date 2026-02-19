import type { ReactNode } from "react";

type Props = {
  header: ReactNode;
  children: ReactNode;
};

export function AppShell({ header, children }: Props) {
  return (
    <main className="flex flex-col gap-6">
      {header}
      {children}
    </main>
  );
}
