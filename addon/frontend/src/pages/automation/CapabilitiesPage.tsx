import { useMemo, useState } from "react";
import { Copy, Pencil, Plus, Search, Trash2 } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { useCapabilities } from "@/hooks/useCapabilities";
import { useCapabilityEditor } from "@/hooks/useCapabilityEditor";
import { categoriesFromCapabilities, countActions } from "@/lib/automation";
import type { CapabilityTemplate } from "@/types/automation";

function duplicateTemplate(input: CapabilityTemplate): CapabilityTemplate {
  const suffix = Math.floor(Date.now() / 1000).toString(36);
  return {
    ...input,
    id: `${input.id}.copy_${suffix}`,
    label: `${input.label} (Copy)`
  };
}

export function CapabilitiesPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState("all");
  const capabilityEditor = useCapabilityEditor(null);

  const capabilitiesQuery = useCapabilities({
    search,
    category: category === "all" ? "" : category
  });

  const categories = useMemo(
    () => categoriesFromCapabilities(capabilitiesQuery.data ?? []),
    [capabilitiesQuery.data]
  );

  const handleDelete = async (item: CapabilityTemplate) => {
    try {
      await capabilityEditor.deleteMutation.mutateAsync(item.id);
      toast.success(`Capability ${item.label} deleted`);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Delete failed";
      toast.error(message);
    }
  };

  const handleDuplicate = async (item: CapabilityTemplate) => {
    try {
      const created = await capabilityEditor.createMutation.mutateAsync(duplicateTemplate(item));
      navigate(`/automation/capabilities/${encodeURIComponent(created.id)}`);
      toast.success(`Capability ${item.label} duplicated`);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Duplicate failed";
      toast.error(message);
    }
  };

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Capabilities</h1>
          <p className="text-sm text-muted-foreground">
            Define what devices can do and what actions are executed behind each control.
          </p>
        </div>
        <Button asChild>
          <Link to="/automation/capabilities/new">
            <Plus className="mr-2 h-4 w-4" />
            New capability
          </Link>
        </Button>
      </header>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Filters</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 sm:grid-cols-2">
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="pl-9"
              placeholder="Search by name or id"
            />
          </div>

          <Select value={category} onValueChange={setCategory}>
            <SelectTrigger>
              <SelectValue placeholder="All categories" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All categories</SelectItem>
              {categories.map((item) => (
                <SelectItem key={item} value={item}>
                  {item}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Scope</TableHead>
                <TableHead>Control</TableHead>
                <TableHead>States</TableHead>
                <TableHead>HA expose</TableHead>
                <TableHead>Actions</TableHead>
                <TableHead className="text-right">Manage</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(capabilitiesQuery.data ?? []).map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <p className="font-medium">{item.label}</p>
                    <p className="text-xs text-muted-foreground">{item.id}</p>
                  </TableCell>
                  <TableCell>{item.category || "-"}</TableCell>
                  <TableCell>
                    <Badge variant={item.scope === "global" ? "secondary" : "outline"}>
                      {item.scope}
                    </Badge>
                  </TableCell>
                  <TableCell className="capitalize">{item.control.type}</TableCell>
                  <TableCell>{Object.keys(item.states).length}</TableCell>
                  <TableCell>
                    {item.ha_expose.enabled ? (
                      <Badge variant="secondary">Enabled</Badge>
                    ) : (
                      <Badge variant="outline">Disabled</Badge>
                    )}
                  </TableCell>
                  <TableCell>{countActions(item)}</TableCell>
                  <TableCell className="text-right">
                    <div className="inline-flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => navigate(`/automation/capabilities/${encodeURIComponent(item.id)}`)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          void handleDuplicate(item);
                        }}
                        disabled={capabilityEditor.createMutation.isPending}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          void handleDelete(item);
                        }}
                        disabled={capabilityEditor.deleteMutation.isPending}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
