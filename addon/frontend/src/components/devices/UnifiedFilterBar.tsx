import { type RefObject, useMemo } from "react";
import {
  CheckCircle2,
  ChevronDown,
  Circle,
  CircleDashed,
  Filter,
  ListFilter,
  Save,
  Search,
  Wifi,
  WifiOff,
  XCircle,
  X
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList
} from "@/components/ui/command";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import type { SavedView } from "@/hooks/useSavedViews";
import type { OnlineScope, RegistrationScope } from "@/lib/device-semantics";

type FacetKind = "vendors" | "sources" | "subnets";
type ChipKind = "query" | "segmentation" | "online" | FacetKind;

type Props = {
  isMobile: boolean;
  searchInputRef: RefObject<HTMLInputElement>;
  query: string;
  segmentation: RegistrationScope;
  onlineScope: OnlineScope;
  facets: {
    vendors: string[];
    sources: string[];
    subnets: string[];
  };
  options: {
    vendors: string[];
    sources: string[];
    subnets: string[];
  };
  savedViews: SavedView[];
  onQueryChange: (value: string) => void;
  onClearQuery: () => void;
  onSegmentationChange: (value: RegistrationScope) => void;
  onOnlineScopeChange: (value: OnlineScope) => void;
  onToggleFacet: (kind: FacetKind, value: string) => void;
  onClearAll: () => void;
  onSaveCurrentView: () => void;
  onApplyView: (id: string) => void;
  onDeleteView: (id: string) => void;
};

