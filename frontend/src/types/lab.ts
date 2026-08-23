export type Difficulty = "easy" | "medium" | "hard";
export type LabStatus = "not_started" | "in_progress" | "completed";

export type LabSummary = {
  slug: string;
  name: string;
  description: string;
  category: string;
  difficulty: Difficulty;
  xp: number;
  status: LabStatus;
  xp_awarded: number;
  completed_at: string | null;
};

export type LabDetail = LabSummary & {
  objective: string;
  hints: string[];
  vulnerability_type: string | null;
};

export type ProgressSummary = {
  total_labs: number;
  completed_labs: number;
  xp_available: number;
  xp_earned: number;
};

export type ProgressRecord = {
  lab_slug: string;
  status: LabStatus;
  xp_awarded: number;
  completed_at: string | null;
};
