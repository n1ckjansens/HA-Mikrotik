export function Header() {
  return (
    <header className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-semibold">MikroTik Presence</h1>
        <p className="text-sm text-muted-foreground">RouterOS v7 device monitoring</p>
      </div>
    </header>
  );
}
