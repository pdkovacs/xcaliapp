import React from "react";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Select } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { AsyncOperationState, getDrawingContent, getDrawingList, selectDrawingList, selectDrawingListStatus } from "./drawingSlice";

import style from "./Drawing.module.css";

interface OpenDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
}

export const OpenDrawingDialog = ({ open, onClose }: OpenDrawingDialogProps) => {

	const drawingListStatus = useAppSelector(selectDrawingListStatus);

	const [titleOfSelectedDrawing, setTitleOfSelectedDrawing] = useState<string>("");

	const dispatch = useAppDispatch();

	useEffect(() => {
		if (open) {
			dispatch(getDrawingList());
		}
	}, [open]);

	const handleOk = () => {
		dispatch(getDrawingContent(titleOfSelectedDrawing));
		onClose();
	};

	return (
		<Dialog
			open={open}
			onClose={onClose}
		>
			<DialogTitle>Open drawing</DialogTitle>
			<DialogContent>
				<div className={style.openSaveDrawingDialogContent}>{
					drawingListStatus === AsyncOperationState.inProgress
						? <CircularProgress />
						: drawingListStatus === AsyncOperationState.failed
							? <Alert severity="error">Failed to load drawing list</Alert>
							: <DrawingSelector selectedDrawing={titleOfSelectedDrawing} onChange={selection => setTitleOfSelectedDrawing(selection)} />
				}</div>
			</DialogContent>
			<DialogActions>
				<Button onClick={onClose}>Cancel</Button>
				<Button onClick={handleOk} autoFocus>OK</Button>
			</DialogActions>
		</Dialog>
	);
};

interface DrawingSelectorProps {
	readonly selectedDrawing: string;
	readonly onChange: (selectedDrawing: string) => void;
}

const DrawingSelector = ({ selectedDrawing, onChange }: DrawingSelectorProps) => {
	const drawingList = useAppSelector(selectDrawingList);

	const selectedTitle = selectedDrawing || drawingList[0] || "";
	useEffect(() => {
		onChange(selectedTitle);
	}, [selectedTitle]);
	return (
		<Select className={style.openSaveDrawingDialogContent}
			label="Drawing"
			onChange={event => onChange(event.target.value)}
			value={selectedTitle}
		>{
				drawingList.map(drawingTitle => <MenuItem key={drawingTitle} value={drawingTitle}>{drawingTitle}</MenuItem>)
			}</Select>
	);
};