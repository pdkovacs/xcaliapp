import React from "react";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, FormControl, InputLabel, MenuItem, Select } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { AsyncOperationState, getDrawingContent, getDrawingLists, selectDrawingLists, selectDrawingListStatus, selectDrawingRepos } from "./drawingSlice";

import style from "./Drawing.module.css";
import { DrawingRepoRef, DrawingRepoItem } from "./drawingAPI";
import { isEmpty, isNil } from "lodash";

interface OpenDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
}

export const OpenDrawingDialog = ({ open, onClose }: OpenDrawingDialogProps) => {

	const drawingListStatus = useAppSelector(selectDrawingListStatus);

	const [drawingSelection, setDrawingSelection] = useState<SelectedDrawing | undefined>(undefined);

	const dispatch = useAppDispatch();

	useEffect(() => {
		if (open) {
			dispatch(getDrawingLists());
		}
	}, [open]);

	const handleOk = () => {
		if (isNil(drawingSelection?.drawing?.id)) {
			console.warn("selectedDrawing?.id is nil: ", drawingSelection);
			return;
		}
		dispatch(getDrawingContent({ repoId: drawingSelection.repo.name, drawingId: drawingSelection.drawing.id }));
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
							: <DrawingSelector selection={drawingSelection} onChange={selection => setDrawingSelection(selection)} />
				}</div>
			</DialogContent>
			<DialogActions>
				<Button onClick={onClose}>Cancel</Button>
				<Button onClick={handleOk} autoFocus>OK</Button>
			</DialogActions>
		</Dialog>
	);
};

interface SelectedDrawing {
	readonly repo: DrawingRepoRef;
	readonly drawing: DrawingRepoItem | undefined;
}

interface DrawingSelectorProps {
	readonly selection: SelectedDrawing | undefined;
	readonly onChange: (selection: SelectedDrawing | undefined) => void;
}

const DrawingSelector = ({ selection, onChange }: DrawingSelectorProps) => {
	const { repo: selectedRepo, drawing: selectedDrawing } = selection ?? { repo: undefined, drawing: undefined };
	const repos = useAppSelector(selectDrawingRepos);
	const repoToShowSelected = selectedRepo || repos[0] || "";

	const drawingLists = useAppSelector(selectDrawingLists);
	const selectedRepoContent = drawingLists?.[repoToShowSelected.name]?.items;
	const drawingToShowSelected = selectedDrawing || selectedRepoContent?.[0] || "";

	useEffect(() => {
		if (isEmpty(repos) || isEmpty(drawingLists)) {
			return;
		}

		const repo = repos[0];
		const drawingsInRepo = drawingLists[repo.name];
		if (isNil(drawingsInRepo) || isEmpty(drawingsInRepo.items)) {
			console.warn("No drawings in repo ", repo.label, " (yet). Funny! (drawingsInRepo: ", drawingsInRepo, ")");
			return;
		}

		onChange({
			repo,
			drawing: drawingsInRepo.items[0]
		});
	}, [repos, drawingLists]);


	return <div className={style.openSaveDrawingDialogContent}>
		<FormControl className={style.selectRepo}>
			<InputLabel>Repository</InputLabel>
			<Select
				label="Repository"
				onChange={event => {
					const repo = repos.find(repo => repo.name === event.target.value) ?? repoToShowSelected;
					const drawing = drawingLists[repo.name].items[0];
					onChange({ repo, drawing });
				}}
				value={repoToShowSelected.name}
			>
				{
					repos.map(repo => <MenuItem key={repo.name} value={repo.name}>{repo.label}</MenuItem>)
				}
			</Select>
		</FormControl>

		<FormControl className={style.selectDrawing}>
			<InputLabel>Drawing</InputLabel>
			<Select
				label="Drawing"
				onChange={event => onChange({
					repo: repoToShowSelected,
					drawing: selectedRepoContent.find(drawing => drawing.id === event.target.value) ?? drawingToShowSelected
				})}
				value={drawingToShowSelected.id}
			>
				{
					selectedRepoContent.map(drawing => <MenuItem key={drawing.id} value={drawing.id}>{drawing.title}</MenuItem>)
				}
			</Select>
		</FormControl>
	</div >;
};