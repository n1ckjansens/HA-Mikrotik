import { Fragment } from "react";
import { Link, Outlet, useLocation } from "react-router-dom";

import { AppSidebar } from "@/components/app-sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger
} from "@/components/ui/sidebar";

type Crumb = {
  label: string;
  href?: string;
};

function decodeSegment(segment: string) {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

function buildBreadcrumbs(pathname: string): Crumb[] {
  if (pathname === "/") {
    return [{ label: "Devices" }];
  }

  const segments = pathname.split("/").filter(Boolean);
  if (segments.length === 0) {
    return [{ label: "Devices" }];
  }

  if (segments[0] !== "automation") {
    return [{ label: "Devices", href: "/" }];
  }

  const crumbs: Crumb[] = [{ label: "Automation", href: "/automation/capabilities" }];
  const section = segments[1] ?? "capabilities";

  if (section === "capabilities") {
    crumbs.push({ label: "Capabilities", href: "/automation/capabilities" });
    if (segments[2] === "new") {
      crumbs.push({ label: "New" });
    } else if (segments[2]) {
      crumbs.push({ label: decodeSegment(segments[2]) });
    }
    return crumbs;
  }

  if (section === "assignments") {
    crumbs.push({ label: "Assignments" });
    return crumbs;
  }

  if (section === "primitives") {
    crumbs.push({ label: "Primitives" });
    return crumbs;
  }

  crumbs.push({ label: decodeSegment(section) });
  return crumbs;
}

export function AppLayout() {
  const location = useLocation();
  const breadcrumbs = buildBreadcrumbs(location.pathname);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                {breadcrumbs.map((crumb, index) => {
                  const isLast = index === breadcrumbs.length - 1;
                  const hideOnMobile = index === 0 && breadcrumbs.length > 1;

                  return (
                    <Fragment key={`${crumb.label}-${index}`}>
                      <BreadcrumbItem className={hideOnMobile ? "hidden md:block" : ""}>
                        {isLast || !crumb.href ? (
                          <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                        ) : (
                          <BreadcrumbLink asChild>
                            <Link to={crumb.href}>{crumb.label}</Link>
                          </BreadcrumbLink>
                        )}
                      </BreadcrumbItem>

                      {!isLast ? (
                        <BreadcrumbSeparator
                          className={hideOnMobile ? "hidden md:block" : ""}
                        />
                      ) : null}
                    </Fragment>
                  );
                })}
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-4 p-4 pt-4 md:p-6">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
