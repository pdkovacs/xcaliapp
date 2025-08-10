import React, { JSX, useEffect } from "react";
import { createRoot } from "react-dom/client";
import { Provider } from "react-redux";
import App from "./App";
import { store } from "./app/store";
import "./index.css";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";
import { SnackbarProvider } from "notistack";
import { ErrorBoundary } from "react-error-boundary";
import { useReporters } from "./utils/use-reporters";

const ErrorFallback = ({ error }: { error: Error; }): JSX.Element|null => {

	const { reportError } = useReporters();

	useEffect(() => {
		reportError(error.message);
	}, [error]);

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
