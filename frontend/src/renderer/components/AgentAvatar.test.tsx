import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AgentAvatar } from "./AgentAvatar";

describe("AgentAvatar", () => {
	it("renders the Prime Agent brand asset", () => {
		render(<AgentAvatar provider="prime-agent" />);

		expect(screen.getByRole("img", { name: "prime-agent" })).toHaveAttribute(
			"src",
			expect.stringContaining("prime-agent.png"),
		);
	});

	it("renders the Amp brand asset", () => {
		render(<AgentAvatar provider="amp" />);

		expect(screen.getByRole("img", { name: "amp" })).toHaveAttribute("src", expect.stringContaining("amp.png"));
	});

	it("renders the OMP brand asset", () => {
		render(<AgentAvatar provider="omp" />);

		expect(screen.getByRole("img", { name: "omp" })).toHaveAttribute("src", expect.stringContaining("omp.png"));
	});

	it("renders the DeepSeek Harness brand asset", () => {
		render(<AgentAvatar provider="deepseek-harness" />);

		const img = screen.queryByRole("img", { name: "deepseek-harness" });
		if (img?.tagName.toLowerCase() === "img") {
			expect(img).toHaveAttribute("src", expect.stringContaining("data:image/svg+xml"));
		} else {
			const el = screen.getByLabelText("deepseek-harness");
			expect(
				["svg", "img"].includes(el.tagName.toLowerCase()) || el.getAttribute("role") === "img",
			).toBe(true);
		}
	});
});
