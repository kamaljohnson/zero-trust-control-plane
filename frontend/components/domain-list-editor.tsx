"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

/**
 * Parses bulk domain text into a list of non-empty trimmed entries.
 * Splits on newlines and commas; filters empty and trims each part.
 *
 * @param text - Raw input (e.g. pasted list, one per line or comma-separated)
 * @returns Array of non-empty trimmed strings
 */
function parseBulkDomains(text: string): string[] {
  return text
    .split(/[\n,]+/)
    .map((s) => s.trim())
    .filter(Boolean);
}

export interface DomainListEditorProps {
  /** Current list of domains. */
  value: string[];
  /** Called when the list changes. */
  onChange: (domains: string[]) => void;
  /** Label shown above the list (e.g. "Allowed domains"). */
  label: string;
  /** Optional aria-label for the list container. */
  "aria-label"?: string;
  /** Optional class name for the root element. */
  className?: string;
}

/**
 * Renders an editable list of domains with per-item remove, single-domain add,
 * and bulk "Add from list" (paste multiple lines). Deduplication is
 * case-sensitive; new entries already present in the list are skipped.
 */
export function DomainListEditor({
  value,
  onChange,
  label,
  "aria-label": ariaLabel,
  className,
}: DomainListEditorProps) {
  const [singleInput, setSingleInput] = useState("");
  const [bulkInput, setBulkInput] = useState("");
  const list = value ?? [];
  const set = new Set(list);

  const addOne = () => {
    const domain = singleInput.trim();
    if (!domain || set.has(domain)) return;
    onChange([...list, domain]);
    setSingleInput("");
  };

  const addFromBulk = () => {
    const parsed = parseBulkDomains(bulkInput);
    const newDomains = parsed.filter((d) => !set.has(d));
    if (newDomains.length === 0) {
      setBulkInput("");
      return;
    }
    onChange([...list, ...newDomains]);
    setBulkInput("");
  };

  const remove = (domain: string) => {
    onChange(list.filter((d) => d !== domain));
  };

  return (
    <div className={cn("space-y-3", className)} aria-label={ariaLabel}>
      <Label>{label}</Label>

      {list.length === 0 ? (
        <p className="text-sm text-muted-foreground">No domains added yet. Add domains below.</p>
      ) : (
        <ul className="flex flex-wrap gap-2 list-none p-0 m-0">
          {list.map((domain) => (
            <li key={domain} className="inline-flex items-center gap-1 rounded-md border border-input bg-muted/50 px-2 py-1 text-sm font-mono">
              <span>{domain}</span>
              <button
                type="button"
                onClick={() => remove(domain)}
                className="rounded p-0.5 hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                aria-label={`Remove ${domain}`}
              >
                ×
              </button>
            </li>
          ))}
        </ul>
      )}

      <div className="flex gap-2">
        <Input
          type="text"
          placeholder="e.g. example.com"
          value={singleInput}
          onChange={(e) => setSingleInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), addOne())}
          className="flex-1 max-w-xs font-mono text-sm"
        />
        <Button type="button" variant="secondary" size="default" onClick={addOne}>
          Add
        </Button>
      </div>

      <div className="space-y-2">
        <textarea
          className="w-full min-h-[80px] rounded-md border border-input bg-background px-3 py-2 font-mono text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          placeholder="Paste multiple domains (one per line or comma-separated)"
          value={bulkInput}
          onChange={(e) => setBulkInput(e.target.value)}
        />
        <Button type="button" variant="outline" size="sm" onClick={addFromBulk}>
          Add from list
        </Button>
      </div>
    </div>
  );
}
