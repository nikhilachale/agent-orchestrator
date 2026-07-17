import { useEffect } from "react";
import { Info, X } from "lucide-react";
import { cn } from "../lib/utils";
import { useToastStore } from "../stores/toast-store";

const toastDurationMs = 6000;

export function GlobalToastViewport() {
	const toast = useToastStore((state) => state.toast);
	const dismissToast = useToastStore((state) => state.dismissToast);

	useEffect(() => {
		if (!toast) return;
		const timeout = window.setTimeout(() => dismissToast(toast.id), toastDurationMs);
		return () => window.clearTimeout(timeout);
	}, [dismissToast, toast]);

	if (!toast) return null;

	return (
		<div className="fixed bottom-4 right-4 z-overlay flex max-w-[min(24rem,calc(100vw-2rem))] justify-end">
			<div
				className={cn(
					"grid grid-cols-[auto_minmax(0,1fr)_auto] gap-2.5 rounded-md border border-border bg-overlay px-3 py-2.5 text-foreground shadow-lg",
				)}
				role="status"
			>
				<Info className="mt-0.5 size-icon-base text-muted-foreground" aria-hidden="true" />
				<div className="min-w-0">
					<div className="text-caption font-semibold leading-tight">{toast.title}</div>
					{toast.body ? <div className="mt-1 text-caption leading-body text-muted-foreground">{toast.body}</div> : null}
				</div>
				<button
					aria-label="Dismiss notification"
					className="grid size-control-sm place-items-center rounded-sm text-muted-foreground hover:bg-interactive-hover hover:text-foreground"
					onClick={() => dismissToast(toast.id)}
					type="button"
				>
					<X className="size-icon-xs" aria-hidden="true" />
				</button>
			</div>
		</div>
	);
}
