import { createAction, createListenerMiddleware, isRejected, type Action } from "@reduxjs/toolkit";
import { enqueueSnackbar } from "notistack";

type ListenerMiddleware = ReturnType<typeof createListenerMiddleware>;

export type NotificationPayload = {
	message: string;
	variant?: "error" | "info" | "success" | "warning";
	persist?: boolean;
};

export type ErrorLabelResolver = (action: Action) => string | undefined;

export const showNotification = createAction<NotificationPayload>("notifications/show");

export const registerNotificationListeners = (
	middleware: ListenerMiddleware,
	resolveErrorLabel: ErrorLabelResolver
) => {
	middleware.startListening({
		matcher: isRejected,
		effect: (action: Action, api) => {
			const meta = (action as { meta?: { aborted?: boolean; condition?: boolean } }).meta;
			if (meta?.aborted || meta?.condition) {
				return;
			}
			const message = resolveErrorLabel(action);
			if (!message) {
				return;
			}
			api.dispatch(showNotification({ message, variant: "error", persist: true }));
		}
	});

	middleware.startListening({
		actionCreator: showNotification,
		effect: action => {
			const { message, variant = "info", persist = false } = action.payload;
			enqueueSnackbar(message, { variant, persist });
		}
	});
};
