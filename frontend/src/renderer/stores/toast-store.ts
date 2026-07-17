import { create } from "zustand";

export type GlobalToast = {
	id: number;
	title: string;
	body?: string;
};

type ToastState = {
	toast: GlobalToast | null;
	showToast: (toast: Omit<GlobalToast, "id">) => void;
	dismissToast: (id: number) => void;
};

let nextToastId = 1;

export const useToastStore = create<ToastState>((set) => ({
	toast: null,
	showToast: (toast) => set({ toast: { ...toast, id: nextToastId++ } }),
	dismissToast: (id) =>
		set((state) => {
			if (state.toast?.id !== id) return state;
			return { toast: null };
		}),
}));

export function showGlobalToast(toast: Omit<GlobalToast, "id">) {
	useToastStore.getState().showToast(toast);
}
