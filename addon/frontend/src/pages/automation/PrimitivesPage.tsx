import { Workflow } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { useActionTypes } from "@/hooks/useActionTypes";
import { useStateSourceTypes } from "@/hooks/useStateSourceTypes";

export function PrimitivesPage() {
  const actionTypesQuery = useActionTypes();
  const stateSourceTypesQuery = useStateSourceTypes();

  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Primitives</h1>
        <p className="text-sm text-muted-foreground">
          Backend-supported action types and state sources available for automation.
        </p>
      </header>

      <div className="space-y-3">
        {(actionTypesQuery.data ?? []).map((actionType) => (
          <Card key={actionType.id}>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Workflow className="h-4 w-4" />
                {actionType.label}
              </CardTitle>
              <p className="text-xs text-muted-foreground">{actionType.id}</p>
              <p className="text-sm text-muted-foreground">{actionType.description}</p>
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Key</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Required</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Options</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {actionType.param_schema.map((field) => (
                    <TableRow key={field.key}>
                      <TableCell>{field.key}</TableCell>
                      <TableCell>{field.kind}</TableCell>
                      <TableCell>
                        {field.required ? (
                          <Badge variant="secondary">required</Badge>
                        ) : (
                          <Badge variant="outline">optional</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {field.description || "-"}
                      </TableCell>
                      <TableCell>
                        {(field.options ?? []).length > 0
                          ? (field.options ?? []).join(", ")
                          : "-"}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="space-y-3">
        {(stateSourceTypesQuery.data ?? []).map((sourceType) => (
          <Card key={sourceType.id}>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Workflow className="h-4 w-4" />
                {sourceType.label}
              </CardTitle>
              <p className="text-xs text-muted-foreground">{sourceType.id}</p>
              <p className="text-sm text-muted-foreground">{sourceType.description}</p>
              <div>
                <Badge variant="outline">output: {sourceType.output_type}</Badge>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Key</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Required</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Options</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sourceType.param_schema.map((field) => (
                    <TableRow key={field.key}>
                      <TableCell>{field.key}</TableCell>
                      <TableCell>{field.kind}</TableCell>
                      <TableCell>
                        {field.required ? (
                          <Badge variant="secondary">required</Badge>
                        ) : (
                          <Badge variant="outline">optional</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {field.description || "-"}
                      </TableCell>
                      <TableCell>
                        {(field.options ?? []).length > 0
                          ? (field.options ?? []).join(", ")
                          : "-"}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
