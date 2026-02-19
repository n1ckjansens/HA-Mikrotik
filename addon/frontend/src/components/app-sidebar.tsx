import { useEffect, useState } from "react";
import { ChevronRight, Monitor, SlidersHorizontal, ToggleLeft, Workflow } from "lucide-react";
import { NavLink, useLocation } from "react-router-dom";

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem
} from "@/components/ui/sidebar";

const automationItems = [
  { to: "/automation/capabilities", label: "Capabilities", icon: SlidersHorizontal },
  { to: "/automation/assignments", label: "Assignments", icon: ToggleLeft },
  { to: "/automation/primitives", label: "Primitives", icon: Workflow }
];

export function AppSidebar() {
  const location = useLocation();
  const pathname = location.pathname;

  const automationOpen = pathname.startsWith("/automation");
  const [isAutomationExpanded, setIsAutomationExpanded] = useState(automationOpen);

  useEffect(() => {
    if (automationOpen) {
      setIsAutomationExpanded(true);
    }
  }, [automationOpen]);

  return (
    <Sidebar variant="inset">
      <SidebarHeader>
        <p className="text-sm font-semibold">MikroTik Presence</p>
        <p className="text-xs text-muted-foreground">Home Assistant add-on</p>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild isActive={pathname === "/"}>
                  <NavLink to="/">
                    <Monitor className="h-4 w-4" />
                    <span>Devices</span>
                  </NavLink>
                </SidebarMenuButton>
              </SidebarMenuItem>

              <SidebarMenuItem>
                <Collapsible
                  open={isAutomationExpanded}
                  onOpenChange={setIsAutomationExpanded}
                  className="group/collapsible"
                >
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton isActive={automationOpen}>
                      <Workflow className="h-4 w-4" />
                      <span className="flex-1 text-left">Automation</span>
                      <ChevronRight className="h-4 w-4 transition-transform group-data-[state=open]/collapsible:rotate-90" />
                    </SidebarMenuButton>
                  </CollapsibleTrigger>

                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {automationItems.map((item) => {
                        const active = pathname === item.to;
                        return (
                          <SidebarMenuSubItem key={item.to}>
                            <SidebarMenuSubButton asChild isActive={active}>
                              <NavLink to={item.to}>
                                <item.icon className="mr-2 h-3.5 w-3.5" />
                                <span>{item.label}</span>
                              </NavLink>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        );
                      })}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </Collapsible>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
