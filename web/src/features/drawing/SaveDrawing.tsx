import React from "react";
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { emptyDrawing, selectSavedDrawing, setSavedDrawing } from "./drawingSlice";
import {
	useCreateDrawingMutation,
	useSaveDrawingMutation,
	type DrawingRepoRef,
	type XcalidrawContent
} from "./drawingApi";

import style from "./Drawing.module.css";
import { RepositorySelector } from "./RepositorySelector";

interface SaveDrawingDialogProps {
	readonly open: boolean;
	readonly availableRepos: DrawingRepoRef[];
	readonly onClose: () => void;
	readonly currentContent: XcalidrawContent;
}

export const SaveDrawingDialog = ({ open, availableRepos, onClose, currentContent }: SaveDrawingDialogProps) => {
	if (!open) {
		return null;
	}
	return <SaveDrawingDialogContent open availableRepos={availableRepos} onClose={onClose} currentContent={currentContent} />;
};

const SaveDrawingDialogContent = ({ open, availableRepos, onClose, currentContent }: SaveDrawingDialogProps) => {

	const savedDrawing = useAppSelector(selectSavedDrawing);
	const [createDrawing, { isLoading: isCreating }] = useCreateDrawingMutation();
	const [saveDrawing, { isLoading: isSaving }] = useSaveDrawingMutation();
	const inProgress = isCreating || isSaving;

	const [selectedRepo, setSelectedRepo] = useState<DrawingRepoRef>(availableRepos[0]);
	const [selectedTitle, setSelectedTitle] = useState<string>("");

	const updateExistingDrawing = selectedTitle === savedDrawing.title;

	const dispatch = useAppDispatch();

	const handleOk = async () => {
		if (updateExistingDrawing) {
			const updated = {
				...savedDrawing,
				elements: currentContent
			};
			await saveDrawing(updated).unwrap();
			dispatch(setSavedDrawing(updated));
		} else {
			createDrawing({
				repo: selectedRepo,
				title: selectedTitle,
				elements: currentContent
			});
		}
		onClose();
	};

	const titleToOffer = selectedTitle || savedDrawing.title;

	useEffect(() => {
		setSelectedTitle(titleToOffer);
	}, [titleToOffer]);

	useEffect(() => {
		if (savedDrawing === emptyDrawing) {
			setSelectedRepo(availableRepos[0]);
			setSelectedTitle("");
		} else {
			setSelectedRepo(savedDrawing.repo);
			setSelectedTitle(savedDrawing.title);
		}
	}, [savedDrawing, availableRepos]);


	return (
		<Dialog
			open={open}
			onClose={onClose}
		>
			<DialogTitle>Save drawing</DialogTitle>
			<DialogContent>
				<div>{
					inProgress
						? <div className={style.fetchInProgress}>
							<CircularProgress />
						</div>
						: <div className={style.openSaveDrawingDialogContent}>
							<RepositorySelector
								availableRepos={[...availableRepos]}
								currentSelection={selectedRepo}
								disabled={updateExistingDrawing}
								requestSelectionChange={setSelectedRepo}
							/>
							<TitleSelector title={titleToOffer} onChange={title => setSelectedTitle(title)} />
						</div>
				}</div>
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

