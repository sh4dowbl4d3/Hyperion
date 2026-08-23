export type PlaygroundParam = {
  name: string;
  label: string;
  placeholder?: string;
};

export type PlaygroundEndpoint = {
  /** Path under /api/v1/targets/<slug>/ */
  path: string;
  method: "GET" | "POST";
  label: string;
  description: string;
  params?: PlaygroundParam[];
  /** JSON body template; %PARAM:name% is substituted. */
  bodyTemplate?: Record<string, string>;
};

/**
 * Per-lab playground presets. Each preset drives the lab's vulnerable
 * endpoint (and its safe twin where one exists) without needing curl.
 * All requests go through the authenticated apiFetch client.
 */
export const PLAYGROUNDS: Record<string, PlaygroundEndpoint[]> = {
  sqli: [
    {
      path: "search",
      method: "GET",
      label: "Vulnerable search",
      description: "q is concatenated straight into the SQL text.",
      params: [{ name: "q", label: "Search term", placeholder: "ali" }],
    },
    {
      path: "search-safe",
      method: "GET",
      label: "Safe twin",
      description: "q is bound as a parameter — injection is inert.",
      params: [{ name: "q", label: "Search term", placeholder: "' OR true --" }],
    },
  ],
  xss: [
    {
      path: "comments",
      method: "POST",
      label: "Post comment",
      description: "The vulnerable wall stores bodies verbatim.",
      params: [
        { name: "author", label: "Author" },
        { name: "body", label: "Comment", placeholder: "<script>FLAG-XSS-77b1e4</script>" },
      ],
      bodyTemplate: { author: "%PARAM:author%", body: "%PARAM:body%" },
    },
    {
      path: "comments/feed",
      method: "GET",
      label: "View wall (raw HTML)",
      description: "Serves stored comments as markup — payloads execute here.",
    },
    {
      path: "comments-safe/feed",
      method: "GET",
      label: "View safe wall",
      description: "Same page, escaped rendering.",
    },
  ],
  idor: [
    {
      path: "documents",
      method: "GET",
      label: "My documents",
      description: "Only your own files are listed here.",
    },
    {
      path: "documents/:id",
      method: "GET",
      label: "Read document (vulnerable)",
      description: "No ownership check — try IDs outside your range.",
      params: [{ name: "id", label: "Document ID", placeholder: "3" }],
    },
    {
      path: "documents-safe/:id",
      method: "GET",
      label: "Read document (safe)",
      description: "Cross-tenant reads return not-found.",
      params: [{ name: "id", label: "Document ID", placeholder: "3" }],
    },
  ],
  jwt: [
    {
      path: "guest-token",
      method: "GET",
      label: "Mint guest badge",
      description: "Issues a low-privilege badge signed with a weak secret.",
    },
    {
      path: "admin-panel",
      method: "GET",
      label: "Admin panel (vulnerable)",
      description: "Requires X-Lab-Badge. Forge one with alg:none or the weak secret, then paste it here.",
      params: [{ name: "badge", label: "Badge token", placeholder: "eyJhbGciOiJub25lIn0..." }],
    },
    {
      path: "admin-panel-safe",
      method: "GET",
      label: "Admin panel (safe)",
      description: "Pinned algorithm + strong secret rejects every forgery.",
      params: [{ name: "badge", label: "Badge token" }],
    },
  ],
  ssrf: [
    {
      path: "config",
      method: "GET",
      label: "Webhook config",
      description: "Reveals the internal service base URL — start here.",
    },
    {
      path: "preview",
      method: "GET",
      label: "URL preview (vulnerable)",
      description: "Fetches any URL from the server's network position.",
      params: [{ name: "url", label: "Target URL", placeholder: "http://127.0.0.1:PORT/secret/credentials.txt" }],
    },
    {
      path: "preview-safe",
      method: "GET",
      label: "URL preview (safe)",
      description: "Loopback/private/link-local targets are rejected.",
      params: [{ name: "url", label: "Target URL" }],
    },
  ],
};

/** The race lab gets a dedicated helper instead of a plain form. */
export function hasRaceHelper(slug: string): boolean {
  return slug === "race-condition";
}
