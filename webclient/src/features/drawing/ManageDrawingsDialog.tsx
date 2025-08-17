import React, { useEffect, useMemo, useState } from "react";
import { Alert, Button, Checkbox, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, LinearProgress, Menu, MenuItem, TextField } from "@mui/material";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { AsyncOperationState, getDrawingList, selectDrawingList, selectDrawingListStatus } from "./drawingSlice";

import style from "./ManageDrawingsDialog.module.css";
import { isEmpty } from "lodash";
import classNames from "classnames";
import { ErrorDialog, ErrorDialogData } from "../../utils/ErrorDialog";
import { deleteDrawings, renameDrawing } from "./drawingAPI";

export interface ManageDrawingsDialogProps {
	readonly open: boolean
	readonly onClose: () => void;
}

export const ManageDrawingsDialog = ({ open, onClose }: ManageDrawingsDialogProps) => {
	const drawingListStatus = useAppSelector(selectDrawingListStatus);

	const [workInProgress, setWorkInProgress] = useState(false);

	const dispatch = useAppDispatch();

	const [selectedTitles, setSelectedTitles] = useState<string[]>([]);

	const [renameDrawingDialogOpen, setRenameDrawingDialogOpen] = useState(false);
	const [deleteDrawingsDialogOpen, setDeleteDrawingsDialogOpen] = useState(false);
	const [actionExecutionError, setActionExecutionError] = useState<ErrorDialogData | null>(null);

	useEffect(() => {
		if (open) {
			dispatch(getDrawingList());
		}
	}, [open, selectedTitles]);

	const enabledActions = useMemo(() => {
		if (isEmpty(selectedTitles)) {
			return [];
		}
		return selectedTitles.length === 1
			? [Action.RENAME, Action.DELETE]
			: [Action.DELETE];
	}, [selectedTitles]);

	const handleActionDialogOpen = (selectedAction: Action) => {
		switch (selectedAction) {
			case Action.RENAME:
				setRenameDrawingDialogOpen(true);
				break;
			case Action.DELETE:
				setDeleteDrawingsDialogOpen(true);
				break;
		}
	};

	const handleActionDialogClosed = async (action: Action, confirmed: boolean | string) => {
		setWorkInProgress(true);
		try {
			switch (action) {
				case Action.RENAME:
					setRenameDrawingDialogOpen(false);
					if (confirmed !== false) {
						await new Promise(resolve => setTimeout(resolve, 2000));
						await renameDrawing(selectedTitles[0], confirmed as string);
						setSelectedTitles([]);
					}
					break;
				case Action.DELETE:
					setDeleteDrawingsDialogOpen(false);
					if (confirmed) {
						await new Promise(resolve => setTimeout(resolve, 2000));
						await deleteDrawings(selectedTitles);
						setSelectedTitles([]);
					}
					break;
			}
		} catch (err) {
			console.error(err);
			setActionExecutionError({
				title: `Failed to ${action.toLowerCase()} drawing(s)`,
				message: (err as string).toString()
			});
		} finally {
			setWorkInProgress(false);
		}
	};

	const handleTitleSelectionChange = (title: string, checked: boolean) => {
		if (checked && !selectedTitles.includes(title)) {
			setSelectedTitles(selectedTitles.concat(title));
		} else {
			setSelectedTitles(selectedTitles.filter(selectedTitle => selectedTitle !== title));
		}
	};

	return <>
		<Dialog maxWidth="xl" open={open} onClose={onClose}>
			<DialogTitle className={style.dialogTitle}>
				<span>Manage Drawings</span>
				<ActionsMenu enabledActions={enabledActions} onActionSelected={selectedAction => handleActionDialogOpen(selectedAction)} />
			</DialogTitle>
			<DialogContent>
				{workInProgress && <LinearProgress />}
				<div className={classNames(style.dialogContent, { [style.workInProgress]: workInProgress })}>{
					drawingListStatus === AsyncOperationState.inProgress
						? <CircularProgress />
						: drawingListStatus === AsyncOperationState.failed
							? <Alert severity="error">Failed to load drawing list</Alert>
							: <ManageDrawingsPanel selectedTitles={selectedTitles} onTitleSelectionChanged={(title, checked) => handleTitleSelectionChange(title, checked)} />
				}</div>
			</DialogContent>
			<DialogActions>
				<Button onClick={onClose}>Close</Button>
			</DialogActions>
		</Dialog>

		<RenameDrawingDialog open={renameDrawingDialogOpen} title={selectedTitles[0]} onClose={confirmed => handleActionDialogClosed(Action.RENAME, confirmed)} />
		<DeleteDrawingsDialog open={deleteDrawingsDialogOpen} titles={selectedTitles} onClose={confirmed => handleActionDialogClosed(Action.DELETE, confirmed)} />
		<ErrorDialog open={actionExecutionError !== null} onClose={() => setActionExecutionError(null)} data={actionExecutionError!} />
	</>;
};

