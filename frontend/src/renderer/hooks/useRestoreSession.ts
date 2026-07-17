import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { apiClient, apiErrorMessage } from "../lib/api-client";
import { showGlobalToast } from "../stores/toast-store";
import { workspaceQueryKey } from "./useWorkspaceQuery";

export type RestoreSessionResult =
	| { status: "success"; restoreMode?: "native_resume" | "prompt_replay" | "fresh_launch" }
	| { status: "not_resumable" }
	| { status: "error"; message: string };

export function useRestoreSession(): (sessionId: string) => Promise<RestoreSessionResult> {
	const queryClient = useQueryClient();

	return useCallback(
		async (sessionId: string) => {
			try {
				const { data, error } = await apiClient.POST("/api/v1/sessions/{sessionId}/restore", {
					params: { path: { sessionId } },
				});
				if (error) {
					const code = (error as { code?: string }).code;
					if (code === "SESSION_NOT_RESUMABLE") {
						return { status: "not_resumable" };
					}
					return { status: "error", message: apiErrorMessage(error, "Unable to restore session") };
				}
				await queryClient.invalidateQueries({ queryKey: workspaceQueryKey });
				const restoreMode = data?.restoreMode;
				if (restoreMode === "prompt_replay") {
					showGlobalToast({
						title: "Started a new conversation",
						body: "AO could not find the native session to resume, so it restored from the saved prompt.",
					});
				}
				return { status: "success", restoreMode };
			} catch (err) {
				return {
					status: "error",
					message: err instanceof Error ? err.message : "Unable to restore session",
				};
			}
		},
		[queryClient],
	);
}
