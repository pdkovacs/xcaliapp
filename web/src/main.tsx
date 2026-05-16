import React, { JSX, useEffect } from "react";
import { createRoot } from "react-dom/client";
import { Provider } from "react-redux";
import { Button } from "@mui/material";
import App from "./App";
import { store } from "./app/store";
import "./index.css";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";
import { closeSnackbar, SnackbarProvider, type SnackbarKey } from "notistack";
import { ErrorBoundary, FallbackProps } from "react-error-boundary";
import { showNotification } from "./features/notifications/notifications";
import { useAppDispatch } from "./app/hooks";

const dismissAction = (snackbarId: SnackbarKey) => (
	<Button color="inherit" onClick={() => closeSnackbar(snackbarId)}>
		Dismiss
	</Button>
);

const ErrorFallback = ({ error }: FallbackProps): JSX.Element | null => {
	const dispatch = useAppDispatch();

	useEffect(() => {
		const message = error instanceof Error ? error.message : String(error);
		dispatch(showNotification({ message, variant: "error", persist: true }));
	}, [error, dispatch]);

	return null;
};

const container = document.getElementById("root");

if (container) {
	const root = createRoot(container);

	root.render(
		<React.StrictMode>
			<Provider store={store}>
				<SnackbarProvider
					maxSnack={3}
					action={dismissAction}
				>
					<ErrorBoundary
						FallbackComponent={ErrorFallback}
						onReset={() => undefined}
					>
						<App />
					</ErrorBoundary>
				</SnackbarProvider>
			</Provider>
		</React.StrictMode>
	);
} else {
	throw new Error(
		"Root element with ID 'root' was not found in the document. Ensure there is a corresponding HTML element with the ID 'root' in your HTML file."
	);
}
