import { Bot, ChevronLeft, Clock3, GitBranch, Maximize2, Minimize2, Minus, Plus, Shield } from "lucide-react";
import { useCallback, useEffect, useRef, useState, type WheelEvent } from "react";
import type { Theme } from "../stores/ui-store";
import type { TerminalTarget } from "../types/terminal";
import { isOrchestratorSession, workerDisplayStatus, type WorkspaceSession, type WorkerDisplayStatus } from "../types/workspace";
import { formatTimeCompact } from "../lib/format-time";
import { TerminalPane } from "./TerminalPane";
import { OrchestratorIcon } from "./icons";
import { cn } from "../lib/utils";

type CenterPaneProps = {
	session?: WorkspaceSession;
	theme: Theme;
	daemonReady: boolean;
	terminalTarget?: TerminalTarget;
	onSelectWorkerTerminal?: () => void;
};

const terminalFontSizeStorageKey = "ao.terminal.fontSize";
const DEFAULT_TERMINAL_FONT_SIZE = 12;
const MIN_TERMINAL_FONT_SIZE = 10;
const MAX_TERMINAL_FONT_SIZE = 20;
const WHEEL_ZOOM_THRESHOLD = 80;
const WHEEL_ZOOM_RESET_MS = 250;

function clampTerminalFontSize(size: number): number {
	return Math.min(MAX_TERMINAL_FONT_SIZE, Math.max(MIN_TERMINAL_FONT_SIZE, size));
}

function initialTerminalFontSize(): number {
	if (typeof window === "undefined") return DEFAULT_TERMINAL_FONT_SIZE;
	const raw = window.localStorage?.getItem(terminalFontSizeStorageKey);
	const parsed = raw === null ? Number.NaN : Number(raw);
	if (!Number.isFinite(parsed)) return DEFAULT_TERMINAL_FONT_SIZE;
	return clampTerminalFontSize(parsed);
}

export function CenterPane({ session, theme, daemonReady, terminalTarget, onSelectWorkerTerminal }: CenterPaneProps) {
	const paneRef = useRef<HTMLDivElement | null>(null);
	const wheelZoomRemainderRef = useRef(0);
	const lastWheelZoomAtRef = useRef(0);
	const [fontSize, setFontSize] = useState(initialTerminalFontSize);
	const [isFullscreen, setIsFullscreen] = useState(false);
	const target = terminalTarget ?? { kind: "worker" };

	useEffect(() => {
		const handleFullscreenChange = () => setIsFullscreen(document.fullscreenElement === paneRef.current);
		document.addEventListener("fullscreenchange", handleFullscreenChange);
		return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
	}, []);

	const updateFontSize = useCallback((delta: number) => {
		setFontSize((current) => {
			const next = clampTerminalFontSize(current + delta);
			window.localStorage?.setItem(terminalFontSizeStorageKey, String(next));
			return next;
		});
	}, []);

	const toggleFullscreen = useCallback(async () => {
		const pane = paneRef.current;
		if (!pane) return;
		try {
			if (document.fullscreenElement === pane) {
				await document.exitFullscreen();
				return;
			}
			await pane.requestFullscreen();
		} catch (error) {
			console.warn("Unable to toggle terminal fullscreen", error);
		}
	}, []);

	const handleWheelZoom = useCallback(
		(event: WheelEvent<HTMLDivElement>) => {
			if (!event.ctrlKey && !event.metaKey) return;
			event.preventDefault();
			event.stopPropagation();

			if (event.timeStamp - lastWheelZoomAtRef.current > WHEEL_ZOOM_RESET_MS) {
				wheelZoomRemainderRef.current = 0;
			}
			lastWheelZoomAtRef.current = event.timeStamp;
			wheelZoomRemainderRef.current += event.deltaY;

			const steps = Math.floor(Math.abs(wheelZoomRemainderRef.current) / WHEEL_ZOOM_THRESHOLD);
			if (steps === 0) return;

			const direction = wheelZoomRemainderRef.current > 0 ? -1 : 1;
			updateFontSize(direction * steps);
			wheelZoomRemainderRef.current -= Math.sign(wheelZoomRemainderRef.current) * steps * WHEEL_ZOOM_THRESHOLD;
		},
		[updateFontSize],
	);

	return (
		<div
			ref={paneRef}
			className="terminal-pane-frame flex h-full min-h-0 min-w-0 flex-col bg-background"
			onWheelCapture={handleWheelZoom}
		>
			<SessionWorkspaceHeader session={session} />
			<TerminalToolbar
				fontSize={fontSize}
				isFullscreen={isFullscreen}
				onDecreaseFont={() => updateFontSize(-1)}
				onIncreaseFont={() => updateFontSize(1)}
				onToggleFullscreen={() => void toggleFullscreen()}
			/>
			{target.kind === "reviewer" ? (
				<div className="reviewer-terminal-header">
					<button
						aria-label="Back to agent terminal"
						className="reviewer-terminal-header__back"
						onClick={onSelectWorkerTerminal}
						type="button"
					>
						<ChevronLeft aria-hidden="true" />
						<span>agent</span>
					</button>
					<span className="reviewer-terminal-header__role">
						<Shield aria-hidden="true" />
						Reviewer
					</span>
					<span className="reviewer-terminal-header__harness">{target.harness}</span>
				</div>
			) : null}
			<div className="min-h-0 flex-1">
				<TerminalPane
					daemonReady={daemonReady}
					fontSize={fontSize}
					session={session}
					terminalTarget={target}
					theme={theme}
				/>
			</div>
		</div>
	);
}

