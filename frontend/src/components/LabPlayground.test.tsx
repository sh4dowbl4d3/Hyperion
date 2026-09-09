import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { LabPlayground } from "./LabPlayground";
import * as api from "../lib/api";

describe("LabPlayground", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it("calls apiFetch with the lab slug in the target path for GET endpoints", async () => {
    const fetchSpy = vi.spyOn(api, "apiFetch").mockResolvedValue({ contacts: [] });
    const onComplete = vi.fn();

    render(<LabPlayground slug="sqli" onComplete={onComplete} />);

    const inputs = screen.getAllByPlaceholderText("ali");
    fireEvent.change(inputs[0], { target: { value: "' OR true --" } });

    const sendButtons = screen.getAllByRole("button", { name: /send/i });
    fireEvent.click(sendButtons[0]);

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith(
        "/targets/sqli/search?q='%20OR%20true%20--",
        expect.objectContaining({ method: "GET" }),
      );
    });

    expect(onComplete).toHaveBeenCalled();
  });

  it("sends object body for POST endpoints and excludes body params from query string", async () => {
    const fetchSpy = vi.spyOn(api, "apiFetch").mockResolvedValue({ comment: { id: 1 } });
    const onComplete = vi.fn();

    render(<LabPlayground slug="xss" onComplete={onComplete} />);

    const commentInput = screen.getByPlaceholderText("<script>FLAG-XSS-77b1e4</script>");
    // The author input is the preceding input
    const inputs = screen.getAllByRole("textbox");
    fireEvent.change(inputs[0], { target: { value: "Attacker" } });
    fireEvent.change(commentInput, { target: { value: "<script>alert(1)</script>" } });

    const sendButtons = screen.getAllByRole("button", { name: /send/i });
    fireEvent.click(sendButtons[0]);

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith(
        "/targets/xss/comments",
        expect.objectContaining({
          method: "POST",
          body: {
            author: "Attacker",
            body: "<script>alert(1)</script>",
          },
        }),
      );
    });

    expect(onComplete).toHaveBeenCalled();
  });

  it("passes X-Lab-Badge header for JWT endpoint without putting badge into URL", async () => {
    const fetchSpy = vi.spyOn(api, "apiFetch").mockResolvedValue({ panel: "restricted" });

    render(<LabPlayground slug="jwt" />);

    const badgeInput = screen.getByPlaceholderText("eyJhbGciOiJub25lIn0...");
    fireEvent.change(badgeInput, { target: { value: "token.abc.123" } });

    // Click Send for the admin panel vulnerable endpoint (index 1)
    const sendButtons = screen.getAllByRole("button", { name: /send/i });
    fireEvent.click(sendButtons[1]);

    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith(
        "/targets/jwt/admin-panel",
        expect.objectContaining({
          method: "GET",
          headers: {
            "X-Lab-Badge": "Bearer token.abc.123",
          },
        }),
      );
    });
  });
});
