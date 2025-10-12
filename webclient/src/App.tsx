import React, { useCallback } from "react";
import { isEmpty, isEqual, isNil } from "lodash";
import { useEffect, useMemo, useRef, useState } from "react";
import { createTheme, CssBaseline, LinearProgress, ThemeProvider, useColorScheme, useMediaQuery } from "@mui/material";
import { Excalidraw, MainMenu, THEME } from "@excalidraw/excalidraw";
import { OpenDrawingDialog } from "./features/drawing/OpenDrawing";
import { AppState, type ExcalidrawImperativeAPI } from "@excalidraw/excalidraw/types";
import { useAppDispatch, useAppSelector } from "./app/hooks";
import { selectSavedDrawing, selectDrawingToEditStatus, AsyncOperationState, getDrawingContent, clearCanvas, selectDrawingRepos, getDrawingRepositories } from "./features/drawing/drawingSlice";
import { SaveDrawingDialog } from "./features/drawing/SaveDrawing";

import "@excalidraw/excalidraw/index.css";

import "./App.css";
import { ManageDrawingsDialog } from "./features/drawing/ManageDrawingsDialog";
import { selectErrors } from "./features/app/appSlice";
import { useReporters } from "./utils/use-reporters";
import { XcalidrawContent } from "./features/drawing/drawingAPI";
import { convertToPlainObject } from "./utils/convert-to-plain-object";
import { emptyArray } from "./utils/empty-array";

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

const App = () => {

	const prefersDarkMode = useMediaQuery("(prefers-color-scheme: dark)");

	const drawingRepositories = useAppSelector(selectDrawingRepos);

	const { setMode: setMuiColorScehemeMode } = useColorScheme();

	const [excalidrawAPI, setExcalidrawAPI] = useState<ExcalidrawImperativeAPI | null>(null);
	const excalidrawAPIUnsubscribe = useRef<(() => void) | null>(null);
	const excalidrawAppStateRef = useRef<AppState | null>(null);

	const currentDrawingStatus = useAppSelector(selectDrawingToEditStatus);
	const savedDrawing = useAppSelector(selectSavedDrawing);
	const [currentContent, setCurrentContent] = useState<XcalidrawContent>(emptyArray);
	const [openDrawingDialogOpen, setOpenDrawingDialogOpen] = useState(false);
	const [saveDrawingDialogOpen, setSaveDrawingDialogOpen] = useState(false);
	const [manageDrawingsDialogOpen, setManageDrawingsDialogOpen] = useState(false);

	const appErrors = useAppSelector(selectErrors);
	const { reportError } = useReporters();

	const dispatch = useAppDispatch();

	useEffect(() => {
		dispatch(getDrawingRepositories());
	}, []);

	useEffect(() => {
		if (!isEmpty(appErrors)) {
			reportError(appErrors[0]);
		}
	}, [appErrors]);

	useEffect(() => {
		const drawingId = window.location.pathname.substring("/drawings/".length);
		if (isEmpty(drawingRepositories) || isEmpty(drawingId) || isNil(savedDrawing.id) || drawingId === savedDrawing.id) {
			return;
		}
		const idParts = drawingId.split("-");
		dispatch(getDrawingContent({ repoId: idParts[0], drawingId: idParts[1] }));
	}, [drawingRepositories, window.location]);

	useEffect(() => {
		if (excalidrawAPI) {
			if (excalidrawAPIUnsubscribe.current !== null) {
				excalidrawAPIUnsubscribe.current();
			}
			excalidrawAPIUnsubscribe.current = excalidrawAPI.onChange(() => {
				const excalidrawContent = excalidrawAPI.getSceneElements();
				setCurrentContent(excalidrawContent as XcalidrawContent);

				const xcaliAppState = excalidrawAPI.getAppState();
				setMuiColorScehemeMode(xcaliAppState.theme === THEME.DARK ? "dark" : "light");
				excalidrawAppStateRef.current = xcaliAppState;
			});
		}
	}, [excalidrawAPI]);

	const fullTitle = useMemo(() => savedDrawing.repo?.label ? `${savedDrawing.repo?.label}: ${savedDrawing.title}` : "", [savedDrawing]);

	useEffect(() => {
		if (excalidrawAPI && savedDrawing?.elements) {
			const sceneData = {
				elements: savedDrawing.elements ? convertToPlainObject(savedDrawing.elements) : emptyArray,
				appState: excalidrawAppStateRef.current
			};
			excalidrawAPI?.updateScene(sceneData);
		}
		document.title = fullTitle;
	}, [savedDrawing?.elements]);

	const contentHasChanged = useMemo(() => {
		return !isEqual(savedDrawing.elements, currentContent);
	}, [savedDrawing, currentContent]);

	const excalidrawElements = useMemo(() => {
		const hasSavedContent = !isNil(savedDrawing.elements) && !isEmpty(savedDrawing.elements);
		return hasSavedContent ? savedDrawing.elements : emptyArray;
	}, [savedDrawing.elements]);

	const clear = useCallback(() => {
		dispatch(clearCanvas());
	}, [dispatch, clearCanvas]);

	return (
		<ThemeProvider theme={prefersDarkMode ? darkTheme : lightTheme}>
			<CssBaseline />

			<div>
				<div className="document-title">{fullTitle}</div>
				<div>
					{
						currentDrawingStatus === AsyncOperationState.inProgress && <LinearProgress sx={{ marginTop: "-4px" }} />
					}
					<div className="xcali-area">
						<Excalidraw
							name={savedDrawing.title}
							excalidrawAPI={setExcalidrawAPI}
							theme={prefersDarkMode ? "dark" : "light"}
							autoFocus={true}
							initialData={{
								elements: excalidrawElements
							}}
						>
							<MainMenu>
								<MainMenu.Item onSelect={() => clear()}>
									New
								</MainMenu.Item>
								<MainMenu.Item onSelect={() => setOpenDrawingDialogOpen(true)}>
									Open
								</MainMenu.Item>
								<MainMenu.Item disabled={!contentHasChanged} onSelect={() => setSaveDrawingDialogOpen(true)}>
									Save
								</MainMenu.Item>
								<MainMenu.Item onSelect={() => setManageDrawingsDialogOpen(true)}>
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
					<SaveDrawingDialog open={saveDrawingDialogOpen} onClose={() => setSaveDrawingDialogOpen(false)} currentContent={currentContent} />
					<ManageDrawingsDialog open={manageDrawingsDialogOpen} onClose={() => setManageDrawingsDialogOpen(false)} />
				</div>
			</div>

		</ThemeProvider >
	);
};

export default App;
