import React from "react";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Select } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { AsyncOperationState, getDrawingContent, getDrawingList, selectDrawingList, selectDrawingListStatus } from "./drawingSlice";

import style from "./Drawing.module.css";
import { DrawingListItem } from "./drawingAPI";
import { isNil } from "lodash";

interface OpenDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
}

export const OpenDrawingDialog = ({ open, onClose }: OpenDrawingDialogProps) => {

	const drawingListStatus = useAppSelector(selectDrawingListStatus);

	const [selectedDrawing, setSelectedDrawing] = useState<DrawingListItem | undefined>(undefined);

	const dispatch = useAppDispatch();

	useEffect(() => {
		if (open) {
			dispatch(getDrawingList());
		}
	}, [open]);

	const handleOk = () => {
		if (isNil(selectedDrawing?.id)) {
			console.warn("selectedDrawing?.id is nil");
			return;
		}
		dispatch(getDrawingContent(selectedDrawing.id));
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
							: <DrawingSelector selectedDrawing={selectedDrawing} onChange={selection => setSelectedDrawing(selection)} />
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
	readonly selectedDrawing: DrawingListItem | undefined;
	readonly onChange: (selectedDrawing: DrawingListItem | undefined) => void;
}

const DrawingSelector = ({ selectedDrawing, onChange }: DrawingSelectorProps) => {
	const drawingList = useAppSelector(selectDrawingList);

	const drawingToShowSelected = selectedDrawing || drawingList[0] || "";
	useEffect(() => {
		onChange(drawingToShowSelected);
	}, [drawingToShowSelected]);
	return (
		<Select className={style.openSaveDrawingDialogContent}
			label="Drawing"
			onChange={event => onChange(drawingList.find(drawing => drawing.id === event.target.value))}
			value={drawingToShowSelected.id}
		>{
				drawingList.map(drawing => <MenuItem key={drawing.id} value={drawing.id}>{drawing.title}</MenuItem>)
			}</Select>
	);
};