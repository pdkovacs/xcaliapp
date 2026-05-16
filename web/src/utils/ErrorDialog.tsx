import React from "react";
import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from "@mui/material";

export interface ErrorDialogData {
	readonly title: string;
	readonly message: string;
}

export interface ErrorDialogProps {
	readonly open: boolean;
	readonly onClose: () => void;
	data: ErrorDialogData;
}

export const ErrorDialog = ({ open, onClose, data }: ErrorDialogProps) => {
	return <Dialog
		open={open}
		onClose={onClose}
	>
		<DialogTitle>
			{data?.title ?? "???"}
		</DialogTitle>
		<DialogContent>
			{data?.message ?? "???"}
		</DialogContent>
		<DialogActions>
			<Button onClick={onClose}>Dismiss</Button>
		</DialogActions>
	</Dialog>;
};