function SessionWorkspaceHeader({ session }: { session?: WorkspaceSession }) {
	const isOrchestrator = session ? isOrchestratorSession(session) : false;
	const displayStatus = session && !isOrchestrator ? workerDisplayStatus(session) : undefined;
	const title = !session ? "No session" : isOrchestrator ? "Orchestrator" : session.title;
	const branch = session?.branch || (session ? `session/${session.id}` : "");
	const updated = session ? formatTimeCompact(session.activity?.lastActivityAt ?? session.updatedAt) : undefined;

	return (
		<header className="session-workspace-header">
			<div className="session-workspace-header__icon" aria-hidden="true">
				{isOrchestrator ? <OrchestratorIcon className="size-4" /> : <Bot className="size-4" />}
			</div>
			<div className="session-workspace-header__copy">
				<div className="session-workspace-header__kicker">
					<span>{isOrchestrator ? "Orchestrator session" : "Agent session"}</span>
					{session?.workspaceName ? (
						<>
							<span aria-hidden="true">/</span>
							<span>{session.workspaceName}</span>
						</>
					) : null}
				</div>
				<h1 className="session-workspace-header__title">{title}</h1>
				{session ? (
					<div className="session-workspace-header__meta">
						<span className="session-workspace-header__meta-item">
							<GitBranch aria-hidden="true" />
							<span>{branch}</span>
						</span>
						<span className="session-workspace-header__meta-item">
							<Clock3 aria-hidden="true" />
							<span>{updated}</span>
						</span>
						<span className="session-workspace-header__meta-item">
							<span>{session.provider}</span>
						</span>
					</div>
				) : null}
			</div>
			{displayStatus ? <SessionHeaderStatus status={displayStatus} /> : null}
		</header>
	);
}

function SessionHeaderStatus({ status }: { status: WorkerDisplayStatus }) {
	const meta: Record<WorkerDisplayStatus, { label: string; className: string; pulse?: boolean }> = {
		working: { label: "Working", className: "session-workspace-status--working", pulse: true },
		needs_you: { label: "Needs input", className: "session-workspace-status--warning" },
		mergeable: { label: "Ready", className: "session-workspace-status--success" },
		ci_failed: { label: "CI failed", className: "session-workspace-status--error" },
		no_signal: { label: "No signal", className: "session-workspace-status--muted" },
		done: { label: "Done", className: "session-workspace-status--muted" },
		unknown: { label: "Unknown", className: "session-workspace-status--muted" },
	};
	const entry = meta[status];
	return (
		<span className={cn("session-workspace-status", entry.className)}>
			<span className={cn("session-workspace-status__dot", entry.pulse && "animate-status-pulse")} aria-hidden="true" />
			{entry.label}
		</span>
	);
}

function TerminalToolbar({
	fontSize,
	isFullscreen,
	onDecreaseFont,
	onIncreaseFont,
	onToggleFullscreen,
}: {
	fontSize: number;
	isFullscreen: boolean;
	onDecreaseFont: () => void;
	onIncreaseFont: () => void;
	onToggleFullscreen: () => void;
}) {
	return (
		<div className="terminal-toolbar">
			<div className="terminal-toolbar__label">
				<span className="terminal-toolbar__eyebrow">Terminal</span>
				<span className="terminal-toolbar__session">Live shell</span>
			</div>
			<div className="terminal-toolbar__controls">
				<button
					aria-label="Decrease terminal font size"
					className="terminal-toolbar__control"
					disabled={fontSize <= MIN_TERMINAL_FONT_SIZE}
					onClick={onDecreaseFont}
					title="Decrease terminal font size"
					type="button"
				>
					<Minus className="h-3.5 w-3.5" aria-hidden="true" />
				</button>
				<span className="terminal-toolbar__font-size">{fontSize}px</span>
				<button
					aria-label="Increase terminal font size"
					className="terminal-toolbar__control"
					disabled={fontSize >= MAX_TERMINAL_FONT_SIZE}
					onClick={onIncreaseFont}
					title="Increase terminal font size"
					type="button"
				>
					<Plus className="h-3.5 w-3.5" aria-hidden="true" />
				</button>
				<button
					aria-label={isFullscreen ? "Exit terminal fullscreen" : "Open terminal fullscreen"}
					aria-pressed={isFullscreen}
					className="terminal-toolbar__control terminal-toolbar__control--icon"
					onClick={onToggleFullscreen}
					title={isFullscreen ? "Exit fullscreen" : "Fullscreen terminal"}
					type="button"
				>
					{isFullscreen ? (
						<Minimize2 className="h-3.5 w-3.5" aria-hidden="true" />
					) : (
						<Maximize2 className="h-3.5 w-3.5" aria-hidden="true" />
					)}
				</button>
			</div>
		</div>
	);
}
