import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { ApiError } from "../daemon/api";
import { PairScreen } from "./PairScreen";

function setup(pair = vi.fn(() => Promise.resolve("device-token"))) {
  const onPaired = vi.fn();
  render(<PairScreen api={{ pair }} onPaired={onPaired} />);
  return { pair, onPaired, token: screen.getByLabelText("Pairing token") as HTMLInputElement };
}

test("the pairing token is hidden while typed or pasted, and can be shown on request", () => {
  const { token } = setup();
  expect(token.type).toBe("password");
  fireEvent.click(screen.getByRole("button", { name: "Show token" }));
  expect(token.type).toBe("text");
  fireEvent.click(screen.getByRole("button", { name: "Hide token" }));
  expect(token.type).toBe("password");
});

test("pairing is not possible without a token", () => {
  setup();
  expect(screen.getByRole("button", { name: "Pair device" }).hasAttribute("disabled")).toBe(true);
});

test("a token and a device name are traded for the device token, and nothing stays in the field", async () => {
  const { pair, onPaired, token } = setup();
  fireEvent.change(token, { target: { value: "  pairing-secret  " } });
  fireEvent.change(screen.getByLabelText("Device name"), { target: { value: "Maulana's laptop" } });
  fireEvent.click(screen.getByRole("button", { name: "Pair device" }));
  await waitFor(() => expect(onPaired).toHaveBeenCalledWith("device-token"));
  expect(pair).toHaveBeenCalledWith("pairing-secret", "Maulana's laptop");
  expect(token.value).toBe("");
});

test("each way pairing can fail says what to do next", async () => {
  for (const [error, expected] of [
    [new ApiError(401, "unauthorized", "x"), /expired or been used already: run astrad pair/],
    [new ApiError(0, "unreachable", "x"), /Check that astrad serve is running/],
    [new ApiError(400, "invalid", "a device name is required"), /a device name is required/],
    [new Error("boom"), /Pairing failed/],
  ] as const) {
    const { unmount } = render(<PairScreen api={{ pair: () => Promise.reject(error) }} onPaired={vi.fn()} />);
    const fields = screen.getAllByLabelText("Pairing token");
    fireEvent.change(fields[fields.length - 1] as HTMLElement, { target: { value: "t" } });
    fireEvent.click(screen.getAllByRole("button", { name: "Pair device" }).at(-1) as HTMLElement);
    expect((await screen.findByRole("alert")).textContent).toMatch(expected);
    unmount();
  }
});

test("a notice explains why the screen came back", () => {
  render(<PairScreen api={{ pair: vi.fn() }} notice="This device was disconnected from the daemon." onPaired={vi.fn()} />);
  expect(screen.getByText("This device was disconnected from the daemon.")).toBeTruthy();
  expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
});
