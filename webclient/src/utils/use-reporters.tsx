import React from "react";
import { Button } from "@mui/material";
import type { SnackbarKey } from "notistack";
import { useSnackbar } from "notistack";

export const useReporters = () => {
	const { enqueueSnackbar, closeSnackbar } = useSnackbar();

	// eslint-disable-next-line react/display-name
	const actionScaffold = (label: string, action: () => void) => (snackbarId: SnackbarKey) => (
		<Button color="inherit" onClick={() => {
			action();
			closeSnackbar(snackbarId);
		}}>
			{label}
		</Button>
	);

	const dismissAction = actionScaffold("Dismiss", () => undefined);

	const reportError = (msg: string) => {
		enqueueSnackbar(msg, {
			variant: "error",
			persist: true,
			action: dismissAction
		});
	};

	const reportInfo = (msg: string, persist = false) => {
		enqueueSnackbar(msg, {
			variant: "info",
			persist,
			action: dismissAction
		});
	};

	const reportNotification = (msg: string, label: string, action: () => void) => enqueueSnackbar(msg, {
		variant: "info",
		persist: true,
		action: actionScaffold(label, action)
	});

	return {
		reportError,
		reportInfo,
		reportNotification
	};

};
