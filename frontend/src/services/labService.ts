import { apiFetch } from "../lib/api";
import type { LabDetail, LabSummary, ProgressRecord, ProgressSummary } from "../types/lab";

export async function listLabs(): Promise<LabSummary[]> {
  const res = await apiFetch<{ labs: LabSummary[] }>("/labs");
  return res.labs;
}

export async function getLab(slug: string): Promise<LabDetail> {
  const res = await apiFetch<{ lab: LabDetail }>(`/labs/${encodeURIComponent(slug)}`);
  return res.lab;
}

export async function getProgress(): Promise<{ summary: ProgressSummary; records: ProgressRecord[] }> {
  return apiFetch("/progress");
}
