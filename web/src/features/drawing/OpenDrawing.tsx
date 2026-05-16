import React from "react";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, FormControl, InputLabel, MenuItem, Select } from "@mui/material";
import { useEffect, useState } from "react";
import {
	useGetDrawingListsQuery,
	useLazyGetDrawingQuery,
	type DrawingLists,
	type DrawingRepoItem,
	type DrawingRepoRef
} from "./drawingApi";
import { isEmpty, isNil } from "lodash";
import { RepositorySelector } from "./RepositorySelector";

import style from "./Drawing.module.css";

interface OpenDrawingDialogProps {
	readonly open: boolean;
	readonly availableRepos: DrawingRepoRef[];
	readonly onClose: () => void;
}

export const OpenDrawingDialog = ({ open, availableRepos, onClose }: OpenDrawingDialogProps) => {
	if (!open) {
		return null;
	}
	return <OpenDrawingDialogContent open availableRepos={availableRepos} onClose={onClose} />;
};

const OpenDrawingDialogContent = ({ open, availableRepos, onClose }: OpenDrawingDialogProps) => {

	const { data: drawingLists, isFetching, isError } = useGetDrawingListsQuery(undefined, {
		skip: !open,
		refetchOnMountOrArgChange: true
	});
	const [triggerGetDrawing] = useLazyGetDrawingQuery();

	const [drawingSelection, setDrawingSelection] = useState<SelectedDrawing | undefined>(undefined);

	const handleOk = () => {
		if (isNil(drawingSelection?.drawing?.id)) {
			console.warn("selectedDrawing?.id is nil: ", drawingSelection);
			return;
		}
		triggerGetDrawing({ repoId: drawingSelection.repo.name, drawingId: drawingSelection.drawing.id });
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
					isFetching
						? <CircularProgress />
						: isError
							? <Alert severity="error">Failed to load drawing list</Alert>
							: <DrawingSelector
								repos={availableRepos}
								drawingLists={drawingLists}
								selection={drawingSelection}
								onChange={selection => setDrawingSelection(selection)}
							/>
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
	readonly repos: readonly DrawingRepoRef[];
	readonly drawingLists: DrawingLists | undefined;
	readonly selection: SelectedDrawing | undefined;
	readonly onChange: (selection: SelectedDrawing | undefined) => void;
}

const DrawingSelector = ({ repos, drawingLists, selection, onChange }: DrawingSelectorProps) => {
	const { repo: selectedRepo, drawing: selectedDrawing } = selection ?? { repo: undefined, drawing: undefined };
	const repoToShowSelected = selectedRepo || repos[0] || "";

	const selectedRepoContent = drawingLists?.[(repoToShowSelected as DrawingRepoRef).name]?.items;
	const drawingToShowSelected = selectedDrawing || selectedRepoContent?.[0] || "";

	useEffect(() => {
		if (isEmpty(repos) || isEmpty(drawingLists)) {
			return;
		}

		const repo = repos[0];
		const drawingsInRepo = drawingLists![repo.name];
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
		<RepositorySelector availableRepos={[...repos]} currentSelection={repoToShowSelected as DrawingRepoRef} requestSelectionChange={requestedRepo => {
			const drawing = drawingLists?.[requestedRepo.name]?.items[0];
			onChange({ repo: requestedRepo, drawing });
		}} />
		<FormControl className={style.selectDrawing}>
			<InputLabel>Drawing</InputLabel>
			<Select
				label="Drawing"
				onChange={event => onChange({
					repo: repoToShowSelected as DrawingRepoRef,
					drawing: selectedRepoContent?.find(drawing => drawing.id === event.target.value) ?? (drawingToShowSelected as DrawingRepoItem)
				})}
				value={(drawingToShowSelected as DrawingRepoItem).id}
			>
				{
					selectedRepoContent?.map(drawing => <MenuItem key={drawing.id} value={drawing.id}>{drawing.title}</MenuItem>)
				}
			</Select>
		</FormControl>
	</div >;
};

