import type { ReactNode } from "react";

type Props = {
  header: ReactNode;
  children: ReactNode;
};

export function AppShell({ header, children }: Props) {
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 p-4 md:p-6">
      {header}
      {children}
    </main>
  );
}
