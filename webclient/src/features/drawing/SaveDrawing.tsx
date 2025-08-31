import React from "react";
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { AsyncOperationState, createDrawing, saveDrawingContent, selectSavedDrawing, selectSaveDrawingStatus } from "./drawingSlice";

import style from "./Drawing.module.css";
import { useReporters } from "../../utils/use-reporters";
import { XcalidrawContent } from "./drawingAPI";

interface SaveDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
	readonly currentContent: XcalidrawContent;
}

export const SaveDrawingDialog = ({ open, onClose, currentContent }: SaveDrawingDialogProps) => {

	const savedDrawing = useAppSelector(selectSavedDrawing);
	const savingStatus = useAppSelector(selectSaveDrawingStatus);

	const [selectedTitle, setSelectedTitle] = useState<string>("");

	const dispatch = useAppDispatch();

	const { reportError } = useReporters();

	useEffect(() => {
		if (savingStatus === AsyncOperationState.failed) {
			reportError("Failed to save drawing");
		}
	}, [savingStatus]);

	const handleOk = () => {
		if (selectedTitle === savedDrawing.title) {
			dispatch(saveDrawingContent({
				id: savedDrawing.id,
				title: savedDrawing.title,
				elements: currentContent
			}));
		} else {
			dispatch(createDrawing({
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
		setSelectedTitle(savedDrawing.title);
	}, [savedDrawing]);

	return (
		<Dialog
			open={open}
			onClose={onClose}
		>
			<DialogTitle>Save drawing</DialogTitle>
			<DialogContent>{
				<div className={style.openSaveDrawingDialogContent}>{
					savingStatus === AsyncOperationState.inProgress
						? <div className={style.fetchInProgress}>
							<CircularProgress />
						</div>
						: savingStatus === AsyncOperationState.idle || savingStatus === AsyncOperationState.failed
							? <TitleSelector title={titleToOffer} onChange={title => setSelectedTitle(title)} />
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