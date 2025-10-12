import React from "react";
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { AsyncOperationState, createDrawing, saveDrawingContent, selectDrawingRepos, selectSavedDrawing, selectSaveDrawingStatus } from "./drawingSlice";

import style from "./Drawing.module.css";
import { useReporters } from "../../utils/use-reporters";
import { DrawingRepoRef, XcalidrawContent } from "./drawingAPI";
import { RepositorySelector } from "./RepositorySelector";

interface SaveDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
	readonly currentContent: XcalidrawContent;
}

export const SaveDrawingDialog = ({ open, onClose, currentContent }: SaveDrawingDialogProps) => {

	const availableRepos = useAppSelector(selectDrawingRepos);
	const savedDrawing = useAppSelector(selectSavedDrawing);
	const savingStatus = useAppSelector(selectSaveDrawingStatus);

	const [selectedRepo, setSelectedRepo] = useState<DrawingRepoRef>(savedDrawing.repo);
	const [selectedTitle, setSelectedTitle] = useState<string>("");

	const dispatch = useAppDispatch();

	const { reportError } = useReporters();

	useEffect(() => {
		if (savingStatus === AsyncOperationState.failed) {
			reportError("Failed to save drawing");
		}
	}, [savingStatus]);

	const updateExistingDrawing = selectedTitle === savedDrawing.title;

	const handleOk = () => {
		if (updateExistingDrawing) {
			dispatch(saveDrawingContent({
				...savedDrawing,
				elements: currentContent
			}));
		} else {
			dispatch(createDrawing({
				repo: selectedRepo,
				title: selectedTitle,
				elements: currentContent
			}));
		}
		onClose();
	};

	const titleToOffer = selectedTitle || savedDrawing.title;

	useEffect(() => {
		setSelectedTitle(titleToOffer);
	}, [titleToOffer]);

	useEffect(() => {
		setSelectedRepo(savedDrawing.repo);
		setSelectedTitle(savedDrawing.title);
	}, [savedDrawing]);

	const repoToShowSelected = selectedRepo || availableRepos[0] || "";

	return (
		<Dialog
			open={open}
			onClose={onClose}
		>
			<DialogTitle>Save drawing</DialogTitle>
			<DialogContent>{
				<div>{
					savingStatus === AsyncOperationState.inProgress
						? <div className={style.fetchInProgress}>
							<CircularProgress />
						</div>
						: savingStatus === AsyncOperationState.idle || savingStatus === AsyncOperationState.failed
							? <div className={style.openSaveDrawingDialogContent}	>
								<RepositorySelector
									availableRepos={availableRepos}
									currentSelection={repoToShowSelected}
									disabled={updateExistingDrawing}
									requestSelectionChange={setSelectedRepo}
								/>
								<TitleSelector title={titleToOffer} onChange={title => setSelectedTitle(title)} />
							</div>
							: null
				}</div>
			}
			</DialogContent>
			<DialogActions>
				<Button onClick={() => {
					setSelectedTitle(savedDrawing.title);
					onClose();
				}}>Cancel</Button>
				<Button onClick={handleOk} autoFocus>OK</Button>
			</DialogActions>
		</Dialog>
	);
};

interface TitleSelectorProps {
	readonly title: string;
	readonly onChange: (selectedTitle: string) => void;
}

const TitleSelector = ({ title, onChange }: TitleSelectorProps) => {
	return (
		<TextField label="Title" size="small" sx={{ width: "100%" }} onChange={event => onChange(event.target.value)} value={title} />
	);
};