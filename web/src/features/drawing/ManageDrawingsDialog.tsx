import React, { useEffect, useMemo, useState } from "react";
import { Alert, Button, Checkbox, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, LinearProgress, Menu, MenuItem, TextField } from "@mui/material";

import style from "./ManageDrawingsDialog.module.css";
import { isEmpty } from "lodash";
import classNames from "classnames";
import { ErrorDialog, ErrorDialogData } from "../../utils/ErrorDialog";
import {
	useDeleteDrawingsMutation,
	useGetDrawingListsQuery,
	useLazyGetDrawingQuery,
	useSaveDrawingMutation,
	type DrawingRepoItem,
	type DrawingRepoRef
} from "./drawingApi";
import { emptyArray } from "../../utils/empty-array";
import { RepositorySelector } from "./RepositorySelector";

export interface ManageDrawingsDialogProps {
	readonly open: boolean;
	readonly availableRepos: DrawingRepoRef[];
	readonly onClose: () => void;
}

export const ManageDrawingsDialog = ({ open, availableRepos, onClose }: ManageDrawingsDialogProps) => {
	if (!open) {
		return null;
	}
	return <ManageDrawingsDialogContent open availableRepos={availableRepos} onClose={onClose} />;
};

const ManageDrawingsDialogContent = ({ open, availableRepos, onClose }: ManageDrawingsDialogProps) => {
	const { data: drawingLists, isFetching, isError } = useGetDrawingListsQuery(undefined, {
		skip: !open,
		refetchOnMountOrArgChange: true
	});
	const [selectedRepo, setSelectedRepo] = useState<DrawingRepoRef>(availableRepos[0]);

	const [getDrawingTrigger] = useLazyGetDrawingQuery();
	const [saveDrawingMutation] = useSaveDrawingMutation();
	const [deleteDrawingsMutation] = useDeleteDrawingsMutation();

	const [workInProgress, setWorkInProgress] = useState(false);

	const [selectedDrawings, setSelectedDrawings] = useState<DrawingRepoItem[]>(emptyArray);

	const [renameDrawingDialogOpen, setRenameDrawingDialogOpen] = useState(false);
	const [deleteDrawingsDialogOpen, setDeleteDrawingsDialogOpen] = useState(false);
	const [actionExecutionError, setActionExecutionError] = useState<ErrorDialogData | null>(null);

	useEffect(() => {
		if (!open) {
			setSelectedDrawings(emptyArray);
		}
	}, [open]);

	const enabledActions = useMemo(() => {
		if (isEmpty(selectedDrawings)) {
			return emptyArray;
		}
		return selectedDrawings.length === 1
			? [Action.RENAME, Action.DELETE]
			: [Action.DELETE];
	}, [selectedDrawings]);

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
						const drawingId = selectedDrawings[0].id;
						const newTitle = confirmed as string;
						const content = await getDrawingTrigger({ repoId: selectedRepo.name, drawingId }).unwrap();
						await saveDrawingMutation({
							repo: selectedRepo,
							id: drawingId,
							title: newTitle,
							elements: content.elements
						}).unwrap();
						setSelectedDrawings(emptyArray);
					}
					break;
				case Action.DELETE:
					setDeleteDrawingsDialogOpen(false);
					if (confirmed) {
						await deleteDrawingsMutation(selectedDrawings.map(drawing => drawing.id)).unwrap();
						setSelectedDrawings(emptyArray);
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

	const handleDrawingSelectionChange = (drawing: DrawingRepoItem, checked: boolean) => {
		if (checked && !selectedDrawings.includes(drawing)) {
			setSelectedDrawings(selectedDrawings.concat(drawing));
		} else {
			setSelectedDrawings(selectedDrawings.filter(selectedDrawing => selectedDrawing !== drawing));
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
				<div className={classNames(style.dialogContent, { [style.workInProgress]: workInProgress })}>
					<RepositorySelector
						availableRepos={[...availableRepos]}
						currentSelection={selectedRepo}
						requestSelectionChange={setSelectedRepo}
					/>
					{
						isFetching
							? <CircularProgress />
							: isError
								? <Alert severity="error">Failed to load drawing list</Alert>
								: <ManageDrawingsPanel
									drawingLists={drawingLists}
									selectedRepo={selectedRepo}
									selectedDrawings={selectedDrawings}
									onDrawingSelectionChanged={(drawing, checked) => handleDrawingSelectionChange(drawing, checked)}
								/>
					}
				</div>
			</DialogContent>
			<DialogActions>
				<Button onClick={onClose}>Close</Button>
			</DialogActions>
		</Dialog>

		<RenameDrawingDialog open={renameDrawingDialogOpen} drawing={selectedDrawings[0] ?? ""} onClose={confirmed => handleActionDialogClosed(Action.RENAME, confirmed)} />
		<DeleteDrawingsDialog open={deleteDrawingsDialogOpen} drawings={selectedDrawings} onClose={confirmed => handleActionDialogClosed(Action.DELETE, confirmed)} />
		<ErrorDialog open={actionExecutionError !== null} onClose={() => setActionExecutionError(null)} data={actionExecutionError!} />
	</>;
};

interface ManageDrawingsPanelProps {
	readonly drawingLists: ReturnType<typeof useGetDrawingListsQuery>["data"];
	readonly selectedRepo: DrawingRepoRef;
	readonly selectedDrawings: DrawingRepoItem[];
	readonly onDrawingSelectionChanged: (drawing: DrawingRepoItem, checked: boolean) => void;
}

const ManageDrawingsPanel = ({ drawingLists, selectedRepo, selectedDrawings, onDrawingSelectionChanged }: ManageDrawingsPanelProps) => {
	const items: DrawingRepoItem[] = drawingLists?.[selectedRepo?.name]?.items ?? emptyArray;

	return <div className={classNames(style.wideDialogContent, style.overflowingDialogContent, style.drawingListContainer)}>
		<div>
			{
				items.map(drawing => {
					return <div className={style.drawingListItem} key={drawing.id}>
						<Checkbox checked={selectedDrawings.includes(drawing)} onChange={change => onDrawingSelectionChanged(drawing, change.target.checked)} />
						<div>{drawing.title}</div>
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
	readonly drawing: DrawingRepoItem;
	readonly onClose: (confirmedRenameTo: string | false) => void;
}

const RenameDrawingDialog = ({ open, drawing, onClose }: RenameDrawingDialogProps) => {
	const [titleToRenameTo, setTitleToRenameTo] = useState<string | null>(null);

	const handleOnClose = (confirmedRenameTo: string | false) => {
		setTitleToRenameTo(null);
		onClose(confirmedRenameTo);
	};

	return <Dialog open={open} maxWidth="xl">
		<DialogTitle>Rename drawing</DialogTitle>
		<DialogContent className={style.wideDialogContent + " " + style.overflowingDialogContent}>
			<TextField label="New title" size="small" sx={{ marginTop: "16px", width: "100%" }} onChange={event => setTitleToRenameTo(event.target.value)} value={titleToRenameTo ?? drawing.title} />
		</DialogContent>
		<DialogActions>
			<Button onClick={() => handleOnClose(false)}>Cancel</Button>
			<Button disabled={titleToRenameTo === null || titleToRenameTo === drawing.title} onClick={() => handleOnClose(titleToRenameTo!)}>Rename</Button>
		</DialogActions>
	</Dialog>;
};

interface DeleteDrawingsDialogProps {
	readonly open: boolean;
	readonly drawings: DrawingRepoItem[];
	readonly onClose: (conformed: boolean) => void;
}

const DeleteDrawingsDialog = ({ open, drawings, onClose }: DeleteDrawingsDialogProps) => {
	return <Dialog open={open} maxWidth="xl">
		<DialogTitle>Delete drawings</DialogTitle>
		<DialogContent className={style.wideDialogContent + " " + style.overflowingDialogContent}>
			<p>Are you sure to delete the following drawings:</p>
			<div>
				{
					drawings.map(drawing => <div key={drawing.id}>{drawing.title}</div>)
				}
			</div>
		</DialogContent>
		<DialogActions>
			<Button onClick={() => onClose(false)}>Cancel</Button>
			<Button color="error" onClick={() => onClose(true)}>Delete</Button>
		</DialogActions>
	</Dialog>;
};