interface ManageDrawingsPanelProps {
	readonly selectedTitles: string[];
	readonly onTitleSelectionChanged: (title: string, checked: boolean) => void;
}

const ManageDrawingsPanel = ({ selectedTitles, onTitleSelectionChanged }: ManageDrawingsPanelProps) => {
	const drawingList = useAppSelector(selectDrawingList);

	return <div className={style.wideDialogContent + " " + style.overflowingDialogContent}>
		<div>
			{
				drawingList.map(drawingTitle => {
					return <div className={style.drawingListItem} key={drawingTitle}>
						<Checkbox checked={selectedTitles.includes(drawingTitle)} onChange={change => onTitleSelectionChanged(drawingTitle, change.target.checked)} />
						<div>{drawingTitle}</div>
					</div>;
				})
			}
		</div>
	</div>;
};

enum Action {
	RENAME = "RENAME",
	DELETE = "DELETE"
}

interface ActionsMenuProps {
	readonly enabledActions: Action[];
	readonly onActionSelected: (selectedAction: Action) => void;
}

const ActionsMenu = ({ enabledActions, onActionSelected }: ActionsMenuProps) => {
	const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
	const open = Boolean(anchorEl);

	const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
		setAnchorEl(event.currentTarget);
	};

	const handleClose = () => {
		setAnchorEl(null);
	};

	const handleActionSelected = (selectedAction: Action) => {
		handleClose();
		onActionSelected(selectedAction);
	};

	return <div>
		<Button
			id="basic-button"
			onClick={handleClick}
		>
			Actions
		</Button>
		<Menu
			id="basic-menu"
			anchorEl={anchorEl}
			open={open}
			onClose={handleClose}
		>
			<MenuItem disabled={!enabledActions.includes(Action.RENAME)} onClick={() => handleActionSelected(Action.RENAME)}>Rename</MenuItem>
			<MenuItem disabled={!enabledActions.includes(Action.DELETE)} onClick={() => handleActionSelected(Action.DELETE)}>Delete</MenuItem>
		</Menu>
	</div>;
};

interface RenameDrawingDialogProps {
	readonly open: boolean;
	readonly title: string;
	readonly onClose: (confirmedRenameTo: string | false) => void;
}

const RenameDrawingDialog = ({ open, title, onClose }: RenameDrawingDialogProps) => {
	const [titleToRenameTo, setTitleToRenameTo] = useState<string | null>(null);

	const handleOnClose = (confirmedRenameTo: string | false) => {
		setTitleToRenameTo(null);
		onClose(confirmedRenameTo);
	};

	return <Dialog open={open} maxWidth="xl">
		<DialogTitle>Rename drawing</DialogTitle>
		<DialogContent className={style.wideDialogContent + " " + style.overflowingDialogContent}>
			<TextField label="New title" size="small" sx={{ width: "100%" }} onChange={event => setTitleToRenameTo(event.target.value)} value={titleToRenameTo ?? title} />
		</DialogContent>
		<DialogActions>
			<Button onClick={() => handleOnClose(false)}>Cancel</Button>
			<Button disabled={titleToRenameTo === null || titleToRenameTo === title} onClick={() => handleOnClose(titleToRenameTo!)}>Rename</Button>
		</DialogActions>
	</Dialog>;
};

interface DeleteDrawingsDialogProps {
	readonly open: boolean;
	readonly titles: string[];
	readonly onClose: (conformed: boolean) => void;
}

const DeleteDrawingsDialog = ({ open, titles, onClose }: DeleteDrawingsDialogProps) => {
	return <Dialog open={open} maxWidth="xl">
		<DialogTitle>Delete drawings</DialogTitle>
		<DialogContent className={style.wideDialogContent + " " + style.overflowingDialogContent}>
			<p>Are you sure to delete the following drawings:</p>
			<div>
				{
					titles.map(title => <div key={title}>{title}</div>)
				}
			</div>
		</DialogContent>
		<DialogActions>
			<Button onClick={() => onClose(false)}>Cancel</Button>
			<Button color="error" onClick={() => onClose(true)}>Delete</Button>
		</DialogActions>
	</Dialog>;
};
