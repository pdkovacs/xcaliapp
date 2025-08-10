import React from "react";
import { isEqual } from "lodash";
import { useEffect, useMemo, useRef, useState } from "react";
import { LinearProgress, useMediaQuery } from "@mui/material";
import { Excalidraw, MainMenu, THEME } from "@excalidraw/excalidraw";
import { OpenDrawingDialog } from "./features/drawing/OpenDrawing";
import type { ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types";
import { useAppDispatch, useAppSelector } from "./app/hooks";
import { selectSavedDrawing, selectDrawingToEditStatus, selectCurrentDrawingContent, drawingContentChanged } from "./features/drawing/drawingSlice";
import { SaveDrawingDialog } from "./features/drawing/SaveDrawing";

import { ThemeProvider, createTheme, useColorScheme } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";


import "@excalidraw/excalidraw/index.css";

import "./App.css";

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

	const dispatch = useAppDispatch();

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


	return (
		<ThemeProvider theme={darkTheme}>
			<CssBaseline />

			<div>
				<div className="document-title">{savedDrawing.title}</div>
				<div>
					{currentDrawingStatus === "loading" && <LinearProgress sx={{ marginTop: "-4px" }} />
					}
					<div className="xcali-area">
						<Excalidraw
							excalidrawAPI={api => setExcalidrawAPI(api)}
							theme={prefersDarkMode ? "dark" : "light"}
						>
							<MainMenu>
								<MainMenu.Item onSelect={() => setOpenDrawingDialogOpen(true)}>
									Open
								</MainMenu.Item>
								<MainMenu.Item disabled={!contentHasChanged} onSelect={() => setSaveDrawingDialogOpen(true)}>
									Save
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
				</div>
			</div>

		</ThemeProvider>
	);
};

export default App;
