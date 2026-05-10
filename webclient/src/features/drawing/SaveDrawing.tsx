import React from "react";
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { useEffect, useState } from "react";
import { selectSavedDrawing, setSavedDrawing } from "./drawingSlice";
import {
	useCreateDrawingMutation,
	useGetDrawingRepositoriesQuery,
	useSaveDrawingMutation,
	type DrawingRepoRef,
	type XcalidrawContent
} from "./drawingApi";

import style from "./Drawing.module.css";
import { RepositorySelector } from "./RepositorySelector";
import { emptyArray } from "../../utils/empty-array";

interface SaveDrawingDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
	readonly currentContent: XcalidrawContent;
}

export const SaveDrawingDialog = ({ open, onClose, currentContent }: SaveDrawingDialogProps) => {

	const { data: availableRepos = emptyArray } = useGetDrawingRepositoriesQuery();
	const savedDrawing = useAppSelector(selectSavedDrawing);
	const [createDrawing, { isLoading: isCreating }] = useCreateDrawingMutation();
	const [saveDrawing, { isLoading: isSaving }] = useSaveDrawingMutation();
	const inProgress = isCreating || isSaving;

	const [selectedRepo, setSelectedRepo] = useState<DrawingRepoRef>(savedDrawing.repo);
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
		setSelectedRepo(savedDrawing.repo);
		setSelectedTitle(savedDrawing.title);
	}, [savedDrawing]);

	const repoToShowSelected = selectedRepo || availableRepos[0];

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
								currentSelection={repoToShowSelected}
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
