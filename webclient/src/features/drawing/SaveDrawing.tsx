import React from "react";
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { AsyncOperationState, saveDrawingContent, selectCurrentDrawingContent, selectSavedDrawing, selectSaveDrawingStatus } from "./drawingSlice";

import style from "./Drawing.module.css";
import { useReporters } from "../../utils/use-reporters";

interface SaveDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
}

export const SaveDrawingDialog = ({ open, onClose }: SaveDrawingDialogProps) => {

	const savedDrawing = useAppSelector(selectSavedDrawing);
	const currentContent = useAppSelector(selectCurrentDrawingContent);
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
		dispatch(saveDrawingContent({ title: selectedTitle, content: currentContent }));
		onClose();
	};

	const titleToOffer = selectedTitle || savedDrawing.title;

	console.log(">>>>>>> selectedTitle: ", selectedTitle, ", savedDrawing.title: ", savedDrawing.title, ", titleToOffer:", titleToOffer);

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