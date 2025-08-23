import React from "react";
import { isEmpty, isEqual } from "lodash";
import { useEffect, useMemo, useRef, useState } from "react";
import { createTheme, CssBaseline, LinearProgress, ThemeProvider, useColorScheme, useMediaQuery } from "@mui/material";
import { Excalidraw, MainMenu, THEME } from "@excalidraw/excalidraw";
import { OpenDrawingDialog } from "./features/drawing/OpenDrawing";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types";
import { useAppDispatch, useAppSelector } from "./app/hooks";
import { selectSavedDrawing, selectDrawingToEditStatus, selectCurrentDrawingContent, drawingContentChanged, AsyncOperationState } from "./features/drawing/drawingSlice";
import { SaveDrawingDialog } from "./features/drawing/SaveDrawing";

import "@excalidraw/excalidraw/index.css";

import "./App.css";
import { ManageDrawingsDialog } from "./features/drawing/ManageDrawingsDialog";
import { selectErrors } from "./features/app/appSlice";
import { useReporters } from "./utils/use-reporters";

const App = () => {

	const prefersDarkMode = useMediaQuery("(prefers-color-scheme: dark)");

	const { setMode: setMuiColorScehemeMode } = useColorScheme();

	const [excalidrawAPI, setExcalidrawAPI] = useState<ExcalidrawImperativeAPI | null>(null);
	const excalidrawAPIUnsubscribe = useRef<(() => void) | null>(null);

	const currentDrawingStatus = useAppSelector(selectDrawingToEditStatus);
	const savedDrawing = useAppSelector(selectSavedDrawing);
	const currentContent = useAppSelector(selectCurrentDrawingContent);

	const [openDrawingDialogOpen, setOpenDrawingDialogOpen] = useState(false);
	const [saveDrawingDialogOpen, setSaveDrawingDialogOpen] = useState(false);
	const [manageDrawingsDialogOpen, setManageDrawingsDialogOpen] = useState(false);

	const appErrors = useAppSelector(selectErrors);
	const { reportError } = useReporters();

	const dispatch = useAppDispatch();

	useEffect(() => {
		if (!isEmpty(appErrors)) {
			reportError(appErrors[0]);
		}
	}, [appErrors]);

	useEffect(() => {
		if (excalidrawAPI) {
			if (excalidrawAPIUnsubscribe.current !== null) {
				excalidrawAPIUnsubscribe.current();
			}
			excalidrawAPIUnsubscribe.current = excalidrawAPI.onChange(() => {
				const excalidrawContent = excalidrawAPI.getSceneElements();
				dispatch(drawingContentChanged(JSON.stringify(excalidrawContent)));

				const xcaliAppState = excalidrawAPI.getAppState();
				setMuiColorScehemeMode(xcaliAppState.theme === THEME.DARK ? "dark" : "light");
			});
		}
	}, [excalidrawAPI, savedDrawing]);

	useEffect(() => {
		const sceneData = {
			elements: savedDrawing.content ? JSON.parse(savedDrawing.content).elements : [],
			appState: {}
		};

		excalidrawAPI?.updateScene(sceneData);

		document.title = savedDrawing.title;
	}, [savedDrawing]);

	const contentHasChanged = useMemo(() => {
		return !isEqual(savedDrawing.content, currentContent);
	}, [savedDrawing, currentContent]);


	const darkTheme = createTheme({
		colorSchemes: {
			dark: true
		}
	});

	const lightTheme = createTheme({
		colorSchemes: {
			light: true
		}
	});

	return (
		<ThemeProvider theme={prefersDarkMode ? darkTheme : lightTheme}>
			<CssBaseline />

			<div>
				<div className="document-title">{savedDrawing.title}</div>
				<div>
					{currentDrawingStatus === AsyncOperationState.inProgress && <LinearProgress sx={{ marginTop: "-4px" }} />
					}
					<div className="xcali-area">
						<Excalidraw
							name={savedDrawing.title}
							excalidrawAPI={api => setExcalidrawAPI(api)}
							theme={prefersDarkMode ? "dark" : "light"}
							autoFocus={true}
						>
							<MainMenu>
								<MainMenu.Item onSelect={() => setOpenDrawingDialogOpen(true)}>
									Open
								</MainMenu.Item>
								<MainMenu.Item disabled={!contentHasChanged} onSelect={() => setSaveDrawingDialogOpen(true)}>
									Save
								</MainMenu.Item>
								<MainMenu.Item disabled={!contentHasChanged} onSelect={() => setManageDrawingsDialogOpen(true)}>
									Manage drawings
								</MainMenu.Item>
								<MainMenu.DefaultItems.Export />
								<MainMenu.DefaultItems.SaveAsImage />
								<MainMenu.DefaultItems.CommandPalette />
								<MainMenu.Separator />
								<MainMenu.Separator />
								<MainMenu.DefaultItems.ChangeCanvasBackground />
							</MainMenu>
						</Excalidraw>
					</div>
					<OpenDrawingDialog open={openDrawingDialogOpen} onClose={() => setOpenDrawingDialogOpen(false)} />
					<SaveDrawingDialog open={saveDrawingDialogOpen} onClose={() => setSaveDrawingDialogOpen(false)} />
					<ManageDrawingsDialog open={manageDrawingsDialogOpen} onClose={() => setManageDrawingsDialogOpen(false)} />
				</div>
			</div>

		</ThemeProvider>
	);
};

export default App;
