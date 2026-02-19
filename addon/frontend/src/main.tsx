import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { AppLayout } from "@/components/layout/AppLayout";
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { APP_BASE_PATH } from "@/lib/ingress";
import { DevicesPage } from "@/pages/DevicesPage";
import { AssignmentsPage } from "@/pages/automation/AssignmentsPage";
import { CapabilitiesPage } from "@/pages/automation/CapabilitiesPage";
import { CapabilityEditorPage } from "@/pages/automation/CapabilityEditorPage";
import { GlobalCapabilitiesPage } from "@/pages/automation/GlobalCapabilitiesPage";
import { PrimitivesPage } from "@/pages/automation/PrimitivesPage";

import "./index.css";

const queryClient = new QueryClient();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <BrowserRouter basename={APP_BASE_PATH}>
          <Routes>
            <Route element={<AppLayout />}>
              <Route path="/" element={<DevicesPage />} />
              <Route path="/automation" element={<Navigate to="/automation/capabilities" replace />} />
              <Route path="/automation/capabilities" element={<CapabilitiesPage />} />
              <Route path="/automation/capabilities/new" element={<CapabilityEditorPage />} />
              <Route path="/automation/capabilities/:id" element={<CapabilityEditorPage />} />
              <Route path="/automation/global" element={<GlobalCapabilitiesPage />} />
              <Route path="/automation/assignments" element={<AssignmentsPage />} />
              <Route path="/automation/primitives" element={<PrimitivesPage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </TooltipProvider>
      <Toaster richColors closeButton />
    </QueryClientProvider>
  </StrictMode>
);
