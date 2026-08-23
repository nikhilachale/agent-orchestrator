import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ProviderQuota } from "../../hooks/useProviderQuota";

const hookState = vi.hoisted(() => ({
	providers: [] as ProviderQuota[],
	refreshAll: vi.fn(),
	refreshProvider: vi.fn(),
	refreshProviderError: null as Error | null,
	refreshProviderPending: false,
}));

vi.mock("../../hooks/useProviderQuota", () => ({
	useProviderQuota: () => ({
		data: hookState.providers,
		error: null,
		isError: false,
		isLoading: false,
		isSuccess: true,
	}),
	useRefreshAllProviderQuota: () => ({ mutate: hookState.refreshAll }),
	useRefreshProviderQuota: () => ({
		error: hookState.refreshProviderError,
		isError: hookState.refreshProviderError != null,
		isPending: hookState.refreshProviderPending,
		mutate: hookState.refreshProvider,
	}),
}));

vi.mock("../CenterPanelShell", () => ({
	CenterPanelShell: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const { PlanUsagePage } = await import("./PlanUsagePage");

function quota(overrides: Partial<ProviderQuota> = {}): ProviderQuota {
	return {
		accountId: "default",
		balances: [],
		capabilities: {
			supportsCredits: false,
			supportsHistory: false,
			supportsRead: false,
			supportsSpendLimits: false,
			supportsSubscribe: true,
		},
		completeness: "partial",
		freshness: "fresh",
		limits: [],
		observedAt: new Date().toISOString(),
		provider: "claude",
		severity: "normal",
		...overrides,
	};
}

describe("PlanUsagePage", () => {
	beforeEach(() => {
		hookState.providers = [];
		hookState.refreshAll.mockClear();
		hookState.refreshProvider.mockClear();
		hookState.refreshProviderError = null;
		hookState.refreshProviderPending = false;
	});

	it("shows an actionable empty state before providers report quota", () => {
		render(<PlanUsagePage />);

		expect(screen.getByText("No provider quota observed yet")).toBeInTheDocument();
		expect(screen.getByText(/checking connected providers/i)).toBeInTheDocument();
		expect(hookState.refreshAll).toHaveBeenCalledOnce();
	});

	it("renders Codex and Claude through the same provider-neutral card", () => {
		hookState.providers = [
			quota({
				accountLabel: "Codex Team",
				balances: [{ id: "codex:credits", name: "Codex credits", unlimited: false, value: "50" }],
				capabilities: {
					supportsCredits: true,
					supportsHistory: true,
					supportsRead: true,
					supportsSpendLimits: true,
					supportsSubscribe: true,
				},
				completeness: "complete",
				limits: [{
					category: "requests",
					id: "primary",
					remainingPercent: 8,
					scope: "account",
					severity: "critical",
					usedPercent: 92,
					windowDurationSeconds: 18_000,
					windowType: "rolling",
				}],
				provider: "codex",
				severity: "critical",
			}),
			quota({
				accountLabel: "Claude Pro",
				limits: [{
					category: "requests",
					id: "five_hour",
					remainingPercent: 72,
					scope: "account",
					severity: "normal",
					usedPercent: 28,
					windowDurationSeconds: 18_000,
					windowType: "rolling",
				}],
			}),
		];

		render(<PlanUsagePage />);

		expect(screen.getByText("Codex Team")).toBeInTheDocument();
		expect(screen.getByText("Claude Pro")).toBeInTheDocument();
		expect(screen.getByText("8% remaining")).toHaveClass("text-status-exited");
		expect(screen.getByText("72% remaining")).toBeInTheDocument();
		expect(screen.queryByRole("button", { name: /refresh/i })).not.toBeInTheDocument();
		expect(screen.queryByText("Observed usage history")).not.toBeInTheDocument();
	});

	it("renders reported balances and retries failed account refreshes", () => {
		hookState.providers = [quota({
			accountLabel: "Codex Team",
			balances: [
				{ id: "reset-credits", name: "Reset credits", unlimited: false, value: "2" },
				{ id: "codex:credits", name: "Codex credits", unlimited: true },
			],
			capabilities: {
				supportsCredits: true,
				supportsHistory: false,
				supportsRead: true,
				supportsSpendLimits: true,
				supportsSubscribe: true,
			},
			completeness: "complete",
			provider: "codex",
			refreshError: "provider timed out",
		})];

		render(<PlanUsagePage />);

		expect(screen.getByText("Credits and balances")).toBeInTheDocument();
		expect(screen.getByText("Reset credits")).toBeInTheDocument();
		expect(screen.getByText("2")).toBeInTheDocument();
		expect(screen.getByText("Codex credits")).toBeInTheDocument();
		expect(screen.getByText("Unlimited")).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Retry Codex Team usage refresh" }));

		expect(hookState.refreshProvider).toHaveBeenCalledOnce();
	});

	it("renders an unknown future provider without frontend adapter code", () => {
		hookState.providers = [quota({ provider: "future-ai", accountLabel: undefined })];

		render(<PlanUsagePage />);

		expect(screen.getByText("Future Ai")).toBeInTheDocument();
	});

	it("renders absolute and non-numeric spend states without percentage bars", () => {
		hookState.providers = [quota({
			accountLabel: "Cursor",
			limits: [
				{ category: "spend_limit", id: "fixed", name: "On-Demand", scope: "account", severity: "exhausted", state: "active", usedValue: 333.68, totalValue: 1, remainingValue: 0, unit: "USD" },
				{ category: "spend_limit", id: "unlimited", scope: "account", severity: "normal", state: "unlimited", unit: "USD" },
				{ category: "spend_limit", id: "disabled", scope: "account", severity: "unknown", state: "disabled", unit: "USD" },
				{ category: "spend_limit", id: "unavailable", scope: "account", severity: "unknown", state: "unavailable", unit: "USD" },
			],
			provider: "cursor",
		})];

		render(<PlanUsagePage />);

		expect(screen.getByText("$333.68 / $1.00")).toBeInTheDocument();
		expect(screen.getByText("$0.00 remaining")).toBeInTheDocument();
		expect(screen.getAllByText("Unlimited")).toHaveLength(2);
		expect(screen.getAllByText("Disabled")).toHaveLength(2);
		expect(screen.getAllByText("Unavailable")).toHaveLength(2);
		expect(screen.queryByRole("progressbar")).not.toBeInTheDocument();
	});
});