function FacetPopover({
  label,
  values,
  selected,
  onToggle
}: {
  label: string;
  values: string[];
  selected: string[];
  onToggle: (value: string) => void;
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={`h-9 gap-2 ${selected.length > 0 ? "bg-accent text-accent-foreground" : ""}`}
        >
          <Filter className="h-4 w-4" />
          {label}
          <Badge variant="secondary" className="h-5 min-w-5 justify-center px-1 text-xs">
            {selected.length}
          </Badge>
          <ChevronDown className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput placeholder={`Filter ${label.toLowerCase()}...`} />
          <CommandList>
            <CommandEmpty>No matches.</CommandEmpty>
            <CommandGroup>
              {values.map((value) => {
                const active = selected.includes(value);
                return (
                  <CommandItem
                    key={value}
                    value={value}
                    onSelect={() => onToggle(value)}
                    className="flex items-center justify-between gap-2"
                  >
                    <span className="truncate">{value}</span>
                    {active ? (
                      <Badge variant="secondary" className="text-xs">
                        Active
                      </Badge>
                    ) : null}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

function SavedViewsMenu({
  views,
  onSaveCurrentView,
  onApplyView,
  onDeleteView
}: {
  views: SavedView[];
  onSaveCurrentView: () => void;
  onApplyView: (id: string) => void;
  onDeleteView: (id: string) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-9 gap-2">
          <Save className="h-4 w-4" />
          Saved views
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={onSaveCurrentView}>
          <Save className="mr-2 h-4 w-4" />
          Save current view
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        {views.length === 0 ? <DropdownMenuItem disabled>No saved views</DropdownMenuItem> : null}
        {views.map((view) => (
          <DropdownMenuItem key={`apply-${view.id}`} onClick={() => onApplyView(view.id)}>
            {view.name}
          </DropdownMenuItem>
        ))}
        {views.length > 0 ? <DropdownMenuSeparator /> : null}
        {views.map((view) => (
          <DropdownMenuItem
            key={`delete-${view.id}`}
            onClick={() => onDeleteView(view.id)}
            className="text-destructive focus:text-destructive"
          >
            Delete {view.name}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function SegmentationMenu({
  value,
  onChange
}: {
  value: RegistrationScope;
  onChange: (value: RegistrationScope) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-9 gap-2">
          <ListFilter className="h-4 w-4" />
          <span>Status: {segmentationLabel(value)}</span>
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        <DropdownMenuRadioGroup
          value={value}
          onValueChange={(next) => {
            if (
              next === "all" ||
              next === "new" ||
              next === "registered" ||
              next === "unregistered"
            ) {
              onChange(next);
            }
          }}
        >
          <DropdownMenuRadioItem value="all" className="gap-2">
            <Circle className="h-4 w-4" />
            All
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="new" className="gap-2">
            <CircleDashed className="h-4 w-4" />
            New
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="registered" className="gap-2">
            <CheckCircle2 className="h-4 w-4" />
            Registered
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="unregistered" className="gap-2">
            <XCircle className="h-4 w-4" />
            Unregistered
          </DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function OnlineScopeMenu({
  value,
  onChange
}: {
  value: OnlineScope;
  onChange: (value: OnlineScope) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-9 gap-2">
          <Wifi className="h-4 w-4" />
          <span>Connection: {onlineLabel(value)}</span>
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        <DropdownMenuRadioGroup
          value={value}
          onValueChange={(next) => {
            if (next === "any" || next === "online" || next === "offline") {
              onChange(next);
            }
          }}
        >
          <DropdownMenuRadioItem value="any" className="gap-2">
            <Circle className="h-4 w-4" />
            Any
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="online" className="gap-2">
            <Wifi className="h-4 w-4" />
            Online
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="offline" className="gap-2">
            <WifiOff className="h-4 w-4" />
            Offline
          </DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function segmentationLabel(value: RegistrationScope) {
  if (value === "all") {
    return "All";
  }
  if (value === "new") {
    return "New";
  }
  if (value === "registered") {
    return "Registered";
  }
  return "Unregistered";
}

function onlineLabel(value: OnlineScope) {
  if (value === "any") {
    return "Any";
  }
  if (value === "online") {
    return "Online";
  }
  return "Offline";
}

export function UnifiedFilterBar({
  isMobile,
  searchInputRef,
  query,
  segmentation,
  onlineScope,
  facets,
  options,
  savedViews,
  onQueryChange,
  onClearQuery,
  onSegmentationChange,
  onOnlineScopeChange,
  onToggleFacet,
  onClearAll,
  onSaveCurrentView,
  onApplyView,
  onDeleteView
}: Props) {
  const chips = useMemo(() => {
    const items: Array<{ kind: ChipKind; value: string; label: string }> = [];
    const normalizedQuery = query.trim();
    if (normalizedQuery !== "") {
      items.push({ kind: "query", value: normalizedQuery, label: `Search: ${normalizedQuery}` });
    }
    if (segmentation !== "all") {
      items.push({
        kind: "segmentation",
        value: segmentation,
        label: `Type: ${segmentationLabel(segmentation)}`
      });
    }
    if (onlineScope !== "any") {
      items.push({
        kind: "online",
        value: onlineScope,
        label: `Status: ${onlineLabel(onlineScope)}`
      });
    }
    for (const vendor of facets.vendors) {
      items.push({ kind: "vendors", value: vendor, label: `Vendor: ${vendor}` });
    }
    for (const source of facets.sources) {
      items.push({ kind: "sources", value: source, label: `Source: ${source}` });
    }
    for (const subnet of facets.subnets) {
      items.push({ kind: "subnets", value: subnet, label: `Subnet: ${subnet}` });
    }
    return items;
  }, [facets.sources, facets.subnets, facets.vendors, onlineScope, query, segmentation]);

  const handleRemoveChip = (chip: { kind: ChipKind; value: string }) => {
    if (chip.kind === "query") {
      onClearQuery();
      return;
    }
    if (chip.kind === "segmentation") {
      onSegmentationChange("all");
      return;
    }
    if (chip.kind === "online") {
      onOnlineScopeChange("any");
      return;
    }
    onToggleFacet(chip.kind, chip.value);
  };

  return (
    <section className="space-y-3 py-4">
      <div className="flex flex-col gap-3">
        <div className="relative w-full">
          <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            ref={searchInputRef}
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search devices by name, MAC, vendor, IP"
            className="h-9 pr-9 pl-8 text-sm"
          />
          {query.trim() !== "" ? (
            <Button
              variant="ghost"
              size="sm"
              className="absolute right-0 top-0 h-9 w-9 p-0"
              onClick={onClearQuery}
              aria-label="Clear search"
            >
              <X className="h-4 w-4" />
            </Button>
          ) : null}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <SegmentationMenu value={segmentation} onChange={onSegmentationChange} />
        <OnlineScopeMenu value={onlineScope} onChange={onOnlineScopeChange} />

        <Separator orientation="vertical" className="hidden h-6 opacity-40 md:block" />

        {isMobile ? (
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="sm" className="h-9 gap-2">
                <Filter className="h-4 w-4" />
                Filters
                <Badge variant="secondary" className="h-5 min-w-5 justify-center px-1 text-xs">
                  {facets.vendors.length + facets.sources.length + facets.subnets.length}
                </Badge>
              </Button>
            </PopoverTrigger>
            <PopoverContent align="start" className="w-80 space-y-3 p-3">
              <div className="flex flex-wrap items-center gap-2">
                <FacetPopover
                  label="Vendor"
                  values={options.vendors}
                  selected={facets.vendors}
                  onToggle={(value) => onToggleFacet("vendors", value)}
                />
                <FacetPopover
                  label="Source"
                  values={options.sources}
                  selected={facets.sources}
                  onToggle={(value) => onToggleFacet("sources", value)}
                />
                <FacetPopover
                  label="Subnet"
                  values={options.subnets}
                  selected={facets.subnets}
                  onToggle={(value) => onToggleFacet("subnets", value)}
                />
              </div>
            </PopoverContent>
          </Popover>
        ) : (
          <>
            <div className="flex flex-wrap items-center gap-2">
              <FacetPopover
                label="Vendor"
                values={options.vendors}
                selected={facets.vendors}
                onToggle={(value) => onToggleFacet("vendors", value)}
              />
              <FacetPopover
                label="Source"
                values={options.sources}
                selected={facets.sources}
                onToggle={(value) => onToggleFacet("sources", value)}
              />
              <FacetPopover
                label="Subnet"
                values={options.subnets}
                selected={facets.subnets}
                onToggle={(value) => onToggleFacet("subnets", value)}
              />
            </div>
          </>
        )}

        <Separator orientation="vertical" className="hidden h-6 opacity-40 md:block" />

        <SavedViewsMenu
          views={savedViews}
          onSaveCurrentView={onSaveCurrentView}
          onApplyView={onApplyView}
          onDeleteView={onDeleteView}
        />
      </div>

      {chips.length > 0 ? (
        <div className="flex flex-wrap items-center gap-2">
          {chips.map((chip) => (
            <Badge key={`${chip.kind}-${chip.value}`} variant="secondary" className="h-7 gap-1 pr-1 text-xs">
              {chip.label}
              <Button
                variant="ghost"
                size="sm"
                className="h-5 w-5 p-0"
                onClick={() => handleRemoveChip({ kind: chip.kind, value: chip.value })}
                aria-label={`Remove ${chip.label}`}
              >
                <X className="h-4 w-4" />
              </Button>
            </Badge>
          ))}

          <Button
            variant="outline"
            size="sm"
            className="h-7 rounded-full gap-1 px-2 text-xs"
            onClick={onClearAll}
            aria-label="Clear all filters"
          >
            Clear all
            <X className="h-4 w-4" />
          </Button>
        </div>
      ) : null}
    </section>
  );
}
